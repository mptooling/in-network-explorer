package main

import (
	"context"
	"log/slog"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/config"
	"github.com/pavlomaksymov/in-network-explorer/internal/dynamo"
)

func runScrape(ctx context.Context, cfg config.Config, log *slog.Logger) {
	if cfg.SearchQuery == "" || cfg.SearchLocation == "" {
		log.ErrorContext(ctx, "SEARCH_QUERY and SEARCH_LOCATION must be set")
		return
	}

	repo, limiter, err := buildDynamoDeps(ctx, cfg, log)
	if err != nil {
		return
	}

	// BrowserClient adapter is not yet implemented. When internal/linkedin is
	// ready, construct it here and pass to NewScrapeUseCase.
	log.ErrorContext(ctx, "browser adapter not yet implemented — scrape requires internal/linkedin")
	_ = explorer.NewScrapeUseCase(repo, nil, limiter, log, cfg.MaxProspectsPerRun)
}

// buildDynamoDeps creates the DynamoDB prospect repository and rate limiter
// shared by multiple commands.
func buildDynamoDeps(ctx context.Context, cfg config.Config, log *slog.Logger) (*dynamo.ProspectRepository, *dynamo.RateLimiter, error) {
	client, err := config.NewDynamoClient(ctx, cfg)
	if err != nil {
		log.ErrorContext(ctx, "dynamodb client", "error", err)
		return nil, nil, err
	}
	repo := dynamo.NewProspectRepository(client, cfg.DynamoTableName)
	limiter := dynamo.NewRateLimiter(client, cfg.DynamoTableName, map[string]int{
		"profile_views":       cfg.MaxProfileViewsPerDay,
		"connection_requests": cfg.MaxConnectionReqsPerDay,
	})
	return repo, limiter, nil
}
