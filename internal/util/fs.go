package util

import (
	"errors"
	"os"
	"syscall"
)

// IsPathUnavailable reports whether the error indicates a missing or unavailable device/path.
func IsPathUnavailable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrNotExist) {
		return true
	}
	return errors.Is(err, syscall.ENODEV) || errors.Is(err, syscall.ENXIO) || errors.Is(err, syscall.ESTALE)
}
