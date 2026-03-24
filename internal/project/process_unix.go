//go:build !windows

package project

import "syscall"

// IsProcessAlive checks whether a process with the given PID is still running.
func IsProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}
