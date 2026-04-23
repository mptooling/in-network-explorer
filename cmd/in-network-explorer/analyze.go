package main

import (
	"context"
	"log/slog"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/config"
)

func runAnalyze(ctx context.Context, cfg config.Config, log *slog.Logger) {
	repo, _, err := buildDynamoDeps(ctx, cfg, log)
	if err != nil {
		return
	}

	// LLMClient adapter is not yet implemented. When internal/bedrock is
	// ready, construct it here and pass to NewAnalyzeUseCase.
	log.ErrorContext(ctx, "LLM adapter not yet implemented — analyze requires internal/bedrock")
	_ = explorer.NewAnalyzeUseCase(repo, nil, log, cfg.AnalyzeConcurrency)
}
