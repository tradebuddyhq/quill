//go:build windows

package main

import "os/exec"

func setProcGroup(cmd *exec.Cmd) {
	// No process group support on Windows
}

func killProcGroup(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	cmd.Process.Kill()
}
