package main

import (
	"context"
	"log/slog"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/bedrock"
	"github.com/pavlomaksymov/in-network-explorer/internal/config"
)

func runCalibrate(ctx context.Context, cfg config.Config, log *slog.Logger) {
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

	uc := explorer.NewCalibrateUseCase(repo, llm, log, 10)
	result, err := uc.Run(ctx)
	if err != nil {
		log.ErrorContext(ctx, "calibrate failed", "error", err)
		return
	}
	log.InfoContext(ctx, "calibration result",
		"checked", result.Checked,
		"mean_diff", result.MeanDiff,
		"max_diff", result.MaxDiff,
	)
}
