# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Summary

In-Network-Explorer is an AI-powered LinkedIn prospecting engine for IT professionals in Berlin. It scrapes LinkedIn profiles, scores them via LLM, and drafts personalized connection messages. The system prepares data; the human performs final outreach.

## Tech Stack

- **Language:** Go 1.25+
- **Browser Automation:** go-rod/rod (headless Chromium)
- **Infrastructure (v1):** AWS EC2 + cron jobs, DynamoDB, Secrets Manager
- **Infrastructure (future):** AWS Lambda (Docker/ECR), EventBridge
- **AI:** Amazon Bedrock (Claude 3.5 Sonnet) or OpenAI API

## Build & Run

```bash
go build ./...
go test ./...
go test -run TestName ./path/to/package   # single test
go test -race ./...                        # race condition detection
go vet ./...                               # static analysis
```

## Architecture (Planned)

The system has four phases that execute as a multi-day pipeline per prospect:

1. **Scraper & Parser** — Rod-based browser automation logs in via session cookies (from Secrets Manager), monitors Berlin IT keywords/hashtags, extracts profile data (URL, name, role, location, recent posts).

2. **AI Analysis Engine** — LLM scores prospects (1-10) against a target persona using location, field relevance, and activity. Drafts 150-300 char connection messages referencing specific post details.

3. **Persistence (DynamoDB)** — PK: LinkedIn Profile URL. Tracks status (`Scanned → Liked → Drafted → Sent`), worthiness score, drafted message, last action date.

4. **Human-in-the-Loop** — Generates a prospect report (JSON/HTML) every 48 hours with profile links, drafted messages, and copy-to-clipboard functionality.

## Anti-Detection Design

All browser interactions must include randomized jitter (20-40% variance on delays). The multi-step warming sequence is critical:
- Day 1: Visit profile only
- Day 2: Like a post or wait
- Day 3: Propose message to user

## Deployment

**V1 (current target):** EC2 instance running the Go binary + Chromium as a long-running process, orchestrated by system cron jobs. Design the application as a CLI with subcommands (e.g., `scrape`, `analyze`, `report`) so cron can invoke each phase independently.

**Future:** Migrate to Lambda (Docker/ECR, 2048MB RAM, 1-min timeout) + EventBridge scheduling. Keep this migration path in mind — isolate infrastructure concerns behind interfaces so the swap is minimal.

IAM roles for DynamoDB and Bedrock access.

---

## Skills & Working Guidelines

### 1. Code Commits

- Follow [Conventional Commits](https://www.conventionalcommits.org/): `type(scope): description` (e.g., `feat(scraper): add profile visit jitter logic`).
- Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `ci`, `perf`.
- Commit messages are short (50 chars subject), imperative mood, lowercase.
- Body (when needed) explains **why**, not what. Keep it to 2-3 lines max.
- Never reference AI tooling or assistants in commit messages.
- One logical change per commit. Don't bundle unrelated changes.

### 2. Go Engineering

**Philosophy:** Prefer Go stdlib over third-party libraries. When a small, well-maintained library significantly improves readability or delivery speed without bloating the binary, use it — but justify the dependency. The balance is: native Go for performance-critical paths and infrastructure glue; libraries for complex domains (e.g., rod for browser automation, AWS SDK).

**Quality attributes (in priority order):**
1. **Security** — sanitize all external input, never log secrets, use env vars for config.
2. **Maintainability** — clean architecture with layers (domain, usecase, adapter, infrastructure). Idiomatic Go: small interfaces, explicit error handling, no magic.
3. **Performance** — efficient where it matters (connection pooling, minimal allocations in hot paths), but never at the cost of readability.
4. **Scalability** — stateless design, easy to horizontally scale from EC2 to Lambda.
5. **Availability** — graceful shutdown, retries with backoff for external calls.

**Patterns:**
- **Clean Architecture layers:** `domain/` (entities, value objects), `usecase/` (business logic), `adapter/` (HTTP, CLI, DB implementations), `infrastructure/` (AWS clients, config).
- **DDD when warranted:** Use aggregates and domain events for the prospect lifecycle (Scanned → Liked → Drafted → Sent). Don't force DDD on simple CRUD.
- **TDD always:** Write the test first. Red → Green → Refactor. Table-driven tests. Use interfaces for test doubles — no mocking frameworks unless truly necessary.
- **Configuration:** All config via environment variables. Use a single config struct loaded at startup. No hardcoded values.
- **Error handling:** Wrap errors with `fmt.Errorf("context: %w", err)`. Return errors, don't panic. Use sentinel errors for domain-level conditions.
- **Concurrency:** Explicit goroutine lifecycle management. Always pass `context.Context`. Use `errgroup` for parallel work.

### 3. Code Review

When reviewing code (PR or console):
- Check for **security issues**: injection, secret leakage, improper input validation, unsafe concurrency.
- Check for **race conditions**: shared state without synchronization, goroutine leaks, missing context cancellation.
- Check for **performance**: unnecessary allocations, N+1 queries, unbounded goroutines, missing timeouts on external calls.
- Check for **Go idioms**: error handling, naming conventions, interface satisfaction, package organization.
- Check for **test quality**: meaningful assertions, edge cases covered, no test interdependence.
- PR comments: attach to the specific line. Be precise — state the problem, the risk, and a suggested fix.
- Console review: structured output grouped by severity (critical / warning / suggestion).

### 4. Solution Architecture

Understand the infrastructure context before writing code. The deployment target shapes the design:

- **V1 (EC2 + cron):** The app is a long-lived binary. It can hold state in memory between phases, use local filesystem for temp storage, and run Chromium as a child process. Cron schedules each subcommand. No cold-start concerns. Use `os/signal` for graceful shutdown.
- **Future (Lambda):** Stateless, event-driven. Each invocation is isolated. Chromium must be bundled in the container image. DynamoDB becomes the only shared state. Keep execution under 1 min.

Design decisions must account for this migration path. Abstractions at the infrastructure boundary (repository interfaces, LLM client interfaces, browser automation interfaces) allow swapping the runtime without touching business logic.

### 5. Technical Documentation

- Method-level: Go doc comments on all exported types and functions. Short first sentence, then details if needed.
- Project-level: Markdown files in `docs/` directory.
- Focus on **business value and workflows**, not implementation details that are obvious from the code.
- Use Mermaid diagrams for: class diagrams (domain model), sequence diagrams (prospect lifecycle), and deployment architecture.
- Keep docs short and scannable. Update docs when the code they describe changes.

### 6. Dependency Management

- Keep `go.mod` minimal. Every dependency is a liability — audit before adding.
- Prefer stdlib: `net/http`, `encoding/json`, `database/sql`, `testing`, `context`, `log/slog`.
- Acceptable dependencies: `go-rod/rod` (browser), `aws-sdk-go-v2` (AWS), small focused libs with good maintenance.
- Run `go mod tidy` after any dependency change.
- Pin major versions. Review changelogs before upgrading.

### 7. Error Handling & Observability

- Use `log/slog` (structured logging) from stdlib. No third-party loggers.
- Log at boundaries: incoming requests, outgoing calls, errors, state transitions.
- Include correlation fields: prospect URL, phase, timestamp.
- Never log secrets, cookies, tokens, or full profile data.
- Errors bubble up with context wrapping. Log once at the top of the call stack, not at every level.

### 8. Security Hardening

- All secrets (LinkedIn cookies, API keys) loaded from env vars or AWS Secrets Manager. Never in code, config files, or logs.
- Validate and sanitize all data extracted from LinkedIn before storing or passing to LLM.
- Rate-limit outbound requests. Respect anti-detection jitter requirements.
- Use `context.Context` with timeouts on all external calls (DynamoDB, Bedrock, LinkedIn).
- Audit for OWASP Top 10 where applicable (especially injection in LLM prompts).
