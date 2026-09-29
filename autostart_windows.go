//go:build windows

package main

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows/registry"
)

// The logon entry lives in the per-user Run key. It is written through the registry API instead of
// by shelling out to reg.exe, for three reasons:
//
//   - reg.exe is a console program, so launching it from this GUI process flashes a console window
//     (the same defect 012 fixed for desktop notifications);
//   - exec.Command().Run() throws stderr away, so a failure surfaces as a bare "exit status 1";
//   - security suites routinely intercept writes to startup entries. A write that is let through
//     but silently dropped would leave the UI claiming "autostart is on" while nothing ever runs —
//     exactly the state this file exists to prevent, so every write is read back.
const (
	autoStartKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
	autoStartValue   = "Task Manager"
)

// autoStartPath returns the value stored in the Run key. The Run value is a command line, so a path
// containing spaces — this app's folder is "Task Manager" — has to be quoted.
func autoStartPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取程序路径失败：%w", err)
	}
	return `"` + exePath + `"`, nil
}

// readAutoStartValue returns the raw Run value, or "" when the value (or the whole key) is absent.
func readAutoStartValue(valueName string) (string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, autoStartKeyPath, registry.QUERY_VALUE)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	defer key.Close()
	value, _, err := key.GetStringValue(valueName)
	if err != nil {
		if errors.Is(err, registry.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

// writeAutoStartValue creates or removes the named Run entry, then reads the value back. The
// read-back is the point: a write that reports success but does not stick is reported as an error
// rather than being mistaken for a working setting.
func writeAutoStartValue(valueName string, enabled bool) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, autoStartKeyPath,
		registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("打开开机启动项注册表失败：%w", err)
	}
	defer key.Close()

	if !enabled {
		if err := key.DeleteValue(valueName); err != nil && !errors.Is(err, registry.ErrNotExist) {
			return fmt.Errorf("取消开机自启失败：%w", err)
		}
		return nil
	}

	want, err := autoStartPath()
	if err != nil {
		return err
	}
	if err = key.SetStringValue(valueName, want); err != nil {
		return fmt.Errorf("写入开机启动项失败：%w", err)
	}
	got, _, err := key.GetStringValue(valueName)
	if err != nil {
		return fmt.Errorf("开机启动项写入后读回失败：%w", err)
	}
	if got != want {
		return fmt.Errorf("开机启动项未生效（写入 %s，读回 %s），可能被安全软件拦截", want, got)
	}
	return nil
}

// setAutoStart applies the user's choice from the settings dialog.
func setAutoStart(enabled bool) error {
	return writeAutoStartValue(autoStartValue, enabled)
}

// syncAutoStart repairs the logon entry on every start. The stored setting can say "on" while the
// entry is missing (an earlier write was blocked, or a cleanup tool removed it) or while it still
// points at a previous location of the executable (the app keeps its data beside itself, so users
// move it around). Rewriting only when the value differs keeps the normal case write-free.
func syncAutoStart(enabled bool) error {
	return syncAutoStartValue(autoStartValue, enabled)
}

// syncAutoStartValue is syncAutoStart for a named entry, so tests can exercise the repair path
// without touching the logon entry the user actually depends on.
func syncAutoStartValue(valueName string, enabled bool) error {
	if !enabled {
		return nil
	}
	want, err := autoStartPath()
	if err != nil {
		return err
	}
	if got, _ := readAutoStartValue(valueName); got == want {
		return nil
	}
	return writeAutoStartValue(valueName, true)
}
