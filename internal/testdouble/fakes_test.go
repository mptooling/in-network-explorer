package testdouble_test

import (
	"context"
	"testing"
	"time"

	"github.com/pavlomaksymov/in-network-explorer/domain"
	"github.com/pavlomaksymov/in-network-explorer/internal/testdouble"
)

// ── FakeProspectRepository ──────────────────────────────────────────────────

func TestFakeProspectRepository_SaveAndGet(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	p := &domain.Prospect{
		ProfileURL: "https://linkedin.com/in/alice",
		State:      domain.StateScanned,
	}
	if err := repo.Save(context.Background(), p); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := repo.Get(context.Background(), p.ProfileURL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ProfileURL != p.ProfileURL {
		t.Fatalf("got ProfileURL %q, want %q", got.ProfileURL, p.ProfileURL)
	}
}

func TestFakeProspectRepository_Get_NotFound(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	_, err := repo.Get(context.Background(), "https://linkedin.com/in/nobody")
	if err == nil {
		t.Fatal("expected ErrNotFound, got nil")
	}
}

func TestFakeProspectRepository_InsertIfNew(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	p := &domain.Prospect{ProfileURL: "https://linkedin.com/in/bob", State: domain.StateScanned}

	inserted, err := repo.InsertIfNew(context.Background(), p)
	if err != nil {
		t.Fatalf("first InsertIfNew: %v", err)
	}
	if !inserted {
		t.Fatal("expected inserted=true on first call")
	}

	inserted, err = repo.InsertIfNew(context.Background(), p)
	if err != nil {
		t.Fatalf("second InsertIfNew: %v", err)
	}
	if inserted {
		t.Fatal("expected inserted=false on duplicate")
	}
}

func TestFakeProspectRepository_ListByState(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	now := time.Now()

	scannedDue := &domain.Prospect{
		ProfileURL:   "https://linkedin.com/in/due",
		State:        domain.StateScanned,
		NextActionAt: now.Add(-1 * time.Hour), // overdue
	}
	scannedFuture := &domain.Prospect{
		ProfileURL:   "https://linkedin.com/in/future",
		State:        domain.StateScanned,
		NextActionAt: now.Add(24 * time.Hour), // not due yet
	}
	liked := &domain.Prospect{
		ProfileURL:   "https://linkedin.com/in/liked",
		State:        domain.StateLiked,
		NextActionAt: now.Add(-1 * time.Hour),
	}

	for _, p := range []*domain.Prospect{scannedDue, scannedFuture, liked} {
		if err := repo.Save(context.Background(), p); err != nil {
			t.Fatal(err)
		}
	}

	results, err := repo.ListByState(context.Background(), domain.StateScanned, now)
	if err != nil {
		t.Fatalf("ListByState: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ProfileURL != scannedDue.ProfileURL {
		t.Fatalf("unexpected prospect %q", results[0].ProfileURL)
	}
}

func TestFakeProspectRepository_ListByStateOrderedByScore(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()

	prospects := []*domain.Prospect{
		{ProfileURL: "https://linkedin.com/in/score5", State: domain.StateDrafted, WorthinessScore: 5},
		{ProfileURL: "https://linkedin.com/in/score9", State: domain.StateDrafted, WorthinessScore: 9},
		{ProfileURL: "https://linkedin.com/in/score3", State: domain.StateDrafted, WorthinessScore: 3},
	}
	for _, p := range prospects {
		if err := repo.Save(context.Background(), p); err != nil {
			t.Fatal(err)
		}
	}

	results, err := repo.ListByStateOrderedByScore(context.Background(), domain.StateDrafted, 10)
	if err != nil {
		t.Fatalf("ListByStateOrderedByScore: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3, got %d", len(results))
	}
	if results[0].WorthinessScore != 9 || results[1].WorthinessScore != 5 || results[2].WorthinessScore != 3 {
		t.Fatalf("wrong order: %v", results)
	}
}

// ── FakeRateLimiter ─────────────────────────────────────────────────────────

func TestFakeRateLimiter_AcquireAndCurrent(t *testing.T) {
	rl := testdouble.NewFakeRateLimiter(3)
	ctx := context.Background()

	for i := range 3 {
		if err := rl.Acquire(ctx, "profile_views"); err != nil {
			t.Fatalf("Acquire %d: %v", i, err)
		}
	}
	count, err := rl.Current(ctx, "profile_views")
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count=3, got %d", count)
	}
}

func TestFakeRateLimiter_ExceedsMax(t *testing.T) {
	rl := testdouble.NewFakeRateLimiter(2)
	ctx := context.Background()

	_ = rl.Acquire(ctx, "scope")
	_ = rl.Acquire(ctx, "scope")

	err := rl.Acquire(ctx, "scope")
	if err == nil {
		t.Fatal("expected ErrRateLimitExceeded, got nil")
	}
}

func TestFakeRateLimiter_ScopesAreIndependent(t *testing.T) {
	rl := testdouble.NewFakeRateLimiter(1)
	ctx := context.Background()

	if err := rl.Acquire(ctx, "profile_views"); err != nil {
		t.Fatal(err)
	}
	// Different scope should not be affected.
	if err := rl.Acquire(ctx, "connection_requests"); err != nil {
		t.Fatalf("unexpected error for different scope: %v", err)
	}
}
