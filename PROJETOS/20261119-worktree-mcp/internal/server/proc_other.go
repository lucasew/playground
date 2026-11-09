//go:build !unix

package server

import "syscall"

func detachedAttr() *syscall.SysProcAttr { return nil }

func processAlive(pid int) bool { return pid > 0 }
