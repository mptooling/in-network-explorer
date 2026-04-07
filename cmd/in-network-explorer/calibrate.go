package main

import (
	"context"
	"log/slog"

	"github.com/pavlomaksymov/in-network-explorer/infrastructure"
)

func runCalibrate(ctx context.Context, cfg infrastructure.Config, log *slog.Logger) {
	log.InfoContext(ctx, "not yet implemented", "command", "calibrate")
}
