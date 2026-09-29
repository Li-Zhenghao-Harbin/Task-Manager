//go:build !windows

package main

// The app ships for Windows only. These stubs exist so the package still builds and vets on other
// platforms; they mirror the signatures in autostart_windows.go.

func readAutoStartValue(string) (string, error) { return "", nil }

func setAutoStart(bool) error { return nil }

func syncAutoStart(bool) error { return nil }

func syncAutoStartValue(string, bool) error { return nil }
