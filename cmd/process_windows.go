//go:build windows

package cmd

import (
	"os"
	"os/exec"
)

func detachProcess(_ *exec.Cmd) {}

func terminateProcess(pid int) error {
	return killProcess(pid)
}

func killProcess(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	defer p.Release()
	return p.Kill()
}
