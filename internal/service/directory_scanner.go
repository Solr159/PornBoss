package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"pornboss/internal/common"
	"pornboss/internal/common/logging"
	"pornboss/internal/db"
)

var (
	ErrDirectoryScanInProgress = errors.New("directory scan in progress")
	dirScanMu                  sync.Mutex
	dirScanActive              = map[int64]struct{}{}
)

func tryLockDirectory(id int64) bool {
	if id == 0 {
		return false
	}
	dirScanMu.Lock()
	defer dirScanMu.Unlock()
	if _, ok := dirScanActive[id]; ok {
		return false
	}
	dirScanActive[id] = struct{}{}
	return true
}

func unlockDirectory(id int64) {
	if id == 0 {
		return
	}
	dirScanMu.Lock()
	delete(dirScanActive, id)
	dirScanMu.Unlock()
}

// ScanDirectories scans all directories sequentially.
func ScanDirectories(ctx context.Context) error {
	if common.DB == nil {
		return errors.New("nil db")
	}
	dirs, err := db.ListActiveDirectories(ctx)
	if err != nil {
		return err
	}
	if _, err := SyncAllDirectories(ctx, dirs); err != nil {
		if errors.Is(err, ErrDirectoryScanInProgress) {
			return nil
		}
		return err
	}
	return nil
}

// StartDirectoryScanner periodically scans directories.
// The task runs immediately once, then every interval until ctx is done.
func StartDirectoryScanner(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			if err := ScanDirectories(ctx); err != nil {
				logging.Error("periodic directory scan failed: %v", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}
