package explorer_test

import (
	"context"
	"testing"
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/testdouble"
)

func TestAnalyzeUseCase_ScoresAndDrafts(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	p := &explorer.Prospect{
		ProfileURL:   "https://linkedin.com/in/alice",
		State:        explorer.StateLiked,
		NextActionAt: time.Now().Add(-1 * time.Hour),
	}
	_ = repo.Save(context.Background(), p)

	llm := &testdouble.FakeLLMClient{
		ScoreResult: explorer.ScoreResult{
			Score:         8,
			Reasoning:     "Great fit",
			Message:       "Hi Alice, love your work on distributed systems!",
			CritiqueScore: 12,
		},
	}
	uc := explorer.NewAnalyzeUseCase(repo, llm, nopLog, 0)

	if err := uc.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, _ := repo.Get(context.Background(), p.ProfileURL)
	if got.State != explorer.StateDrafted {
		t.Errorf("state = %v, want Drafted", got.State)
	}
	if got.WorthinessScore != 8 {
		t.Errorf("score = %d, want 8", got.WorthinessScore)
	}
	if got.DraftedMessage == "" {
		t.Error("drafted message is empty")
	}
}

func TestAnalyzeUseCase_SkipsNotDue(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	p := &explorer.Prospect{
		ProfileURL:   "https://linkedin.com/in/bob",
		State:        explorer.StateLiked,
		NextActionAt: time.Now().Add(48 * time.Hour), // Not due yet.
	}
	_ = repo.Save(context.Background(), p)

	llm := &testdouble.FakeLLMClient{
		ScoreResult: explorer.ScoreResult{Score: 5, Message: "msg"},
	}
	uc := explorer.NewAnalyzeUseCase(repo, llm, nopLog, 0)

	if err := uc.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, _ := repo.Get(context.Background(), p.ProfileURL)
	if got.State != explorer.StateLiked {
		t.Errorf("state = %v, want Liked (unchanged)", got.State)
	}
}

func TestAnalyzeUseCase_UsesFewShotExamples(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()

	// Existing drafted prospect as example.
	example := &explorer.Prospect{
		ProfileURL:      "https://linkedin.com/in/example",
		State:           explorer.StateDrafted,
		WorthinessScore: 9,
	}
	_ = repo.Save(context.Background(), example)

	// Prospect to analyze.
	p := &explorer.Prospect{
		ProfileURL:   "https://linkedin.com/in/target",
		State:        explorer.StateLiked,
		NextActionAt: time.Now().Add(-1 * time.Hour),
	}
	_ = repo.Save(context.Background(), p)

	llm := &testdouble.FakeLLMClient{
		ScoreResult: explorer.ScoreResult{Score: 7, Message: "msg", CritiqueScore: 10},
	}
	uc := explorer.NewAnalyzeUseCase(repo, llm, nopLog, 3)

	if err := uc.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	got, _ := repo.Get(context.Background(), p.ProfileURL)
	if got.State != explorer.StateDrafted {
		t.Errorf("state = %v, want Drafted", got.State)
	}
}

func TestAnalyzeUseCase_NoProspects(t *testing.T) {
	repo := testdouble.NewFakeProspectRepository()
	llm := &testdouble.FakeLLMClient{}
	uc := explorer.NewAnalyzeUseCase(repo, llm, nopLog, 0)

	if err := uc.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
}
