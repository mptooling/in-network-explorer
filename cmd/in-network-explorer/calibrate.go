package main

import (
	"context"
	"log/slog"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/config"
)

func runCalibrate(ctx context.Context, cfg config.Config, log *slog.Logger) {
	repo, _, err := buildDynamoDeps(ctx, cfg, log)
	if err != nil {
		return
	}

	// LLMClient adapter is not yet implemented. When internal/bedrock is
	// ready, construct it here and pass to NewCalibrateUseCase.
	log.ErrorContext(ctx, "LLM adapter not yet implemented — calibrate requires internal/bedrock")
	_ = explorer.NewCalibrateUseCase(repo, nil, log, 10)
}
