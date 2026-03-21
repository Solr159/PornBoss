package dirpicker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func runPickerCommand(ctx context.Context, command string, args ...string) (string, error) {
	if _, err := exec.LookPath(command); err != nil {
		return "", err
	}
	cmd := exec.CommandContext(ctx, command, args...)
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	trimmed := strings.TrimSpace(stdout.String())
	if err != nil {
		if trimmed == "" {
			return "", ErrDirPickerCanceled
		}
		return "", fmt.Errorf("%s failed: %w: %s", command, err, strings.TrimSpace(stderr.String()))
	}
	if trimmed == "" {
		return "", ErrDirPickerCanceled
	}
	return trimmed, nil
}
