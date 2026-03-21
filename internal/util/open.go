package util

import (
	"fmt"
	"pornboss/internal/common/logging"
)

// OpenFile opens a file with the system default application.
func OpenFile(path string) error {
	cmd, err := buildOpenCommand(path, false)
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	return nil
}

// RevealFile opens the containing folder and highlights the file when supported.
func RevealFile(path string) error {
	cmd, err := buildOpenCommand(path, true)
	if err != nil {
		return err
	}
	logging.Info("reveal file command: %v", cmd.Args)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("reveal file: %w", err)
	}
	return nil
}
