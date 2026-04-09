package main

import (
	"context"
	"log/slog"

	"github.com/pavlomaksymov/in-network-explorer/internal/config"
)

func runAnalyze(ctx context.Context, cfg config.Config, log *slog.Logger) {
	log.InfoContext(ctx, "not yet implemented", "command", "analyze")
}
