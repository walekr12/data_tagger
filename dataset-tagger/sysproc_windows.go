//go:build windows
// +build windows

package main

import "syscall"

// getSysProcAttr returns Windows-specific attributes to hide command window
func getSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
