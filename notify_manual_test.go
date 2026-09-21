package main

import (
	"os"
	"testing"
)

// TestManualToast is skipped unless TM_TOAST_TEST is set; it exists to verify the real
// desktop notification path on a machine with a desktop session.
func TestManualToast(t *testing.T) {
	if os.Getenv("TM_TOAST_TEST") == "" {
		t.Skip("set TM_TOAST_TEST=1 to send a real notification")
	}
	if err := showWindowsToast("Agent 已恢复可用", "Codex-GPT 的额度重置时间已到，现已恢复可用。"); err != nil {
		t.Fatalf("showWindowsToast failed: %v", err)
	}
}
