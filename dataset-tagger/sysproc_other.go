//go:build !windows
// +build !windows

package main

import "syscall"

// getSysProcAttr returns nil for non-Windows platforms
func getSysProcAttr() *syscall.SysProcAttr {
	return nil
}
