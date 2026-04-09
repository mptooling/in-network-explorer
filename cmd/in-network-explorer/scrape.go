package main

import (
	"context"
	"log/slog"

	"github.com/pavlomaksymov/in-network-explorer/internal/config"
)

func runScrape(ctx context.Context, cfg config.Config, log *slog.Logger) {
	log.InfoContext(ctx, "not yet implemented", "command", "scrape")
}
