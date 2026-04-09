//go:build integration

package config

import (
	"context"
	"os"
	"testing"

	"github.com/go-rod/rod"
)

func testBrowserAndPage(t *testing.T) *rod.Page {
	t.Helper()
	dir, err := os.MkdirTemp("", "chrome-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp() error = %v", err)
	}
	cfg := Config{ChromeProfileDir: dir}

	browser, cleanup, err := NewBrowser(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewBrowser() error = %v", err)
	}
	t.Cleanup(cleanup)

	page, err := NewStealthPage(browser)
	if err != nil {
		t.Fatalf("NewStealthPage() error = %v", err)
	}
	return page
}

func TestStealth_NavigatorWebdriverFalsy(t *testing.T) {
	page := testBrowserAndPage(t)
	page.MustNavigate("about:blank").MustWaitStable()

	val := page.MustEval("() => navigator.webdriver")
	if val.Bool() {
		t.Error("navigator.webdriver should be falsy, got true")
	}
}

func TestStealth_PluginsNonEmpty(t *testing.T) {
	page := testBrowserAndPage(t)
	page.MustNavigate("about:blank").MustWaitStable()

	count := page.MustEval("() => navigator.plugins.length").Int()
	if count == 0 {
		t.Error("navigator.plugins.length should be > 0")
	}
}

func TestStealth_OuterDimensions(t *testing.T) {
	page := testBrowserAndPage(t)
	page.MustNavigate("about:blank").MustWaitStable()

	outer := page.MustEval("() => window.outerWidth").Int()
	inner := page.MustEval("() => window.innerWidth").Int()
	if outer < inner {
		t.Errorf("outerWidth (%d) < innerWidth (%d)", outer, inner)
	}
}

// Custom patches are registered via EvalOnNewDocument after the initial
// about:blank load, so we must navigate to a new document to trigger them.

func TestCustomPatches_ScreenWidth(t *testing.T) {
	page := testBrowserAndPage(t)
	page.MustNavigate("data:text/html,<h1>test</h1>").MustWaitStable()

	width := page.MustEval("() => screen.width").Int()
	if width != 1920 {
		t.Errorf("screen.width = %d, want 1920", width)
	}
}

func TestCustomPatches_ColorDepth(t *testing.T) {
	page := testBrowserAndPage(t)
	page.MustNavigate("data:text/html,<h1>test</h1>").MustWaitStable()

	depth := page.MustEval("() => screen.colorDepth").Int()
	if depth != 24 {
		t.Errorf("screen.colorDepth = %d, want 24", depth)
	}
}
