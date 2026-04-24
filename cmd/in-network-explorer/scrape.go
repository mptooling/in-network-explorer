package main

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/config"
	"github.com/pavlomaksymov/in-network-explorer/internal/dynamo"
	"github.com/pavlomaksymov/in-network-explorer/internal/linkedin"
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

	browser, browserCleanup, err := config.NewBrowser(ctx, cfg)
	if err != nil {
		log.ErrorContext(ctx, "launch browser", "error", err)
		return
	}
	defer browserCleanup()

	cookies, err := loadCookies(ctx, cfg)
	if err != nil {
		log.ErrorContext(ctx, "load cookies", "error", err)
		return
	}

	client, err := linkedin.New(browser, cookies, log)
	if err != nil {
		log.ErrorContext(ctx, "create linkedin client", "error", err)
		return
	}
	defer func() { _ = client.Close() }()

	uc := explorer.NewScrapeUseCase(repo, client, limiter, log, cfg.MaxProspectsPerRun)
	if err := uc.Run(ctx, cfg.SearchQuery, cfg.SearchLocation); err != nil {
		log.ErrorContext(ctx, "scrape failed", "error", err)
	}
}

func loadCookies(ctx context.Context, cfg config.Config) (string, error) {
	sc, err := config.NewSecretsClient(ctx, cfg)
	if err != nil {
		return "", err
	}
	out, err := sc.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(cfg.LinkedInCookiesSecret),
	})
	if err != nil {
		return "", err
	}
	return *out.SecretString, nil
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
