//go:build !unix

package host

import (
	"os"
	"syscall"
)

func detached() *syscall.SysProcAttr { return nil }

func signalGroup(pid int) {
	if pid <= 0 {
		return
	}
	if p, err := os.FindProcess(pid); err == nil {
		_ = p.Kill()
	}
}
