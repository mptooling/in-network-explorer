package main

import (
	"context"
	"log/slog"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/bedrock"
	"github.com/pavlomaksymov/in-network-explorer/internal/config"
)

func runAnalyze(ctx context.Context, cfg config.Config, log *slog.Logger) {
	repo, _, err := buildDynamoDeps(ctx, cfg, log)
	if err != nil {
		return
	}

	bc, err := config.NewBedrockClient(ctx, cfg)
	if err != nil {
		log.ErrorContext(ctx, "bedrock client", "error", err)
		return
	}
	llm := bedrock.NewClient(bc, cfg.BedrockModelID)

	uc := explorer.NewAnalyzeUseCase(repo, llm, log, cfg.AnalyzeConcurrency, nil, nil)
	if err := uc.Run(ctx); err != nil {
		log.ErrorContext(ctx, "analyze failed", "error", err)
	}
}
