//go:build !linux && !darwin && !windows

package dirpicker

import (
	"context"
	"errors"
)

func pickDirectoryPlatform(ctx context.Context) (string, error) {
	_ = ctx
	return "", errors.New("directory picker not supported on this platform")
}
