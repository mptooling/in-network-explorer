# In-Network-Explorer

AI-powered LinkedIn prospecting engine for IT professionals in Berlin. Scrapes profiles, scores them via LLM, and drafts personalized connection messages. The system prepares the data; the human performs the final outreach.

## Quickstart

```bash
# Clone and install deps
git clone git@github.com:mptooling/in-network-explorer.git
cd in-network-explorer
go mod download

# Create your .env from the example
cp .env.example .env

# Start local infrastructure (DynamoDB, LocalStack, Qdrant)
make dev
```

## Build

```bash
make build
```

## CLI commands

```bash
# Discover prospects and warm up (requires Chromium + LinkedIn cookies)
./in-network-explorer scrape

# LLM-score liked prospects and draft messages (requires AWS Bedrock)
./in-network-explorer analyze

# Generate JSON + HTML report to reports/ directory
./in-network-explorer report

# Serve live report on http://localhost:8080
./in-network-explorer preview

# Re-score sample to check LLM consistency
./in-network-explorer calibrate
```

## Testing

```bash
# Unit tests (no infrastructure needed)
make test

# Integration tests (requires make dev running)
make test-integration

# Lint
make lint
```

## Manual end-to-end test (no real LinkedIn)

```bash
# 1. Start local infra
make dev

# 2. Seed a fake prospect into DynamoDB
aws dynamodb put-item \
  --endpoint-url http://localhost:8881 \
  --table-name prospects-dev \
  --item '{
    "PK":{"S":"https://linkedin.com/in/test-user"},
    "Slug":{"S":"test-user"},
    "Name":{"S":"Test User"},
    "Headline":{"S":"Staff Engineer"},
    "Location":{"S":"Berlin"},
    "State":{"S":"DRAFTED"},
    "WorthinessScore":{"N":"8"},
    "ScoreReasoning":{"S":"Strong Berlin tech match"},
    "DraftedMessage":{"S":"Hi Test, your work on observability is impressive!"},
    "CritiqueScore":{"N":"12"}
  }'

# 3. Generate and preview the report
./in-network-explorer report
./in-network-explorer preview
# Open http://localhost:8080

# 4. Teardown
make dev-down
```

## Architecture

The system runs as a multi-day pipeline per prospect:

1. **Scrape** — Rod-based browser automation discovers profiles, extracts data, saves as `Scanned`.
2. **Warm up** — Next scrape run likes a post on due prospects, advancing to `Liked`.
3. **Analyze** — LLM scores prospects (1-10), drafts connection messages, transitions to `Drafted`.
4. **Report** — Generates HTML/JSON with profile links, scores, and copy-to-clipboard messages.

Anti-detection: randomized jitter on all interactions, multi-day warming sequence, stealth browser patches, extension fingerprint spoofing.

## Tech stack

- **Language:** Go 1.25+
- **Browser:** go-rod/rod (headless Chromium) + go-rod/stealth
- **Infrastructure:** AWS (DynamoDB, Bedrock, Secrets Manager)
- **Vector search:** Qdrant
- **Deployment (v1):** EC2 + cron

## Project structure

```
cmd/in-network-explorer/   CLI entry point and subcommands
internal/                   Domain entities, interfaces, use cases (package explorer)
internal/config/            Config loading, AWS clients, browser factory
internal/dynamo/            DynamoDB prospect repo + rate limiter
internal/linkedin/          Rod-based LinkedIn browser client
internal/bedrock/           Bedrock LLM + embedding clients
internal/qdrant/            Qdrant vector store (REST API)
internal/report/            HTML/JSON rendering + preview server
internal/jitter/            Human behavior simulation (timing, scroll, typing, mouse)
internal/testdouble/        In-memory fakes for all interfaces
```
