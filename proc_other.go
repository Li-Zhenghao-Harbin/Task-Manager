//go:build !windows

package main

import "syscall"

// hiddenProcessAttr is a no-op on platforms that have no console allocation to suppress.
// It exists so notify.go stays free of build tags; the app itself only ships for Windows.
func hiddenProcessAttr() *syscall.SysProcAttr { return nil }
