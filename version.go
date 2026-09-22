package main

import (
	_ "embed"
	"encoding/json"
	"strings"
	"sync"
)

// wailsJSON is a verbatim copy of the build config, embedded at compile time so the
// running binary never needs wails.json to sit next to it on disk.
//
//go:embed wails.json
var wailsJSON []byte

// productVersion has exactly one source of truth: wails.json's info.productVersion.
// Both the exe's file properties (rendered through build/windows/info.json) and the
// About dialog read that same field, so bumping the version means editing one line.
var productVersion = sync.OnceValue(func() string {
	var cfg struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err := json.Unmarshal(wailsJSON, &cfg); err != nil {
		return ""
	}
	return strings.TrimSpace(cfg.Info.ProductVersion)
})

// GetAppVersion returns the product version shown in the About dialog. It mirrors
// wails.json's info.productVersion (expected format x.x.x), so the dialog and the
// exe's "Details" tab can never disagree.
func (a *App) GetAppVersion() string {
	return productVersion()
}
