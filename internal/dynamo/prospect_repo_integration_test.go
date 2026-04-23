//go:build integration

package dynamo

import (
	"context"
	"errors"
	"testing"
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

func TestIntegration_ProspectRepo_SaveAndGet(t *testing.T) {
	client, table := integrationClient(t)
	repo := NewProspectRepository(client, table)
	ctx := context.Background()

	p := &explorer.Prospect{
		ProfileURL:      "https://linkedin.com/in/test-save-" + t.Name(),
		Slug:            "test-save",
		Name:            "Test User",
		Headline:        "Engineer",
		Location:        "Berlin",
		State:           explorer.StateScanned,
		WorthinessScore: 7,
		RecentPosts:     []string{"post1"},
		CreatedAt:       time.Now().UTC().Truncate(time.Second),
		LastActionAt:    time.Now().UTC().Truncate(time.Second),
		NextActionAt:    time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second),
	}

	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.Get(ctx, p.ProfileURL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Test User" {
		t.Errorf("Name = %q, want %q", got.Name, "Test User")
	}
	if got.State != explorer.StateScanned {
		t.Errorf("State = %v, want %v", got.State, explorer.StateScanned)
	}
}

func TestIntegration_ProspectRepo_Get_NotFound(t *testing.T) {
	client, table := integrationClient(t)
	repo := NewProspectRepository(client, table)

	_, err := repo.Get(context.Background(), "https://linkedin.com/in/nonexistent")
	if !errors.Is(err, explorer.ErrNotFound) {
		t.Fatalf("Get: got %v, want ErrNotFound", err)
	}
}

func TestIntegration_ProspectRepo_InsertIfNew(t *testing.T) {
	client, table := integrationClient(t)
	repo := NewProspectRepository(client, table)
	ctx := context.Background()

	p := &explorer.Prospect{
		ProfileURL: "https://linkedin.com/in/test-insert-" + t.Name(),
		State:      explorer.StateScanned,
	}

	inserted, err := repo.InsertIfNew(ctx, p)
	if err != nil {
		t.Fatalf("first InsertIfNew: %v", err)
	}
	if !inserted {
		t.Fatal("expected inserted=true")
	}

	inserted, err = repo.InsertIfNew(ctx, p)
	if err != nil {
		t.Fatalf("second InsertIfNew: %v", err)
	}
	if inserted {
		t.Fatal("expected inserted=false on duplicate")
	}
}

func TestIntegration_ProspectRepo_ListByState(t *testing.T) {
	client, table := integrationClient(t)
	repo := NewProspectRepository(client, table)
	ctx := context.Background()
	now := time.Now().UTC()
	suffix := t.Name()

	due := &explorer.Prospect{
		ProfileURL:   "https://linkedin.com/in/due-" + suffix,
		State:        explorer.StateScanned,
		NextActionAt: now.Add(-1 * time.Hour),
	}
	future := &explorer.Prospect{
		ProfileURL:   "https://linkedin.com/in/future-" + suffix,
		State:        explorer.StateScanned,
		NextActionAt: now.Add(48 * time.Hour),
	}
	for _, p := range []*explorer.Prospect{due, future} {
		if err := repo.Save(ctx, p); err != nil {
			t.Fatal(err)
		}
	}

	results, err := repo.ListByState(ctx, explorer.StateScanned, now)
	if err != nil {
		t.Fatalf("ListByState: %v", err)
	}

	found := false
	for _, r := range results {
		if r.ProfileURL == due.ProfileURL {
			found = true
		}
		if r.ProfileURL == future.ProfileURL {
			t.Fatal("future prospect should not be returned")
		}
	}
	if !found {
		t.Fatal("due prospect not found in results")
	}
}

func TestIntegration_ProspectRepo_ListByStateOrderedByScore(t *testing.T) {
	client, table := integrationClient(t)
	repo := NewProspectRepository(client, table)
	ctx := context.Background()
	suffix := t.Name()

	prospects := []*explorer.Prospect{
		{ProfileURL: "https://linkedin.com/in/s5-" + suffix, State: explorer.StateDrafted, WorthinessScore: 5},
		{ProfileURL: "https://linkedin.com/in/s9-" + suffix, State: explorer.StateDrafted, WorthinessScore: 9},
		{ProfileURL: "https://linkedin.com/in/s3-" + suffix, State: explorer.StateDrafted, WorthinessScore: 3},
	}
	for _, p := range prospects {
		if err := repo.Save(ctx, p); err != nil {
			t.Fatal(err)
		}
	}

	results, err := repo.ListByStateOrderedByScore(ctx, explorer.StateDrafted, 2)
	if err != nil {
		t.Fatalf("ListByStateOrderedByScore: %v", err)
	}

	// We asked for limit=2, so filter to our test data (other tests may have Drafted items).
	var ours []*explorer.Prospect
	for _, r := range results {
		for _, p := range prospects {
			if r.ProfileURL == p.ProfileURL {
				ours = append(ours, r)
			}
		}
	}
	if len(ours) < 2 {
		t.Fatalf("expected at least 2 of our prospects, got %d", len(ours))
	}
	if ours[0].WorthinessScore < ours[1].WorthinessScore {
		t.Fatalf("results not sorted: %d < %d", ours[0].WorthinessScore, ours[1].WorthinessScore)
	}
}
