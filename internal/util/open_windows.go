//go:build windows

package util

import (
	"os/exec"
	"syscall"
)

func buildOpenCommand(path string, reveal bool) (*exec.Cmd, error) {
	if reveal {
		selectArg := "/select,\"" + path + "\""
		cmd := exec.Command("explorer")
		cmd.SysProcAttr = &syscall.SysProcAttr{CmdLine: "explorer " + selectArg}
		return cmd, nil
	}
	return exec.Command("cmd", "/c", "start", "", path), nil
}
