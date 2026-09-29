//go:build windows

package main

import (
	"os"
	"testing"

	"golang.org/x/sys/windows/registry"
)

// probeValue keeps these tests away from the real "Task Manager" logon entry: a unit test must never
// decide whether the user's app starts with Windows.
const probeValue = "TaskManagerAutostartProbe"

func deleteProbe(t *testing.T) {
	t.Helper()
	key, _, err := registry.CreateKey(registry.CURRENT_USER, autoStartKeyPath, registry.SET_VALUE)
	if err != nil {
		t.Fatalf("打开 Run 键失败: %v", err)
	}
	defer key.Close()
	_ = key.DeleteValue(probeValue)
}

func TestAutoStartPathIsQuotedExecutable(t *testing.T) {
	got, err := autoStartPath()
	if err != nil {
		t.Fatalf("autoStartPath() 失败: %v", err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() 失败: %v", err)
	}
	if want := `"` + exe + `"`; got != want {
		t.Fatalf("autoStartPath() = %q, 期望 %q", got, want)
	}
}

func TestWriteAutoStartValueRoundTrip(t *testing.T) {
	deleteProbe(t)
	defer deleteProbe(t)

	if err := writeAutoStartValue(probeValue, true); err != nil {
		t.Fatalf("写入开机启动项失败: %v", err)
	}
	got, err := readAutoStartValue(probeValue)
	if err != nil {
		t.Fatalf("读回开机启动项失败: %v", err)
	}
	want, _ := autoStartPath()
	if got != want {
		t.Fatalf("读回 %q, 期望 %q", got, want)
	}

	if err := writeAutoStartValue(probeValue, false); err != nil {
		t.Fatalf("取消开机自启失败: %v", err)
	}
	if got, err = readAutoStartValue(probeValue); err != nil {
		t.Fatalf("删除后读回失败: %v", err)
	}
	if got != "" {
		t.Fatalf("删除后仍读到 %q", got)
	}
}

// 把开关从「开」拨到「关」时，项可能本来就不存在，这必须算成功。
func TestWriteAutoStartValueDisableIsIdempotent(t *testing.T) {
	deleteProbe(t)
	defer deleteProbe(t)

	if err := writeAutoStartValue(probeValue, false); err != nil {
		t.Fatalf("删除不存在的项应成功，实际: %v", err)
	}
	if err := writeAutoStartValue(probeValue, false); err != nil {
		t.Fatalf("重复删除应成功，实际: %v", err)
	}
}

// 读一个不存在的项要返回空串而不是错误，否则调用方无法区分「没设过」和「读失败」。
func TestReadAutoStartValueMissingIsEmpty(t *testing.T) {
	deleteProbe(t)

	got, err := readAutoStartValue(probeValue)
	if err != nil {
		t.Fatalf("读取不存在的项不应报错，实际: %v", err)
	}
	if got != "" {
		t.Fatalf("读取不存在的项应返回空串，实际 %q", got)
	}
}

// syncAutoStartValue 存在的意义就是修正「指向旧位置」的项（用户移动过程序目录）。
func TestSyncAutoStartValueRepairsStalePath(t *testing.T) {
	deleteProbe(t)
	defer deleteProbe(t)

	key, _, err := registry.CreateKey(registry.CURRENT_USER, autoStartKeyPath,
		registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		t.Fatalf("打开 Run 键失败: %v", err)
	}
	stale := `"C:\somewhere else\Task Manager.exe"`
	if err = key.SetStringValue(probeValue, stale); err != nil {
		key.Close()
		t.Fatalf("写入旧路径失败: %v", err)
	}
	key.Close()

	if err = syncAutoStartValue(probeValue, true); err != nil {
		t.Fatalf("syncAutoStartValue 失败: %v", err)
	}
	got, _ := readAutoStartValue(probeValue)
	want, _ := autoStartPath()
	if got != want {
		t.Fatalf("旧路径未被修正: %q, 期望 %q", got, want)
	}
}

func TestSyncAutoStartValueKeepsCorrectPath(t *testing.T) {
	deleteProbe(t)
	defer deleteProbe(t)

	if err := writeAutoStartValue(probeValue, true); err != nil {
		t.Fatalf("写入失败: %v", err)
	}
	if err := syncAutoStartValue(probeValue, true); err != nil {
		t.Fatalf("syncAutoStartValue 失败: %v", err)
	}
	got, _ := readAutoStartValue(probeValue)
	want, _ := autoStartPath()
	if got != want {
		t.Fatalf("值被意外改动: %q, 期望 %q", got, want)
	}
}

// 关掉时不能动注册表：用户可能自己手填过别的启动方式，程序无权代删。
func TestSyncAutoStartValueDisabledWritesNothing(t *testing.T) {
	deleteProbe(t)
	defer deleteProbe(t)

	if err := syncAutoStartValue(probeValue, false); err != nil {
		t.Fatalf("syncAutoStartValue(false) 不应报错，实际: %v", err)
	}
	if got, _ := readAutoStartValue(probeValue); got != "" {
		t.Fatalf("false 时不应写入，实际读到 %q", got)
	}
}
