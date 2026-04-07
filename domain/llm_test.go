package domain_test

import (
	"context"

	"github.com/pavlomaksymov/in-network-explorer/domain"
)

// Compile-time check: fakeLLM must satisfy LLMClient.

var _ domain.LLMClient = (*fakeLLM)(nil)

type fakeLLM struct{}

func (f *fakeLLM) ScoreAndDraft(_ context.Context, _ *domain.Prospect, _ []domain.Prospect) (domain.ScoreResult, error) {
	return domain.ScoreResult{}, nil
}

func (f *fakeLLM) Critique(_ context.Context, _ string) (int, error) {
	return 0, nil
}
