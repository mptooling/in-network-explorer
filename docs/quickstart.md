# Quickstart

Get the project running locally in under five minutes.

## Prerequisites

- Go 1.25+
- Chromium or Google Chrome installed
- AWS credentials configured (`aws configure` or environment variables)
- A DynamoDB table created in your target region

## 1. Clone and build

```bash
git clone https://github.com/pavlomaksymov/in-network-explorer.git
cd in-network-explorer
go build ./...
```

## 2. Configure environment

Copy the example env file and fill in your values:

```bash
cp .env.example .env
```

Edit `.env` with real values — at minimum the four required variables:

| Variable | Description |
|---|---|
| `AWS_REGION` | AWS region for DynamoDB and Bedrock |
| `DYNAMO_TABLE` | DynamoDB table name |
| `LINKEDIN_COOKIES_SECRET` | Secrets Manager ARN for LinkedIn session cookies |
| `CHROME_PROFILE_DIR` | Path to a Chrome user-data directory |

The application loads `.env` automatically at startup — no need to source it. Variables already set in your shell take precedence over values in the file.

## 3. Run tests

```bash
go test ./...
```

## 4. Run a subcommand

```bash
go run ./cmd scrape
go run ./cmd analyze
go run ./cmd report
go run ./cmd calibrate
```

> Subcommands are stubs for now — they will print "not yet implemented" and exit.

## Environment variable reference

See [`.env.example`](../.env.example) for the full list with defaults. Optional variables are commented out with their default values shown.
