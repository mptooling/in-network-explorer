package explorer_test

import (
	"context"
	"testing"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/testdouble"
)

func TestReportUseCase_GeneratesEntries(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	prospects := []*explorer.Prospect{
		{ProfileURL: "https://linkedin.com/in/a", Name: "Alice", State: explorer.StateDrafted, WorthinessScore: 9, DraftedMessage: "Hi Alice"},
		{ProfileURL: "https://linkedin.com/in/b", Name: "Bob", State: explorer.StateDrafted, WorthinessScore: 7, DraftedMessage: "Hi Bob"},
		{ProfileURL: "https://linkedin.com/in/c", Name: "Carol", State: explorer.StateDrafted, WorthinessScore: 5, DraftedMessage: "Hi Carol"},
	}
	for _, p := range prospects {
		_ = repo.Save(context.Background(), p)
	}

	uc := explorer.NewReportUseCase(repo, nopLog, 2)
	entries, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(entries))
	}
	if entries[0].WorthinessScore < entries[1].WorthinessScore {
		t.Error("entries not sorted by score descending")
	}
	if entries[0].DraftedMessage == "" {
		t.Error("drafted message is empty")
	}
}

func TestReportUseCase_EmptyRepo(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	uc := explorer.NewReportUseCase(repo, nopLog, 10)

	entries, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("entries = %d, want 0", len(entries))
	}
}
