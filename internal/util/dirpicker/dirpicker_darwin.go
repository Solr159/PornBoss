package dirpicker

import (
	"context"
	"errors"
	"os/exec"
)

func pickDirectoryPlatform(ctx context.Context) (string, error) {
	path, err := runPickerCommand(ctx, "osascript", "-e", "POSIX path of (choose folder)")
	if err == nil || !errors.Is(err, exec.ErrNotFound) {
		return path, err
	}
	return "", errors.New("osascript not available for directory picker")
}
