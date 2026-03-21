//go:build !windows

package util

import (
	"net/url"
	"os/exec"
	"path/filepath"
	"runtime"
)

func buildOpenCommand(path string, reveal bool) (*exec.Cmd, error) {
	switch runtime.GOOS {
	case "darwin":
		if reveal {
			return exec.Command("open", "-R", path), nil
		}
		return exec.Command("open", path), nil
	default:
		if reveal {
			if cmd, ok := buildLinuxRevealCommand(path); ok {
				return cmd, nil
			}
			return exec.Command("xdg-open", filepath.Dir(path)), nil
		}
		return exec.Command("xdg-open", path), nil
	}
}

func buildLinuxRevealCommand(path string) (*exec.Cmd, bool) {
	if _, err := exec.LookPath("dbus-send"); err != nil {
		return nil, false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	fileURL := (&url.URL{Scheme: "file", Path: absPath}).String()
	return exec.Command(
		"dbus-send",
		"--session",
		"--dest=org.freedesktop.FileManager1",
		"--type=method_call",
		"/org/freedesktop/FileManager1",
		"org.freedesktop.FileManager1.ShowItems",
		"array:string:"+fileURL,
		"string:",
	), true
}
