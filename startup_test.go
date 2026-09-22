package main

import (
	"testing"

	"github.com/wailsapp/wails/v2/pkg/options"
)

// The window has to come up already maximised. Wails only reaches that state without a visible
// resize when WindowStartState says so: it creates the window hidden and shows it with
// SW_MAXIMIZE in one step. Maximising from OnDomReady happens after the window is already
// visible, which is exactly the grow-into-place animation the user reported.
func TestAppOptionsStartMaximised(t *testing.T) {
	app := newTestApp(t)
	config := appOptions(app)

	if config.WindowStartState != options.Maximised {
		t.Fatalf("WindowStartState = %v, want options.Maximised", config.WindowStartState)
	}
	if config.OnDomReady != nil {
		t.Fatal("OnDomReady must stay nil: maximising after the window is visible brings back the resize animation")
	}
	// The configured size is still the one the window restores to when un-maximised.
	if config.Width != 1280 || config.Height != 800 {
		t.Fatalf("restored size = %dx%d, want 1280x800", config.Width, config.Height)
	}
}
