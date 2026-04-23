// Package config provides application configuration loading and AWS client
// construction.
package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	// AWS
	AWSRegion       string // AWS_REGION, required
	DynamoTableName string // DYNAMO_TABLE, required
	BedrockModelID  string // BEDROCK_MODEL_ID, default: anthropic.claude-haiku-4-5-20251001-v1:0
	BedrockRegion   string // BEDROCK_REGION, default: AWSRegion

	// LinkedIn
	LinkedInCookiesSecret string // LINKEDIN_COOKIES_SECRET, required (Secrets Manager ARN)

	// Local development overrides (optional, empty = use real AWS)
	DynamoEndpoint  string // DYNAMO_ENDPOINT, optional (e.g., http://localhost:8000)
	SecretsEndpoint string // SECRETS_ENDPOINT, optional (e.g., http://localhost:4566)

	// Proxy
	ProxyAddr string // PROXY_ADDR, optional (host:port)
	ProxyUser string // PROXY_USER, optional
	ProxyPass string // PROXY_PASS, optional

	// Chrome
	ChromeBin        string // CHROME_BIN, optional (auto-detect if empty)
	ChromeProfileDir string // CHROME_PROFILE_DIR, required

	// Qdrant
	QdrantAddr       string // QDRANT_ADDR, default: localhost:6334
	QdrantCollection string // QDRANT_COLLECTION, default: prospects

	// Search
	SearchQuery    string // SEARCH_QUERY, optional (e.g., "software engineer")
	SearchLocation string // SEARCH_LOCATION, optional (e.g., "Berlin")

	// Operational
	MaxProfileViewsPerDay   int    // MAX_PROFILE_VIEWS_PER_DAY, default: 40
	MaxConnectionReqsPerDay int    // MAX_CONNECTION_REQS_PER_DAY, default: 10
	MaxProspectsPerRun      int    // MAX_PROSPECTS_PER_RUN, default: 20
	AnalyzeConcurrency      int    // ANALYZE_CONCURRENCY, default: 3
	MaxNavsBeforeRestart    int    // BROWSER_MAX_NAVS, default: 50
	PromptConfigPath        string // PROMPT_CONFIG_PATH, default: prompts/scoring.json
	ReportOutputDir         string // REPORT_OUTPUT_DIR, default: reports
}

// MustLoad loads a .env file (if present) and reads all configuration from
// environment variables. It panics with a descriptive message if any required
// variable is missing. Already-set environment variables take precedence over
// values in the .env file.
func MustLoad() Config {
	loadDotEnv(".env")

	cfg := Config{
		AWSRegion:             required("AWS_REGION"),
		DynamoTableName:       required("DYNAMO_TABLE"),
		LinkedInCookiesSecret: required("LINKEDIN_COOKIES_SECRET"),
		ChromeProfileDir:      required("CHROME_PROFILE_DIR"),

		BedrockModelID: envOrDefault("BEDROCK_MODEL_ID", "anthropic.claude-haiku-4-5-20251001-v1:0"),

		DynamoEndpoint:  os.Getenv("DYNAMO_ENDPOINT"),
		SecretsEndpoint: os.Getenv("SECRETS_ENDPOINT"),

		ProxyAddr: os.Getenv("PROXY_ADDR"),
		ProxyUser: os.Getenv("PROXY_USER"),
		ProxyPass: os.Getenv("PROXY_PASS"),
		ChromeBin: os.Getenv("CHROME_BIN"),

		QdrantAddr:       envOrDefault("QDRANT_ADDR", "localhost:6334"),
		QdrantCollection: envOrDefault("QDRANT_COLLECTION", "prospects"),

		SearchQuery:    os.Getenv("SEARCH_QUERY"),
		SearchLocation: os.Getenv("SEARCH_LOCATION"),

		MaxProfileViewsPerDay:   envOrDefaultInt("MAX_PROFILE_VIEWS_PER_DAY", 40),
		MaxConnectionReqsPerDay: envOrDefaultInt("MAX_CONNECTION_REQS_PER_DAY", 10),
		MaxProspectsPerRun:      envOrDefaultInt("MAX_PROSPECTS_PER_RUN", 20),
		AnalyzeConcurrency:      envOrDefaultInt("ANALYZE_CONCURRENCY", 3),
		MaxNavsBeforeRestart:    envOrDefaultInt("BROWSER_MAX_NAVS", 50),
		PromptConfigPath:        envOrDefault("PROMPT_CONFIG_PATH", "prompts/scoring.json"),
		ReportOutputDir:         envOrDefault("REPORT_OUTPUT_DIR", "reports"),
	}

	cfg.BedrockRegion = envOrDefault("BEDROCK_REGION", cfg.AWSRegion)

	return cfg
}

func required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return v
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("environment variable %s must be an integer, got %q", key, v))
	}
	return n
}

// loadDotEnv reads a KEY=VALUE file and sets each pair in the process
// environment. Variables already set in the environment are not overwritten,
// so real env vars always win. It panics if the file cannot be opened.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if _, exists := os.LookupEnv(key); exists {
			continue // never override existing env
		}
		_ = os.Setenv(key, value)
	}
}
