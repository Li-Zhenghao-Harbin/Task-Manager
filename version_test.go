package main

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"
)

// The About dialog must never disagree with the version baked into the exe's file
// properties. Both come from wails.json's info.productVersion, so this pins the
// embedded copy to the file on disk.
func TestGetAppVersionMatchesWailsJSON(t *testing.T) {
	raw, err := os.ReadFile("wails.json")
	if err != nil {
		t.Fatalf("读取 wails.json 失败: %v", err)
	}
	var cfg struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err = json.Unmarshal(raw, &cfg); err != nil {
		t.Fatalf("解析 wails.json 失败: %v", err)
	}

	got := NewApp().GetAppVersion()
	if got == "" {
		t.Fatal("GetAppVersion() 返回空值：wails.json 的 info.productVersion 缺失或 embed 失效")
	}
	if got != cfg.Info.ProductVersion {
		t.Fatalf("GetAppVersion() = %q, wails.json 里是 %q", got, cfg.Info.ProductVersion)
	}
}

// x.x.x keeps the value readable in the About dialog and still satisfies the
// version format wails/the NSIS installer require.
func TestGetAppVersionUsesThreeSegments(t *testing.T) {
	got := NewApp().GetAppVersion()
	if !regexp.MustCompile(`^\d+\.\d+\.\d+$`).MatchString(got) {
		t.Fatalf("GetAppVersion() = %q, 期望 x.x.x 格式", got)
	}
}
