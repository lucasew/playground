//go:build unix

package server

import (
	"os"
	"os/exec"
	"syscall"
)

// detachedAttr puts the child in its own process group.
func detachedAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

// processAlive reports whether pid names a live process. Signal 0 checks
// for existence without delivering anything.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

var _ = exec.Command
