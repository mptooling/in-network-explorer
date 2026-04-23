package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/config"
	"github.com/pavlomaksymov/in-network-explorer/internal/report"
)

func runReport(ctx context.Context, cfg config.Config, log *slog.Logger) {
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

	if err := writeReportFiles(cfg.ReportOutputDir, entries, log); err != nil {
		log.ErrorContext(ctx, "write report files", "error", err)
	}
}

func writeReportFiles(dir string, entries []explorer.ProspectReport, log *slog.Logger) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create report dir: %w", err)
	}

	stamp := time.Now().UTC().Format("2006-01-02_1504")
	r := report.New(entries)

	jsonPath := filepath.Join(dir, fmt.Sprintf("report_%s.json", stamp))
	if err := writeFile(jsonPath, r.WriteJSON); err != nil {
		return err
	}

	htmlPath := filepath.Join(dir, fmt.Sprintf("report_%s.html", stamp))
	if err := writeFile(htmlPath, r.WriteHTML); err != nil {
		return err
	}

	log.Info("report written", "json", jsonPath, "html", htmlPath, "prospects", len(entries))
	return nil
}

func writeFile(path string, writeFn func(io.Writer) error) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()
	return writeFn(f)
}
