//go:build windows

package main

import "testing"

// A toast is sent by running powershell.exe, which is a console program: launched from a GUI
// process it asks Windows for a console, and that allocation is the cmd window that used to
// flash on screen. The child must be started with CREATE_NO_WINDOW (and HideWindow) so no
// console exists at all — -WindowStyle Hidden alone does not prevent the flash.
func TestToastProcessRunsWithoutAConsoleWindow(t *testing.T) {
	attr := hiddenProcessAttr()
	if attr == nil {
		t.Fatal("hiddenProcessAttr() returned nil on Windows")
	}
	if !attr.HideWindow {
		t.Fatal("HideWindow must be set so the child process' window stays hidden")
	}
	if attr.CreationFlags&createNoWindow == 0 {
		t.Fatalf("CreationFlags = %#x, want the CREATE_NO_WINDOW bit (%#x) set", attr.CreationFlags, createNoWindow)
	}
}
