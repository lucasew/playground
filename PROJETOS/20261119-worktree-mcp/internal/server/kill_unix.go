//go:build unix

package server

import "syscall"

// signalGroup kills the process group of an exec, so the wrapper and every
// command it started die together.
func signalGroup(pid int) {
	if pid <= 0 {
		return
	}
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}
