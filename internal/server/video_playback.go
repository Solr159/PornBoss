package server

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
	"pornboss/internal/models"
	"pornboss/internal/util"
)

const (
	hlsManifestMime = "application/vnd.apple.mpegurl"
	hlsSegmentMime  = "video/MP2T"
	hlsWaitTimeout  = 20 * time.Second
	hlsPollInterval = 200 * time.Millisecond
)

var hlsSegmentNamePattern = regexp.MustCompile(`^\d+\.ts$`)

type playbackResponse struct {
	Mode string `json:"mode"`
	Src  string `json:"src"`
	Type string `json:"type"`
}

type hlsManager struct {
	mu       sync.Mutex
	sessions map[string]*hlsSession
}

type hlsSession struct {
	inputPath    string
	dir          string
	manifestPath string
	completePath string
	mu           sync.Mutex
	started      bool
	done         chan struct{}
	err          error
}

var videoHLSManager = &hlsManager{
	sessions: make(map[string]*hlsSession),
}

func getVideoPlayback(c *gin.Context) {
	video, fullPath, ok := loadVideoByID(c)
	if !ok {
		return
	}
	if err := ensureVideoFileExists(c, fullPath); err != nil {
		return
	}

	meta, err := util.ProbeVideo(fullPath)
	if err != nil {
		logging.Error("probe playback metadata error: %v", err)
	}

	mode, mimeType := util.DetermineBrowserPlayback(fullPath, meta)
	if mode == util.BrowserPlaybackModeHLS {
		if _, err := util.ResolveFFmpegPath(); err != nil {
			logging.Error("resolve ffmpeg for hls playback error: %v", err)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ffmpeg unavailable for browser playback"})
			return
		}
	}

	src := fmt.Sprintf("/videos/%d/stream", video.ID)
	if mode == util.BrowserPlaybackModeHLS {
		src = fmt.Sprintf("/videos/%d/stream.m3u8", video.ID)
	}

	c.JSON(http.StatusOK, playbackResponse{
		Mode: mode,
		Src:  src,
		Type: mimeType,
	})
}

func streamVideoManifest(c *gin.Context) {
	video, fullPath, ok := loadVideoByID(c)
	if !ok {
		return
	}
	if err := ensureVideoFileExists(c, fullPath); err != nil {
		return
	}

	if _, err := videoHLSManager.sessionForFile(fullPath); err != nil {
		logging.Error("prepare hls session error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "prepare hls stream failed"})
		return
	}

	manifest, err := buildStaticHLSManifest(video.DurationSec)
	if err != nil {
		logging.Error("build hls manifest error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "build hls manifest failed"})
		return
	}

	c.Header("Content-Type", hlsManifestMime)
	c.Header("Cache-Control", "no-cache")
	c.String(http.StatusOK, manifest)
}

func streamVideoSegment(c *gin.Context) {
	_, fullPath, ok := loadVideoByID(c)
	if !ok {
		return
	}
	if err := ensureVideoFileExists(c, fullPath); err != nil {
		return
	}

	segment := strings.TrimSpace(c.Param("segment"))
	if !hlsSegmentNamePattern.MatchString(segment) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid segment"})
		return
	}

	session, err := videoHLSManager.sessionForFile(fullPath)
	if err != nil {
		logging.Error("prepare hls session error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "prepare hls stream failed"})
		return
	}

	segmentPath := filepath.Join(session.dir, segment)
	if err := session.waitForFile(segmentPath, hlsWaitTimeout); err != nil {
		writeHLSError(c, err)
		return
	}

	c.Header("Content-Type", hlsSegmentMime)
	c.Header("Cache-Control", "public, max-age=300")
	c.File(segmentPath)
}

func loadVideoByID(c *gin.Context) (*models.Video, string, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return nil, "", false
	}

	video, err := dbpkg.GetVideo(c.Request.Context(), id)
	if err != nil {
		logging.Error("get video error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return nil, "", false
	}
	if video == nil {
		c.Status(http.StatusNotFound)
		return nil, "", false
	}

	fullPath, _, err := resolveVideoPath(video.Path, video.DirectoryRef.Path)
	if err != nil {
		logging.Error("resolve video path by id error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return nil, "", false
	}

	return video, fullPath, true
}

func (m *hlsManager) sessionForFile(fullPath string) (*hlsSession, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	key := buildHLSCacheKey(fullPath, info)
	baseDir := filepath.Join(os.TempDir(), "pornboss-hls", key)

	m.mu.Lock()
	session := m.sessions[key]
	if session == nil {
		session = &hlsSession{
			inputPath:    fullPath,
			dir:          baseDir,
			manifestPath: filepath.Join(baseDir, "manifest.m3u8"),
			completePath: filepath.Join(baseDir, ".complete"),
		}
		m.sessions[key] = session
	}
	m.mu.Unlock()

	if err := session.ensureStarted(); err != nil {
		return nil, err
	}
	return session, nil
}

func buildHLSCacheKey(fullPath string, info os.FileInfo) string {
	sum := sha1.Sum([]byte(fullPath + "|" + strconv.FormatInt(info.Size(), 10) + "|" + strconv.FormatInt(info.ModTime().UTC().UnixNano(), 10)))
	return hex.EncodeToString(sum[:])
}

func (s *hlsSession) ensureStarted() error {
	if fileExists(s.completePath) && fileExists(s.manifestPath) {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if fileExists(s.completePath) && fileExists(s.manifestPath) {
		return nil
	}
	if s.started {
		return nil
	}

	if err := os.RemoveAll(s.dir); err != nil {
		return fmt.Errorf("clear hls dir: %w", err)
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("create hls dir: %w", err)
	}

	s.done = make(chan struct{})
	s.err = nil
	s.started = true
	go s.run()
	return nil
}

func (s *hlsSession) run() {
	ffmpegPath, err := util.ResolveFFmpegPath()
	if err != nil {
		s.finish(err)
		return
	}

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-nostdin",
		"-i", s.inputPath,
		"-map", "0:v:0",
		"-map", "0:a?",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "23",
		"-force_key_frames", "expr:gte(t,n_forced*4)",
		"-c:a", "aac",
		"-ac", "2",
		"-sn",
		"-f", "hls",
		"-start_number", "0",
		"-hls_time", "4",
		"-hls_list_size", "0",
		"-hls_playlist_type", "vod",
		"-hls_base_url", "stream.m3u8/",
		"-hls_flags", "independent_segments+temp_file",
		"-hls_segment_filename", filepath.Join(s.dir, "%06d.ts"),
		s.manifestPath,
	}

	cmd := exec.Command(ffmpegPath, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if runErr := cmd.Run(); runErr != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			err = fmt.Errorf("ffmpeg hls failed: %w: %s", runErr, msg)
		} else {
			err = fmt.Errorf("ffmpeg hls failed: %w", runErr)
		}
	} else {
		err = os.WriteFile(s.completePath, []byte(time.Now().UTC().Format(time.RFC3339)), 0o644)
		if err != nil {
			err = fmt.Errorf("write hls complete marker: %w", err)
		}
	}

	s.finish(err)
}

func (s *hlsSession) finish(err error) {
	s.mu.Lock()
	done := s.done
	s.err = err
	s.started = false
	s.mu.Unlock()

	if done != nil {
		close(done)
	}
}

func (s *hlsSession) waitForFile(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if info, err := os.Stat(path); err == nil && info.Size() > 0 {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for hls output: %s", filepath.Base(path))
		}

		done, runErr, started := s.status()
		if !started && runErr != nil {
			if fileExists(path) {
				return nil
			}
			return runErr
		}
		if done != nil {
			select {
			case <-done:
				if info, err := os.Stat(path); err == nil && info.Size() > 0 {
					return nil
				}
				_, runErr, _ = s.status()
				if runErr != nil {
					return runErr
				}
			default:
			}
		}

		time.Sleep(hlsPollInterval)
	}
}

func buildStaticHLSManifest(durationSec int64) (string, error) {
	if durationSec <= 0 {
		return "", errors.New("invalid video duration")
	}

	segmentCount := int(math.Ceil(float64(durationSec) / 4.0))
	if segmentCount < 1 {
		segmentCount = 1
	}

	var buf strings.Builder
	buf.WriteString("#EXTM3U\n")
	buf.WriteString("#EXT-X-VERSION:3\n")
	buf.WriteString("#EXT-X-PLAYLIST-TYPE:VOD\n")
	buf.WriteString("#EXT-X-TARGETDURATION:4\n")
	buf.WriteString("#EXT-X-MEDIA-SEQUENCE:0\n")

	remaining := float64(durationSec)
	for i := 0; i < segmentCount; i++ {
		segmentDuration := 4.0
		if remaining > 0 && remaining < segmentDuration {
			segmentDuration = remaining
		}
		if segmentDuration <= 0 {
			segmentDuration = 4.0
		}

		buf.WriteString(fmt.Sprintf("#EXTINF:%.3f,\n", segmentDuration))
		buf.WriteString(fmt.Sprintf("stream.m3u8/%06d.ts\n", i))
		remaining -= segmentDuration
	}

	buf.WriteString("#EXT-X-ENDLIST\n")
	return buf.String(), nil
}

func (s *hlsSession) status() (chan struct{}, error, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.done, s.err, s.started
}

func writeHLSError(c *gin.Context, err error) {
	logging.Error("serve hls error: %v", err)
	if errors.Is(err, os.ErrNotExist) {
		c.Status(http.StatusNotFound)
		return
	}
	if strings.Contains(err.Error(), "timeout waiting for hls output") {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "hls stream timeout"})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "hls stream failed"})
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
