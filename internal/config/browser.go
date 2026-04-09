package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/launcher/flags"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

// newLauncher builds a rod Launcher configured with all anti-detection flags.
// It does not launch the browser — call l.Launch() separately.
func newLauncher(cfg Config) *launcher.Launcher {
	l := launcher.New().
		Headless(true).
		NoSandbox(true).
		Set("disable-blink-features", "AutomationControlled").
		Set("window-size", "1920,1080").
		Set("disable-webrtc").
		Set("use-gl", "angle").
		Set("use-angle", "swiftshader").
		Set("lang", "en-US").
		Delete("enable-automation").
		UserDataDir(cfg.ChromeProfileDir).
		Set(flags.KeepUserDataDir)

	if cfg.ChromeBin != "" {
		l.Bin(cfg.ChromeBin)
	}
	if cfg.ProxyAddr != "" {
		l.Proxy(cfg.ProxyAddr)
	}
	return l
}

// NewBrowser launches a Chromium instance with anti-detection configuration
// derived from cfg. The returned cleanup function closes the browser and must
// be called when the browser is no longer needed.
func NewBrowser(ctx context.Context, cfg Config) (*rod.Browser, func(), error) {
	u, err := newLauncher(cfg).Launch()
	if err != nil {
		return nil, nil, fmt.Errorf("launch browser: %w", err)
	}

	browser := rod.New().ControlURL(u).Context(ctx)
	if err := browser.Connect(); err != nil {
		return nil, nil, fmt.Errorf("connect browser: %w", err)
	}

	if err := browser.IgnoreCertErrors(true); err != nil {
		browser.MustClose()
		return nil, nil, fmt.Errorf("ignore cert errors: %w", err)
	}

	if cfg.ProxyAddr != "" {
		go browser.MustHandleAuth(cfg.ProxyUser, cfg.ProxyPass)()
	}

	cleanup := func() { browser.MustClose() }
	return browser, cleanup, nil
}

// customPatchesJS overrides screen properties and adds AudioContext noise to
// defeat fingerprinting that go-rod/stealth does not cover.
// NOTE: This must be raw statements, NOT wrapped in () => { ... }, because
// Chrome's addScriptToEvaluateOnNewDocument evaluates source as a script.
const customPatchesJS = `{
    const d = Object.defineProperty.bind(Object);
    d(screen, 'width',       { get: () => 1920 });
    d(screen, 'height',      { get: () => 1080 });
    d(screen, 'availWidth',  { get: () => 1920 });
    d(screen, 'availHeight', { get: () => 1040 });
    d(screen, 'colorDepth',  { get: () => 24 });
    d(screen, 'pixelDepth',  { get: () => 24 });

    const seed = Date.now() % 1000;
    const OrigAudioBuffer = AudioBuffer;
    AudioBuffer.prototype.getChannelData = function(channel) {
        const data = OrigAudioBuffer.prototype.getChannelData.call(this, channel);
        for (let i = 0; i < data.length; i += 100) {
            data[i] += (seed / 1000000) * 0.0001;
        }
        return data;
    };
}`

// NewStealthPage creates a new page with go-rod/stealth patches and custom
// EvalOnNewDocument evasions applied. Always use this instead of
// browser.MustPage() to avoid bot detection.
func NewStealthPage(browser *rod.Browser) (*rod.Page, error) {
	page, err := stealth.Page(browser)
	if err != nil {
		return nil, fmt.Errorf("create stealth page: %w", err)
	}
	if err := applyCustomPatches(page); err != nil {
		return nil, fmt.Errorf("apply custom patches: %w", err)
	}
	return page, nil
}

func applyCustomPatches(page *rod.Page) error {
	_, err := page.EvalOnNewDocument(customPatchesJS)
	return err
}

// knownExtensions maps Chrome Web Store extension IDs to human-readable names.
// LinkedIn's BrowserGate fires extension probes — a sterile profile with zero
// hits is a detectable signal. Limit to 5 to avoid inverse detection.
var knownExtensions = map[string]string{
	"cjpalhdlnbpafiamejdnhcphjbkeiagm": "uBlock Origin",
	"ghbmnnjooekpmoecnnnilnnbdlolhkhi": "Google Docs Offline",
	"kbfnbcaeplbcioakkpcpgfkobkghlhen": "Grammarly",
	"aeblfdkhhhdcdjpifhhbdiojplfjncoa": "1Password",
	"hdokiejnpimakedhajhdlcegeplioahd": "LastPass",
}

// SetupExtensionHijack installs a HijackRouter that responds to
// chrome-extension:// probes for known extensions and refuses unknown ones.
// The caller must call router.Stop() when done.
func SetupExtensionHijack(browser *rod.Browser) (*rod.HijackRouter, error) {
	router := browser.HijackRequests()
	err := router.Add("chrome-extension://*/*", "", func(ctx *rod.Hijack) {
		handleExtensionRequest(ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("add extension hijack: %w", err)
	}
	go router.Run()
	return router, nil
}

func handleExtensionRequest(ctx *rod.Hijack) {
	extID := ctx.Request.URL().Host
	name, ok := knownExtensions[extID]
	if !ok {
		ctx.Response.Fail(proto.NetworkErrorReasonConnectionRefused)
		return
	}

	if strings.HasSuffix(ctx.Request.URL().Path, "manifest.json") {
		ctx.Response.SetHeader("Content-Type", "application/json")
		ctx.Response.SetBody(fmt.Sprintf(
			`{"manifest_version":3,"name":%q,"version":"1.0"}`, name))
		return
	}
	ctx.Response.SetHeader("Content-Type", "application/javascript")
	ctx.Response.SetBody("// ok")
}
