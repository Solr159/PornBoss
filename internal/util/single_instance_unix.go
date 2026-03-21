//go:build !windows

package util

import (
	"fmt"
	"os"
	"syscall"
)

func AcquireFileLock(path string) (*FileLock, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
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
