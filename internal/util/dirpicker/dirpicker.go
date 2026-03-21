package dirpicker

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
)

var ErrDirPickerCanceled = errors.New("directory picker canceled")

// PickDirectory opens a system directory picker and returns the selected path.
func PickDirectory(ctx context.Context) (string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	path, err := pickDirectoryPlatform(ctx)
	if err != nil {
		return "", err
	}
	path = strings.TrimSpace(path)
	if path == "" {
		return "", ErrDirPickerCanceled
	}
	return filepath.Clean(path), nil
}
