package service

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"pornboss/internal/common"
	"pornboss/internal/common/logging"
	"pornboss/internal/db"
	"pornboss/internal/models"
	"pornboss/internal/util"
)

type FileEntry struct {
	DirectoryID   int64
	DirectoryPath string
	RelativePath  string
	AbsolutePath  string
	Filename      string
	Size          int64
	ModifiedAt    time.Time
	Fingerprint   string
	PathKey       string
	DurationSec   int64
}

// Summary captures the results of a directory synchronisation run.
type Summary struct {
	FilesSeen   int
	Inserted    int
	Updated     int
	Removed     int
	Duration    time.Duration
	Directories int
}

func makePathKey(directoryID int64, relativePath string) string {
	return strconv.FormatInt(directoryID, 10) + ":" + relativePath
}

// SyncAllDirectories scans multiple directories in a single pass so cross-directory moves keep tags/metadata.
func SyncAllDirectories(ctx context.Context, directories []models.Directory) (*Summary, error) {
	if common.DB == nil {
		return nil, errors.New("nil database")
	}
	if len(directories) == 0 {
		return &Summary{}, nil
	}

	locked := make([]int64, 0, len(directories))
	unlockAll := func() {
		for _, id := range locked {
			unlockDirectory(id)
		}
	}
	for _, dir := range directories {
		if dir.ID == 0 || dir.IsDelete {
			continue
		}
		if !tryLockDirectory(dir.ID) {
			unlockAll()
			return nil, ErrDirectoryScanInProgress
		}
		locked = append(locked, dir.ID)
	}
	defer unlockAll()

	start := time.Now()
	summary := &Summary{}

	existing, err := db.AllVideos(ctx)
	if err != nil {
		return nil, err
	}
	existingByPath := make(map[string]*models.Video, len(existing))
	existingByFingerprint := make(map[string]*models.Video, len(existing))
	processedIDs := make(map[int64]struct{}, len(existing))
	for i := range existing {
		video := &existing[i]
		if video.Fingerprint != "" {
			existingByFingerprint[video.Fingerprint] = video
		}
		existingByPath[makePathKey(video.DirectoryID, video.Path)] = video
	}

	scanned := make(map[int64]struct{}, len(directories))
	for _, dir := range directories {
		if dir.ID == 0 || dir.IsDelete {
			continue
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		info, statErr := os.Stat(dir.Path)
		if statErr != nil {
			if util.IsPathUnavailable(statErr) {
				if err := db.SetDirectoryMissingAndHideVideos(ctx, dir.ID, true); err != nil {
					logging.Error("mark directory missing failed id=%d path=%s err=%v", dir.ID, dir.Path, err)
				}
				continue
			}
			return nil, fmt.Errorf("stat directory %s: %w", dir.Path, statErr)
		}
		if !info.IsDir() {
			if err := db.SetDirectoryMissingAndHideVideos(ctx, dir.ID, true); err != nil {
				logging.Error("mark directory missing failed id=%d path=%s err=%v", dir.ID, dir.Path, err)
			}
			continue
		}
		if dir.Missing {
			if err := db.SetDirectoryMissingAndHideVideos(ctx, dir.ID, false); err != nil {
				logging.Error("clear directory missing failed id=%d path=%s err=%v", dir.ID, dir.Path, err)
			}
		}

		scanned[dir.ID] = struct{}{}
		if err := walkDirectoryAndSync(ctx, dir, existingByPath, existingByFingerprint, processedIDs, summary); err != nil {
			return nil, err
		}
	}

	if err := hideUnprocessedVideos(ctx, processedIDs, summary); err != nil {
		return nil, err
	}

	summary.Directories = len(scanned)
	summary.Duration = time.Since(start)
	return summary, nil
}

// walkDirectoryAndSync 扫描单个目录下的文件，实时 upsert 并记录统计/缩略图任务。
func walkDirectoryAndSync(ctx context.Context, directory models.Directory, existingByPath map[string]*models.Video, existingByFingerprint map[string]*models.Video, processedIDs map[int64]struct{}, summary *Summary) error {
	// 边遍历文件边做指纹计算和 DB 更新，避免一次性构建全量快照
	normalizedRoot := filepath.Clean(directory.Path)
	return filepath.WalkDir(normalizedRoot, func(candidatePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if errors.Is(walkErr, fs.ErrPermission) {
				return nil
			}
			return walkErr
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}
		if !util.IsVideo(candidatePath) {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			if errors.Is(err, fs.ErrPermission) {
				return nil
			}
			return err
		}

		// 计算相对路径，确保只处理目录内的文件（防止符号链接等越界）
		normalizedPath := filepath.Clean(candidatePath)
		relativePath, err := filepath.Rel(normalizedRoot, normalizedPath)
		if err != nil {
			return err
		}
		if strings.HasPrefix(relativePath, "..") {
			return nil
		}

		relativePath = filepath.ToSlash(cleanRelativePath(relativePath))
		modTime := info.ModTime().UTC()
		summary.FilesSeen++

		// If file unchanged (size + mtime), skip probe and DB touches but mark as seen.
		pathKey := makePathKey(directory.ID, relativePath)
		if existingVideo, ok := existingByPath[pathKey]; ok {
			if existingVideo.Size == info.Size() && existingVideo.ModifiedAt.Equal(modTime) && existingVideo.Filename == info.Name() {
				processedIDs[existingVideo.ID] = struct{}{}
				if existingVideo.Fingerprint != "" {
					delete(existingByFingerprint, existingVideo.Fingerprint)
				}
				if existingVideo.Hidden {
					existingVideo.Hidden = false
					if err := db.SaveVideo(ctx, existingVideo); err != nil {
						logging.Error("unhide video failed, skip: path=%s err=%v", normalizedPath, err)
						return nil
					}
					summary.Updated++
				}
				return nil
			}
		}

		logging.Info("scan file: root=%s path=%s size=%d", normalizedRoot, relativePath, info.Size())

		meta, err := util.ProbeVideo(normalizedPath)
		if err != nil {
			logging.Error("probe video metadata error: %v", err)
			return nil
		}
		fingerprint := meta.FingerprintV2(info.Size())
		durationSec := int64(math.Round(meta.DurationSeconds))

		fileEntry := &FileEntry{
			DirectoryID:   directory.ID,
			DirectoryPath: normalizedRoot,
			RelativePath:  relativePath,
			AbsolutePath:  normalizedPath,
			Filename:      info.Name(),
			Size:          info.Size(),
			ModifiedAt:    modTime,
			Fingerprint:   fingerprint,
			DurationSec:   durationSec,
		}

		return upsertVideo(ctx, fileEntry, existingByFingerprint, processedIDs, summary)
	})
}

func upsertVideo(ctx context.Context, entry *FileEntry, existingByFingerprint map[string]*models.Video, processedIDs map[int64]struct{}, summary *Summary) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// 已存在：检测元信息/文件是否变化，必要时更新
	var video *models.Video
	if entry.Fingerprint != "" {
		if existingVideo, ok := existingByFingerprint[entry.Fingerprint]; ok {
			video = existingVideo
		}
	}

	if video != nil {
		// 从指纹缓存中移除，避免重复匹配同一条记录。
		if entry.Fingerprint != "" {
			delete(existingByFingerprint, entry.Fingerprint)
		}

		if video.Path != entry.RelativePath || video.DirectoryID != entry.DirectoryID || !video.ModifiedAt.Equal(entry.ModifiedAt) {
			video.Path = entry.RelativePath
			video.DirectoryID = entry.DirectoryID
			video.Filename = entry.Filename
			video.Size = entry.Size
			video.ModifiedAt = entry.ModifiedAt
			video.Fingerprint = entry.Fingerprint
			video.DurationSec = entry.DurationSec
			video.Hidden = false
			if video.Filename != entry.Filename {
				video.JavID = nil
				video.Jav = nil
			}

			if err := db.SaveVideo(ctx, video); err != nil {
				logging.Error("save video failed, skip: path=%s err=%v", entry.AbsolutePath, err)
				processedIDs[video.ID] = struct{}{}
				return nil
			}
			summary.Updated++
		}

		processedIDs[video.ID] = struct{}{}
		return nil
	}

	video = &models.Video{
		Path:        entry.RelativePath,
		DirectoryID: entry.DirectoryID,
		Filename:    entry.Filename,
		Size:        entry.Size,
		ModifiedAt:  entry.ModifiedAt,
		Fingerprint: entry.Fingerprint,
		DurationSec: entry.DurationSec,
		Hidden:      false,
	}

	if err := db.CreateVideo(ctx, video); err != nil {
		logging.Error("create video failed, skip: path=%s err=%v", entry.AbsolutePath, err)
		return nil
	}
	summary.Inserted++
	processedIDs[video.ID] = struct{}{}
	common.ScreenshotManager.EnqueueForVideo(video)
	return nil
}

// hideUnprocessedVideos 将本轮未扫描到的旧数据标记为隐藏（软删除），以便日后恢复。
func hideUnprocessedVideos(ctx context.Context, processedIDs map[int64]struct{}, summary *Summary) error {
	unhiddenIDs, err := db.ListUnhiddenVideoIDs(ctx)
	if err != nil {
		return err
	}
	if len(unhiddenIDs) == 0 {
		return nil
	}

	staleIDs := make([]int64, 0, len(unhiddenIDs))
	for _, id := range unhiddenIDs {
		if _, ok := processedIDs[id]; ok {
			continue
		}
		staleIDs = append(staleIDs, id)
	}

	if len(staleIDs) == 0 {
		return nil
	}

	logging.Info("hiding stale videos: count=%d", len(staleIDs))
	if err := db.HideVideosByIDs(ctx, staleIDs); err != nil {
		return err
	}
	summary.Removed += len(staleIDs)
	return nil
}

func cleanRelativePath(p string) string {
	if p == "" {
		return ""
	}
	cleaned := filepath.Clean(p)
	if cleaned == "." {
		return ""
	}
	return filepath.ToSlash(cleaned)
}
