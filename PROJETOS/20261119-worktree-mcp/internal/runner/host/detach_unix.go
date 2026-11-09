//go:build unix

package host

import "syscall"

// detached puts the child in its own process group, so a kill to one exec
// never reaches the process that started it.
func detached() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

// signalGroup kills the process group led by pid.
func signalGroup(pid int) {
	if pid <= 0 {
		return
	}
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}
