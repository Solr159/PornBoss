//go:build windows

package util

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func AcquireFileLock(path string) (*FileLock, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	var ol windows.Overlapped
	err = windows.LockFileEx(
		windows.Handle(file.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0,
		1,
		0,
		&ol,
	)
	if err != nil {
		_ = file.Close()
		if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return nil, ErrLockHeld
		}
		return nil, fmt.Errorf("lock file: %w", err)
	}

	return &FileLock{file: file, path: path}, nil
}
