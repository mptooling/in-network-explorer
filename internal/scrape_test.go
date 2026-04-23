package explorer_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/testdouble"
)

var nopLog = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestScrapeUseCase_DiscoverAndWarmUp(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	browser := &testdouble.FakeBrowserClient{
		SearchURLs: []string{"https://linkedin.com/in/alice", "https://linkedin.com/in/bob"},
		ProfileDataByURL: map[string]explorer.ProfileData{
			"https://linkedin.com/in/alice": {URL: "https://linkedin.com/in/alice", Slug: "alice", Name: "Alice"},
			"https://linkedin.com/in/bob":   {URL: "https://linkedin.com/in/bob", Slug: "bob", Name: "Bob"},
		},
	}
	limiter := testdouble.NewFakeRateLimiter(100)
	uc := explorer.NewScrapeUseCase(repo, browser, limiter, nopLog, 10)

	if err := uc.Run(context.Background(), "engineer", "Berlin"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Both profiles should be saved as Scanned.
	for _, url := range browser.SearchURLs {
		p, err := repo.Get(context.Background(), url)
		if err != nil {
			t.Fatalf("Get(%q): %v", url, err)
		}
		if p.State != explorer.StateScanned {
			t.Errorf("%s state = %v, want Scanned", url, p.State)
		}
	}
	if len(browser.VisitedURLs) != 2 {
		t.Errorf("visited %d profiles, want 2", len(browser.VisitedURLs))
	}
}

func TestScrapeUseCase_WarmUpLikesAndTransitions(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	// Pre-populate a due scanned prospect.
	p := &explorer.Prospect{
		ProfileURL:   "https://linkedin.com/in/carol",
		State:        explorer.StateScanned,
		NextActionAt: time.Now().Add(-1 * time.Hour),
	}
	_ = repo.Save(context.Background(), p)

	browser := &testdouble.FakeBrowserClient{
		SearchURLs: nil, // No new searches.
	}
	limiter := testdouble.NewFakeRateLimiter(100)
	uc := explorer.NewScrapeUseCase(repo, browser, limiter, nopLog, 10)

	if err := uc.Run(context.Background(), "", ""); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, _ := repo.Get(context.Background(), p.ProfileURL)
	if got.State != explorer.StateLiked {
		t.Errorf("state = %v, want Liked", got.State)
	}
	if len(browser.LikedURLs) != 1 {
		t.Errorf("liked %d, want 1", len(browser.LikedURLs))
	}
}

func TestScrapeUseCase_RateLimitStopsDiscovery(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	browser := &testdouble.FakeBrowserClient{
		SearchURLs: []string{"https://linkedin.com/in/a", "https://linkedin.com/in/b", "https://linkedin.com/in/c"},
		ProfileDataByURL: map[string]explorer.ProfileData{
			"https://linkedin.com/in/a": {URL: "https://linkedin.com/in/a", Slug: "a", Name: "A"},
			"https://linkedin.com/in/b": {URL: "https://linkedin.com/in/b", Slug: "b", Name: "B"},
			"https://linkedin.com/in/c": {URL: "https://linkedin.com/in/c", Slug: "c", Name: "C"},
		},
	}
	limiter := testdouble.NewFakeRateLimiter(1) // Only 1 allowed.
	uc := explorer.NewScrapeUseCase(repo, browser, limiter, nopLog, 10)

	if err := uc.Run(context.Background(), "q", "l"); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Should have visited only 1 profile before rate limit kicked in.
	if len(browser.VisitedURLs) != 1 {
		t.Errorf("visited %d profiles, want 1", len(browser.VisitedURLs))
	}
}

func TestScrapeUseCase_BlockDetected(t *testing.T) {
	browser := &testdouble.FakeBrowserClient{Block: explorer.BlockAuthwall}
	uc := explorer.NewScrapeUseCase(nil, browser, nil, nopLog, 10)

	err := uc.Run(context.Background(), "q", "l")
	if !errors.Is(err, explorer.ErrBlockDetected) {
		t.Fatalf("expected ErrBlockDetected, got %v", err)
	}
}
