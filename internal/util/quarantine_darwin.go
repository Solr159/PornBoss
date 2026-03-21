//go:build darwin

package util

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

const quarantineAttr = "com.apple.quarantine"

// ClearQuarantineRecursive removes the quarantine attribute from a path tree.
func ClearQuarantineRecursive(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return removeQuarantine(path)
	}
	return filepath.WalkDir(path, func(p string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := removeQuarantine(p); err != nil && !isNoAttr(err) {
			return err
		}
		return nil
	})
}

func removeQuarantine(path string) error {
	if err := unix.Removexattr(path, quarantineAttr); err != nil {
		if isNoAttr(err) {
			return nil
		}
		return err
	}
	return nil
}

func isNoAttr(err error) bool {
	return errors.Is(err, unix.ENOATTR)
}
