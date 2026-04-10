//go:build integration

package config

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-rod/rod"
)

func TestSetupExtensionHijack_StartsWithoutError(t *testing.T) {
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

	router, err := SetupExtensionHijack(browser)
	if err != nil {
		t.Fatalf("SetupExtensionHijack() error = %v", err)
	}
	t.Cleanup(func() { _ = router.Stop() })
}

// TestPageHijack_MockResponse verifies the hijack mechanism works by
// intercepting an HTTP request with a page-level hijack. This validates
// the same rod HijackRouter pattern that SetupExtensionHijack uses.
// chrome-extension:// interception cannot be integration-tested in headless
// mode because Chrome handles extension URLs internally without routing
// through the Fetch CDP domain.
func TestPageHijack_MockResponse(t *testing.T) {
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

	// Serve a page that performs a fetch we'll intercept.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body>test</body></html>`)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)

	router := page.HijackRequests()
	t.Cleanup(func() { _ = router.Stop() })

	router.MustAdd(srv.URL+"/api/data", func(ctx *rod.Hijack) {
		ctx.Response.SetHeader("Content-Type", "application/json")
		ctx.Response.SetBody(`{"intercepted":true}`)
	})
	go router.Run()

	page.MustNavigate(srv.URL).MustWaitStable()

	result := page.MustEval(fmt.Sprintf(`() => {
		return fetch("%s/api/data").then(r => r.json())
	}`, srv.URL))

	if !result.Get("intercepted").Bool() {
		t.Error("expected hijacked response with intercepted=true")
	}
}
