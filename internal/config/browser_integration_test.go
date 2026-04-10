//go:build integration

package config

import (
	"context"
	"os"
	"testing"
)

func TestNewBrowser_LaunchesWithoutError(t *testing.T) {
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

	if browser == nil {
		t.Fatal("NewBrowser() returned nil browser")
	}
}
