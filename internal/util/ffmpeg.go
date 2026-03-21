package util

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	ffmpegOnce sync.Once
	ffmpegPath string
	ffmpegErr  error
)

// ResolveFFmpegPath returns the path to the ffmpeg binary by checking:
//  1. FFMPEG_PATH env var
//  2. internal/bin/ffmpeg(.exe) relative to current working directory
//  3. internal/bin/ffmpeg(.exe) relative to the running executable
//  4. ffmpeg(.exe) in PATH
//
// The result is cached after the first resolution attempt.
func ResolveFFmpegPath() (string, error) {
	ffmpegOnce.Do(func() {
		ffmpegPath, ffmpegErr = findFFmpegPath()
	})
	return ffmpegPath, ffmpegErr
}

func findFFmpegPath() (string, error) {
	var candidates []string

	if env := strings.TrimSpace(os.Getenv("FFMPEG_PATH")); env != "" {
		candidates = append(candidates, env)
	}

	binName := "ffmpeg"
	if runtime.GOOS == "windows" {
		binName = "ffmpeg.exe"
	}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "internal", "bin", binName))
	}

	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		candidates = append(candidates, filepath.Join(execDir, "internal", "bin", binName))
	}

	candidates = append(candidates, binName)
	if binName != "ffmpeg" {
		candidates = append(candidates, "ffmpeg")
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if resolved, err := exec.LookPath(candidate); err == nil {
			return resolved, nil
		}
	}

	return "", errors.New("ffmpeg not found; set FFMPEG_PATH or place binary at internal/bin/ffmpeg")
}
