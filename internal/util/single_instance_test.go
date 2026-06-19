package util

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAcquireExistingFileLockDoesNotCreateMissingLock(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "pornboss.lock")

	lock, err := AcquireExistingFileLock(lockPath)
	if lock != nil {
		t.Fatalf("expected no lock for missing file")
	}
	if !errors.Is(err, ErrLockMissing) {
		t.Fatalf("expected ErrLockMissing, got %v", err)
	}
	if _, statErr := os.Stat(lockPath); !os.IsNotExist(statErr) {
		t.Fatalf("expected missing lock file to stay missing, stat err: %v", statErr)
	}
}
