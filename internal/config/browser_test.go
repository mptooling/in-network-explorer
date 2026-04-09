package config

import (
	"testing"

	"github.com/go-rod/rod/lib/launcher/flags"
)

func TestBrowserConfig_ProxySet(t *testing.T) {
	cfg := Config{
		ChromeProfileDir: "/tmp/test-profile",
		ProxyAddr:        "proxy.example.com:8080",
	}
	l := newLauncher(cfg)

	if !l.Has(flags.ProxyServer) {
		t.Fatal("expected proxy-server flag to be set")
	}
	if got := l.Get(flags.ProxyServer); got != "proxy.example.com:8080" {
		t.Errorf("proxy-server = %q, want %q", got, "proxy.example.com:8080")
	}
}

func TestBrowserConfig_NoProxyWhenEmpty(t *testing.T) {
	cfg := Config{ChromeProfileDir: "/tmp/test-profile"}
	l := newLauncher(cfg)

	if l.Has(flags.ProxyServer) {
		t.Fatal("expected proxy-server flag to be absent when ProxyAddr is empty")
	}
}

func TestBrowserConfig_UserDataDir(t *testing.T) {
	cfg := Config{ChromeProfileDir: "/data/chrome-profile"}
	l := newLauncher(cfg)

	if got := l.Get(flags.UserDataDir); got != "/data/chrome-profile" {
		t.Errorf("user-data-dir = %q, want %q", got, "/data/chrome-profile")
	}
}

func TestBrowserConfig_KeepsUserDataDir(t *testing.T) {
	cfg := Config{ChromeProfileDir: "/tmp/test-profile"}
	l := newLauncher(cfg)

	if !l.Has(flags.KeepUserDataDir) {
		t.Fatal("expected rod-keep-user-data-dir flag to be set")
	}
}

func TestBrowserConfig_NoSandbox(t *testing.T) {
	cfg := Config{ChromeProfileDir: "/tmp/test-profile"}
	l := newLauncher(cfg)

	if !l.Has(flags.NoSandbox) {
		t.Fatal("expected no-sandbox flag to be set")
	}
}

func TestBrowserConfig_DisableWebRTC(t *testing.T) {
	cfg := Config{ChromeProfileDir: "/tmp/test-profile"}
	l := newLauncher(cfg)

	if !l.Has("disable-webrtc") {
		t.Fatal("expected disable-webrtc flag to be set")
	}
}

func TestBrowserConfig_AntiDetectionFlags(t *testing.T) {
	cfg := Config{ChromeProfileDir: "/tmp/test-profile"}
	l := newLauncher(cfg)

	cases := []struct {
		flag flags.Flag
		want string
	}{
		{"disable-blink-features", "AutomationControlled"},
		{"window-size", "1920,1080"},
		{"lang", "en-US"},
		{"use-gl", "angle"},
		{"use-angle", "swiftshader"},
	}

	for _, tc := range cases {
		t.Run(string(tc.flag), func(t *testing.T) {
			if !l.Has(tc.flag) {
				t.Fatalf("flag %q not set", tc.flag)
			}
			if got := l.Get(tc.flag); got != tc.want {
				t.Errorf("%s = %q, want %q", tc.flag, got, tc.want)
			}
		})
	}

	// enable-automation must be removed (rod sets it by default).
	if l.Has("enable-automation") {
		t.Error("expected enable-automation flag to be deleted")
	}
}

func TestBrowserConfig_ChromeBinSet(t *testing.T) {
	cfg := Config{
		ChromeProfileDir: "/tmp/test-profile",
		ChromeBin:        "/usr/bin/chromium",
	}
	l := newLauncher(cfg)

	if got := l.Get(flags.Bin); got != "/usr/bin/chromium" {
		t.Errorf("bin = %q, want %q", got, "/usr/bin/chromium")
	}
}

func TestBrowserConfig_ChromeBinDefault(t *testing.T) {
	cfg := Config{ChromeProfileDir: "/tmp/test-profile"}
	l := newLauncher(cfg)

	// When ChromeBin is empty, rod's default bin should still be present.
	if !l.Has(flags.Bin) {
		t.Fatal("expected bin flag to be set by rod default")
	}
}
