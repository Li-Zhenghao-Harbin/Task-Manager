//go:build windows

package main

import "syscall"

// createNoWindow stops Windows from allocating a console for the child process at all.
// Without it, launching powershell.exe from a GUI process makes a console window flash on
// screen even though -WindowStyle Hidden hides the PowerShell host window itself.
const createNoWindow = 0x08000000

// hiddenProcessAttr launches the child process without any window: HideWindow keeps its main
// window hidden and CREATE_NO_WINDOW skips the console allocation that used to flash a cmd
// window every time a desktop notification was shown.
func hiddenProcessAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
}
