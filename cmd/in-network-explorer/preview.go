package main

import (
	"context"
	"log/slog"

	"github.com/pavlomaksymov/in-network-explorer/internal/report"
)

func runPreview(ctx context.Context, log *slog.Logger) {
	data := report.SeedReport()
	srv := report.NewPreviewServer(data, ":8085", log)
	if err := srv.ListenAndServe(ctx); err != nil {
		log.Error("preview server stopped", "err", err)
	}
}
