package explorer_test

import (
	"context"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

// Compile-time check: fakeLLM must satisfy LLMClient.

var _ explorer.LLMClient = (*fakeLLM)(nil)

type fakeLLM struct{}

func (f *fakeLLM) ScoreAndDraft(_ context.Context, _ *explorer.Prospect, _ []explorer.Prospect) (explorer.ScoreResult, error) {
	return explorer.ScoreResult{}, nil
}

func (f *fakeLLM) Critique(_ context.Context, _ string) (int, error) {
	return 0, nil
}
