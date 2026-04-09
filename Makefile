.DEFAULT_GOAL := help

.PHONY: help dev dev-down dev-logs test test-integration lint build clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

dev: ## Start local development infrastructure (DynamoDB, LocalStack, Qdrant)
	docker compose up -d --wait
	@echo ""
	@echo "Local services running:"
	@echo "  DynamoDB Local:  http://localhost:8881"
	@echo "  LocalStack:      http://localhost:4566"
	@echo "  Qdrant REST:     http://localhost:6333"
	@echo "  Qdrant gRPC:     localhost:6334"
	@echo ""
	@echo "Copy .env.local to .env if you haven't, then run: air"

dev-down: ## Stop local infrastructure and remove volumes
	docker compose down -v

dev-logs: ## Tail logs from Docker services
	docker compose logs -f

test: ## Run unit tests with race detector
	go test -race -count=1 ./...

test-integration: ## Run integration tests (requires local services or real AWS)
	go test -tags integration -race -count=1 -v ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

build: ## Build the binary
	go build -o ./in-network-explorer ./cmd/in-network-explorer

clean: ## Remove build artifacts
	rm -f ./in-network-explorer
	rm -rf tmp/
