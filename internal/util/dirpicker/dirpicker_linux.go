package dirpicker

import (
	"context"
	"errors"
	"os/exec"
)

func pickDirectoryPlatform(ctx context.Context) (string, error) {
	path, err := runPickerCommand(ctx, "zenity", "--file-selection", "--directory", "--title=Select Directory")
	if err == nil || !errors.Is(err, exec.ErrNotFound) {
		return path, err
	}

	path, err = runPickerCommand(ctx, "kdialog", "--getexistingdirectory", ".", "--title", "Select Directory")
	if err == nil || !errors.Is(err, exec.ErrNotFound) {
		return path, err
	}

	return "", errors.New("no supported directory picker found (zenity/kdialog)")
}
