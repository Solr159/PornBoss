//go:build !windows

package util

import (
	"fmt"
	"os"
	"syscall"
)

func AcquireFileLock(path string) (*FileLock, error) {
	return acquireFileLock(path, os.O_CREATE|os.O_RDWR, false)
}

func AcquireExistingFileLock(path string) (*FileLock, error) {
	return acquireFileLock(path, os.O_RDWR, true)
}

func acquireFileLock(path string, flag int, missingIsOK bool) (*FileLock, error) {
	file, err := os.OpenFile(path, flag, 0o644)
	if err != nil {
		if missingIsOK && os.IsNotExist(err) {
			return nil, ErrLockMissing
		}
		return nil, fmt.Errorf("open lock file: %w", err)
	}

	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if err == syscall.EWOULDBLOCK || err == syscall.EAGAIN {
			return nil, ErrLockHeld
		}
		return nil, fmt.Errorf("lock file: %w", err)
	}

	return &FileLock{file: file, path: path}, nil
}
