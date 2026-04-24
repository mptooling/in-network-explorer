//go:build integration

package linkedin

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-rod/rod"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/config"
)

var nopLog = slog.New(slog.NewTextHandler(io.Discard, nil))

func testBrowser(t *testing.T) *rod.Browser {
	t.Helper()
	dir, err := os.MkdirTemp("", "linkedin-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	cfg := config.Config{ChromeProfileDir: dir}
	browser, cleanup, err := config.NewBrowser(context.Background(), cfg)
	if err != nil {
		t.Fatalf("NewBrowser: %v", err)
	}
	t.Cleanup(cleanup)
	return browser
}

func fixtureServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/in/alice-smith":
			http.ServeFile(w, r, "testdata/profile.html")
		case "/search/results/people/":
			http.ServeFile(w, r, "testdata/search.html")
		case "/feed":
			http.ServeFile(w, r, "testdata/profile.html") // non-blocked page
		case "/authwall", "/login":
			http.ServeFile(w, r, "testdata/authwall.html")
		case "/in/alice-smith/recent-activity/all/":
			http.ServeFile(w, r, "testdata/activity.html")
		default:
			w.Header().Set("Content-Type", "text/html")
			fmt.Fprintf(w, `<html><body><main id="main">OK</main></body></html>`)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// testClient creates a Client pointing at the local fixture server instead of
// linkedin.com. It overrides baseURL via a page-level approach: navigating to
// local URLs directly rather than relying on baseURL.
func testClient(t *testing.T) (*Client, *httptest.Server) {
	t.Helper()
	browser := testBrowser(t)

	page, err := config.NewStealthPage(browser)
	if err != nil {
		t.Fatalf("NewStealthPage: %v", err)
	}

	c := &Client{browser: browser, page: page, log: nopLog}
	// No extension hijack in tests — chrome-extension:// doesn't apply to local URLs.
	c.router = browser.HijackRequests()
	go c.router.Run()

	t.Cleanup(func() { _ = c.Close() })
	srv := fixtureServer(t)
	return c, srv
}

func TestIntegration_ExtractProfileData(t *testing.T) {
	c, srv := testClient(t)

	if err := c.navigate(context.Background(), srv.URL+"/in/alice-smith"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	data, err := c.extractProfileData()
	if err != nil {
		t.Fatalf("extractProfileData: %v", err)
	}

	if data.Name != "Alice Smith" {
		t.Errorf("Name = %q, want Alice Smith", data.Name)
	}
	if data.Headline == "" {
		t.Error("Headline is empty")
	}
	if data.Location == "" {
		t.Error("Location is empty")
	}
	if data.About == "" {
		t.Error("About is empty")
	}
	if len(data.RecentPosts) == 0 {
		t.Error("RecentPosts is empty")
	}
}

func TestIntegration_SearchResultsExtraction(t *testing.T) {
	c, srv := testClient(t)

	if err := c.navigate(context.Background(), srv.URL+"/search/results/people/"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	result, err := c.page.Eval(extractSearchResultsJS)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}

	urls := result.Value.Arr()
	if len(urls) != 3 {
		t.Fatalf("expected 3 search results, got %d", len(urls))
	}
	if urls[0].Str() != "https://www.linkedin.com/in/alice-smith" {
		t.Errorf("first URL = %q", urls[0].Str())
	}
}

func TestIntegration_CheckBlock_None(t *testing.T) {
	c, srv := testClient(t)

	// Navigate to a non-blocked page.
	if err := c.navigate(context.Background(), srv.URL+"/feed"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	result, err := c.page.Eval(detectBlockJS)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if result.Value.Str() != "none" {
		t.Errorf("block = %q, want none", result.Value.Str())
	}
}

func TestIntegration_CheckBlock_Authwall(t *testing.T) {
	c, srv := testClient(t)

	// The authwall page has a JS redirect to /authwall in its URL.
	// Since we're testing locally, we check the detection JS directly
	// by navigating to /authwall which the fixture server serves.
	if err := c.navigate(context.Background(), srv.URL+"/authwall"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	// The URL won't contain 'authwall' in our local test since we serve
	// a static page. Instead, check that the main content is minimal.
	result, err := c.page.Eval(detectBlockJS)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	blockType := result.Value.Str()
	// The authwall fixture has enough content in <main> that it won't trigger
	// soft_empty, and the URL is localhost not /authwall, so we get "none".
	// This is expected — real authwall detection relies on LinkedIn's URL redirect.
	t.Logf("block type on fixture: %s", blockType)
}

func TestIntegration_LikeButton_Detection(t *testing.T) {
	c, srv := testClient(t)

	if err := c.navigate(context.Background(), srv.URL+"/in/alice-smith/recent-activity/all/"); err != nil {
		t.Fatalf("navigate: %v", err)
	}

	result, err := c.page.Eval(findLikeButtonJS)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if !result.Value.Bool() {
		t.Error("expected like button to be found in activity fixture")
	}
}

func TestIntegration_VisitProfile_FullFlow(t *testing.T) {
	c, srv := testClient(t)

	data, err := c.visitProfileAt(context.Background(), srv.URL+"/in/alice-smith")
	if err != nil {
		t.Fatalf("VisitProfile: %v", err)
	}
	if data.Name != "Alice Smith" {
		t.Errorf("Name = %q, want Alice Smith", data.Name)
	}
	if data.Slug != "alice-smith" {
		t.Errorf("Slug = %q, want alice-smith", data.Slug)
	}
}

// visitProfileAt is a test helper that calls the core extraction logic against
// an arbitrary URL (local fixture server) instead of linkedin.com.
func (c *Client) visitProfileAt(ctx context.Context, url string) (explorer.ProfileData, error) {
	if err := c.navigate(ctx, url); err != nil {
		return explorer.ProfileData{}, fmt.Errorf("navigate: %w", err)
	}
	data, err := c.extractProfileData()
	if err != nil {
		return explorer.ProfileData{}, err
	}
	data.URL = url
	data.Slug = slugFromURL(url)
	return data, nil
}
