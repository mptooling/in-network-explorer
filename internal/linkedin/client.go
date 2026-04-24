package linkedin

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/config"
)

var _ explorer.BrowserClient = (*Client)(nil)

// Client implements BrowserClient using Rod for headless browser automation
// against LinkedIn. It holds a single stealth page that is reused across
// operations, mimicking a real user in one browser tab.
type Client struct {
	browser *rod.Browser
	page    *rod.Page
	router  *rod.HijackRouter
	log     *slog.Logger
}

// New creates a Client with anti-detection stealth page and injected session
// cookies. The caller owns the browser and must call Client.Close when done.
func New(browser *rod.Browser, cookies string, log *slog.Logger) (*Client, error) {
	router, err := config.SetupExtensionHijack(browser)
	if err != nil {
		return nil, fmt.Errorf("setup extension hijack: %w", err)
	}

	page, err := config.NewStealthPage(browser)
	if err != nil {
		_ = router.Stop()
		return nil, fmt.Errorf("create stealth page: %w", err)
	}

	params := parseCookies(cookies)
	if err := page.SetCookies(params); err != nil {
		_ = router.Stop()
		_ = page.Close()
		return nil, fmt.Errorf("set cookies: %w", err)
	}

	return &Client{browser: browser, page: page, router: router, log: log}, nil
}

// Close releases the page and stops the extension hijack router. The browser
// itself is not closed — that is the caller's responsibility.
func (c *Client) Close() error {
	_ = c.router.Stop()
	return c.page.Close()
}

// CheckBlock detects whether the current browser session is restricted by
// LinkedIn. It navigates to the feed page and inspects the result.
func (c *Client) CheckBlock(ctx context.Context) (explorer.BlockType, error) {
	if err := c.navigate(ctx, baseURL+feedPath); err != nil {
		return explorer.BlockNone, fmt.Errorf("navigate to feed: %w", err)
	}
	result, err := c.page.Eval(detectBlockJS)
	if err != nil {
		return explorer.BlockNone, fmt.Errorf("detect block: %w", err)
	}
	switch result.Value.Str() {
	case "authwall":
		return explorer.BlockAuthwall, nil
	case "challenge":
		return explorer.BlockChallenge, nil
	case "soft_empty":
		return explorer.BlockSoftEmpty, nil
	default:
		return explorer.BlockNone, nil
	}
}

// navigate goes to url, waits for the network to idle, and respects context.
func (c *Client) navigate(ctx context.Context, url string) error {
	return rod.Try(func() {
		c.page.Context(ctx).
			Timeout(pageLoadTimeout).
			MustNavigate(url).
			MustWaitDOMStable()
	})
}

// ── cookie helpers ──────────────────────────────────────────────────────────

// parseCookies splits a raw cookie string (e.g. "li_at=ABC; JSESSIONID=xyz")
// into CDP cookie parameters scoped to .linkedin.com.
func parseCookies(raw string) []*proto.NetworkCookieParam {
	var params []*proto.NetworkCookieParam
	for _, part := range strings.Split(raw, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		params = append(params, &proto.NetworkCookieParam{
			Name:   strings.TrimSpace(name),
			Value:  strings.TrimSpace(value),
			Domain: ".linkedin.com",
			Path:   "/",
		})
	}
	return params
}

// slugFromURL extracts the slug from a LinkedIn profile URL.
// Example: "https://www.linkedin.com/in/alice-smith" becomes "alice-smith".
func slugFromURL(profileURL string) string {
	profileURL = strings.TrimRight(profileURL, "/")
	idx := strings.LastIndex(profileURL, "/in/")
	if idx < 0 {
		return ""
	}
	return profileURL[idx+4:]
}
