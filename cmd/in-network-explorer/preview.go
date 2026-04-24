package main

import (
	"context"
	"log/slog"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/config"
	"github.com/pavlomaksymov/in-network-explorer/internal/report"
)

func runPreview(ctx context.Context, cfg config.Config, log *slog.Logger) {
	repo, _, err := buildDynamoDeps(ctx, cfg, log)
	if err != nil {
		return
	}

	uc := explorer.NewReportUseCase(repo, log, cfg.MaxProspectsPerRun)
	entries, err := uc.Run(ctx)
	if err != nil {
		log.ErrorContext(ctx, "report generation failed", "error", err)
		return
	}

	r := report.New(entries)
	addr := ":8080"
	log.InfoContext(ctx, "preview server", "url", report.Addr(addr), "prospects", len(entries))

	srv := report.NewServer(r, log)
	if err := srv.ListenAndServe(addr); err != nil {
		log.ErrorContext(ctx, "server stopped", "error", err)
	}
}
