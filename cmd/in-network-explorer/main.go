package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/pavlomaksymov/in-network-explorer/infrastructure"
)

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

	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	switch cmd {
	case "scrape":
		runScrape(ctx, cfg)
	case "analyze":
		runAnalyze(ctx, cfg)
	case "report":
		runReport(ctx, cfg)
	case "calibrate":
		runCalibrate(ctx, cfg)
	}
}
