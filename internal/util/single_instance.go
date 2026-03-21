package util

import (
	"errors"
	"os"
)

var ErrLockHeld = errors.New("lock already held")

type FileLock struct {
	file *os.File
	path string
}

func (l *FileLock) Release() error {
	if l == nil || l.file == nil {
		return nil
	}
	return l.file.Close()
}
