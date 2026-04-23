package explorer_test

import (
	"context"
	"testing"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/testdouble"
)

func TestCalibrateUseCase_ReportsVariance(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	_ = repo.Save(context.Background(), &explorer.Prospect{
		ProfileURL:      "https://linkedin.com/in/a",
		State:           explorer.StateDrafted,
		WorthinessScore: 8,
	})
	_ = repo.Save(context.Background(), &explorer.Prospect{
		ProfileURL:      "https://linkedin.com/in/b",
		State:           explorer.StateDrafted,
		WorthinessScore: 6,
	})

	llm := &testdouble.FakeLLMClient{
		ScoreResult: explorer.ScoreResult{Score: 7, Message: "msg"},
	}
	uc := explorer.NewCalibrateUseCase(repo, llm, nopLog, 10)

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Checked != 2 {
		t.Errorf("Checked = %d, want 2", result.Checked)
	}
	// Original scores: 8, 6. New score: 7 for both. Diffs: 1, 1.
	if result.MeanDiff != 1.0 {
		t.Errorf("MeanDiff = %f, want 1.0", result.MeanDiff)
	}
	if result.MaxDiff != 1 {
		t.Errorf("MaxDiff = %d, want 1", result.MaxDiff)
	}
}

func TestCalibrateUseCase_NoProspects(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	llm := &testdouble.FakeLLMClient{}
	uc := explorer.NewCalibrateUseCase(repo, llm, nopLog, 5)

	result, err := uc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if result.Checked != 0 {
		t.Errorf("Checked = %d, want 0", result.Checked)
	}
}
