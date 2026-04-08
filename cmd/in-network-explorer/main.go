package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pavlomaksymov/in-network-explorer/infrastructure"
)

// newLogger creates a structured JSON logger pre-configured with the given
// run_id field. Every log line emitted through this logger (or loggers
// derived from it via With) includes the run_id for correlation.
func newLogger(w io.Writer, runID string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(w, nil)).With("run_id", runID)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: explorer [scrape|analyze|report|calibrate]")
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "scrape", "analyze", "report", "calibrate":
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(1)
	}

	cfg := infrastructure.MustLoad()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log := newLogger(os.Stdout, rand.Text())

	switch cmd {
	case "scrape":
		runScrape(ctx, cfg, log)
	case "analyze":
		runAnalyze(ctx, cfg, log)
	case "report":
		runReport(ctx, cfg, log)
	case "calibrate":
		runCalibrate(ctx, cfg, log)
	}
}
