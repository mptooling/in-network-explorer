# In-Network-Explorer: Implementation Plan

**Approach:** Agile — Epics → Stories → Sub-tasks. One Epic = one branch = one PR (squash-merged to `main`). Each sub-task = one commit. All sub-tasks follow TDD: write the failing test first, then implement, then refactor. Each Claude session handles one Epic.

---

## Conventions

### Branch Naming
```
feat/stage-{n}-{slug}
```
Example: `feat/stage-1-foundation`

### Commit Format (Conventional Commits)
```
type(scope): short imperative description (≤50 chars)

[optional body: why, not what — 2-3 lines max]
```
Types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `ci`, `perf`

### PR & Merge Strategy
- One PR per Epic
- PR title = squash commit message (e.g., `feat(foundation): project skeleton, config, domain, CLI`)
- Squash-merge into `main` — linear history, one commit per Epic
- No force-push to `main`

### TDD Protocol (per sub-task)
1. **Red** — write the test file, run `go test ./...`, confirm it fails
2. **Green** — write minimal implementation to pass
3. **Refactor** — clean up without breaking tests
4. Commit only when tests are green

### Test Style
- Table-driven tests for all pure logic
- Interface-based test doubles (no mocking frameworks)
- Test doubles live in `internal/testdouble/` (shared fakes)
- Integration tests marked with `//go:build integration` build tag — not run in CI by default

### Definition of Done (per sub-task)
- [ ] Tests written first and passing (`go test ./...`)
- [ ] `go vet ./...` passes
- [ ] No new linter warnings
- [ ] Commit message follows Conventional Commits

### Definition of Done (per Epic)
- [ ] All sub-tasks committed
- [ ] `go test -race ./...` passes
- [ ] PR squash-merged to `main`
- [ ] No commented-out code, no TODOs without tracking issues

---

## Directory Layout

```
in-network-explorer/
├── .github/
│   └── workflows/
│       ├── ci.yml               # unit tests, vet, build, lint — runs on every push + PR
│       └── integration.yml      # integration tests — manual trigger only
├── cmd/
│   └── in-network-explorer/
│       ├── main.go              # CLI root: os.Args dispatch, DI wiring, signal handling
│       ├── scrape.go            # scrape subcommand
│       ├── analyze.go           # analyze subcommand
│       ├── report.go            # report subcommand
│       └── calibrate.go         # calibrate subcommand
├── internal/                    # package explorer — flat Ben Johnson layout
│   ├── prospect.go              # Prospect struct, State enum, Transition()
│   ├── errors.go                # Sentinel errors
│   ├── repository.go            # ProspectRepository, RateLimiter interfaces
│   ├── llm.go                   # LLMClient interface + ScoreResult type
│   ├── embedding.go             # EmbeddingClient, EmbeddingStore interfaces
│   ├── browser.go               # BrowserClient interface + ProfileData, BlockType
│   ├── scrape.go                # ScrapePhase use case
│   ├── warmup.go                # WarmupPhase use case
│   ├── analyze.go               # AnalyzePhase use case
│   ├── report.go                # ReportPhase use case
│   ├── calibrate.go             # CalibratePhase use case
│   ├── config/
│   │   ├── config.go            # Config struct, MustLoad()
│   │   ├── aws.go               # AWS client constructors
│   │   └── browser.go           # go-rod browser factory
│   ├── dynamo/
│   │   ├── prospect_repo.go     # ProspectRepository implementation
│   │   └── rate_limiter.go      # RateLimiter implementation
│   ├── qdrant/
│   │   └── embedding_store.go   # EmbeddingStore implementation
│   ├── bedrock/
│   │   ├── llm_client.go        # LLMClient implementation
│   │   └── embedding_client.go  # EmbeddingClient implementation
│   ├── linkedin/
│   │   ├── browser_client.go    # BrowserClient implementation
│   │   └── voyager_client.go    # Voyager API client
│   ├── jitter/
│   │   └── jitter.go            # timing distributions, mouse/scroll/typing simulation
│   ├── report/
│   │   └── render.go            # JSON/HTML report renderer
│   └── testdouble/
│       └── fakes.go             # shared in-memory fakes for tests
└── docs/
    └── architecture.md
```

---

## Stage Overview

| # | Epic | Branch | Stories | Commits |
|---|---|---|---|---|
| 1 | Foundation | `feat/stage-1-foundation` | 5 | ~15 |
| 2 | Browser Core | `feat/stage-2-browser-core` | 4 | ~14 |
| 3 | Data Extraction | `feat/stage-3-data-extraction` | 4 | ~13 |
| 4 | Scrape Phase | `feat/stage-4-scrape` | 3 | ~10 |
| 5 | Warm-Up Pipeline | `feat/stage-5-warmup` | 3 | ~9 |
| 6 | AI Analysis | `feat/stage-6-analysis` | 4 | ~13 |
| 7 | RAG & Embeddings | `feat/stage-7-rag` | 3 | ~10 |
| 8 | Calibration | `feat/stage-8-calibration` | 3 | ~9 |
| 9 | Report | `feat/stage-9-report` | 3 | ~8 |
| 10 | Production Hardening | `feat/stage-10-hardening` | 3 | ~9 |

---

## Stage 1 — Foundation

**Branch:** `feat/stage-1-foundation`  
**PR title:** `feat(foundation): project skeleton, CI pipelines, domain model, DynamoDB schema, CLI routing`  
**Goal:** A runnable binary that does nothing useful yet, but has the full domain model, correct DynamoDB schema, working CLI routing, and GitHub Actions CI validating every subsequent commit automatically. Everything a future Claude session can build on without rework.

---

### Story 1.1 — Module Init & Project Layout

**Goal:** Initialise the Go module and create the directory skeleton with placeholder files.

#### Sub-task 1.1.1 — Initialise Go module and create directory skeleton
**Test:** No automated test. Verify manually: `go build ./...` succeeds with no source errors after adding a `package main` placeholder.

**Implementation:**
```bash
go mod init github.com/pavlomaksymov/in-network-explorer
# Create all directories listed in the layout above (empty, with .gitkeep)
```

`go.mod` must declare `go 1.25` minimum. Add no dependencies yet.

**Commit:** `chore(init): initialise go module and create directory skeleton`

---

#### Sub-task 1.1.2 — Add core dependencies to go.mod
**Test:** `go mod tidy` and `go build ./...` pass with no errors.

**Dependencies to add:**
```bash
go get github.com/go-rod/rod
go get github.com/go-rod/stealth
go get github.com/aws/aws-sdk-go-v2/config
go get github.com/aws/aws-sdk-go-v2/service/dynamodb
go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression
go get github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue
go get github.com/aws/aws-sdk-go-v2/service/secretsmanager
go get github.com/aws/aws-sdk-go-v2/service/bedrockruntime
go get github.com/qdrant/go-client
go get github.com/PuerkitoBio/goquery
go get golang.org/x/sync
go mod tidy
```

**Commit:** `chore(deps): add core dependencies`

---

#### Sub-task 1.1.3 — GitHub Actions CI workflow (unit tests + lint)
**Test:** Push this commit to the remote branch and confirm the workflow appears green in the GitHub Actions tab. Verify that a deliberate `go vet` failure (e.g., `fmt.Println(` with unclosed paren) causes the CI job to fail.

**Implementation:** `.github/workflows/ci.yml`

```yaml
name: CI

on:
  push:
    branches: ["**"]
  pull_request:
    branches: [main]

permissions:
  contents: read

jobs:
  test:
    name: Test & Vet
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true            # caches module download + build cache

      - name: Download modules
        run: go mod download

      - name: Vet
        run: go vet ./...

      - name: Build
        run: go build ./...

      - name: Test (race detector)
        run: go test -race -count=1 ./...

  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=5m
```

**Implementation:** `.github/workflows/integration.yml`

```yaml
name: Integration Tests

on:
  workflow_dispatch:         # manual trigger only — requires real AWS + LinkedIn session
    inputs:
      reason:
        description: "Why are you running integration tests?"
        required: true

permissions:
  contents: read

jobs:
  integration:
    name: Integration
    runs-on: ubuntu-latest
    environment: integration   # GitHub environment with protected secrets
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Run integration tests
        env:
          AWS_REGION:            ${{ secrets.AWS_REGION }}
          DYNAMO_TABLE:          ${{ secrets.DYNAMO_TABLE }}
          LINKEDIN_COOKIES_SECRET: ${{ secrets.LINKEDIN_COOKIES_SECRET }}
          QDRANT_ADDR:           ${{ secrets.QDRANT_ADDR }}
        run: go test -tags integration -race -count=1 -v ./...
```

**Implementation:** `.golangci.yml` (linter configuration at repo root)

```yaml
linters:
  enable:
    - errcheck        # no unchecked errors
    - gosimple        # simplification suggestions
    - govet           # go vet checks
    - ineffassign     # detect ineffective assignments
    - staticcheck     # SA* checks (misuse of sync, context, etc.)
    - unused          # detect unused code
    - gofmt           # formatting
    - goimports       # import ordering
    - godot           # comments must end with a period
    - misspell        # English spelling
    - noctx           # HTTP requests must accept context

linters-settings:
  goimports:
    local-prefixes: github.com/pavlomaksymov/in-network-explorer

issues:
  exclude-rules:
    - path: "_test\\.go"
      linters: [errcheck]   # unchecked errors in tests are acceptable
```

**Notes:**
- `go-version-file: go.mod` pins Go version from the module file — no drift between CI and local
- `cache: true` on `setup-go` caches both the module download cache and the build cache, making subsequent runs fast
- The `lint` job runs in parallel with `test` — both must pass for PR merge
- Integration tests run only on manual dispatch with a protected `integration` GitHub environment, keeping secrets isolated from the PR flow
- Branch protection rule to configure on GitHub: **require `test` and `lint` jobs to pass before merge**

**Commit:** `ci: add GitHub Actions workflows for unit tests, lint, and integration tests`

---

### Story 1.5 — Branch Protection & Merge Rules

**Goal:** Enforce the PR/squash-merge strategy via GitHub branch protection so no Claude session can accidentally skip CI.

#### Sub-task 1.5.1 — Document branch protection settings
**Test:** No automated test. Verify manually in GitHub → Settings → Branches.

**Implementation:** `docs/contributing.md`

Document the exact GitHub branch protection settings to configure manually (cannot be committed as code without the GitHub API or Terraform):

```markdown
# Branch Protection: main

Required status checks (must pass before merge):
  - CI / Test & Vet
  - CI / Lint

Settings:
  - Require a pull request before merging: YES
  - Require status checks to pass before merging: YES
  - Require branches to be up to date before merging: YES
  - Squash merge: ALLOWED (set as default merge strategy)
  - Merge commits: DISALLOWED
  - Rebase merges: DISALLOWED
  - Allow force pushes: NO
  - Allow deletions: NO
```

**Commit:** `docs(contributing): document branch protection and merge strategy`

---

### Story 1.2 — Domain Model

**Goal:** The `Prospect` aggregate with full state machine, sentinel errors, and all domain interfaces. This is the most important story — every other layer depends on it.

#### Sub-task 1.2.1 — Sentinel errors
**Test file:** `internal/errors_test.go`

```go
// Tests:
// - ErrNotFound is a non-nil error
// - ErrInvalidTransition unwraps correctly via errors.Is
// - ErrRateLimitExceeded is distinct from ErrInvalidTransition
// - errors.As works for typed wrapping
```

**Implementation:** `internal/errors.go`

```go
var (
    ErrNotFound          = errors.New("prospect not found")
    ErrInvalidTransition = errors.New("invalid state transition")
    ErrRateLimitExceeded = errors.New("daily rate limit exceeded")
    ErrBlockDetected     = errors.New("linkedin block detected")
    ErrSessionExpired    = errors.New("linkedin session expired")
)
```

**Commit:** `feat(domain): add sentinel errors`

---

#### Sub-task 1.2.2 — Prospect state machine
**Test file:** `internal/prospect_test.go`

```go
// Table-driven tests covering:
// TestProspect_Transition_Valid:
//   - Scanned → Liked: allowed
//   - Scanned → Skipped: allowed
//   - Liked → Drafted: allowed
//   - Drafted → Sent: allowed
//   - Sent → Accepted: allowed
//   - Sent → Rejected: allowed
// TestProspect_Transition_Invalid:
//   - Scanned → Drafted: returns ErrInvalidTransition
//   - Liked → Accepted: returns ErrInvalidTransition
//   - Accepted → Sent: returns ErrInvalidTransition (terminal state)
// TestProspect_Transition_UpdatesFields:
//   - LastActionAt changes after transition
//   - NextActionAt is in the future (≥20h) after StateScanned transition
//   - NextActionAt is in the future (≥20h) after StateLiked transition
//   - NextActionAt is zero-value for terminal states
// TestProspect_State_String:
//   - Each state produces the correct uppercase string
```

**Implementation:** `internal/prospect.go`

Full `Prospect` struct with all fields from the research (see Stage 1.2.3 for field list). `State` enum as `uint8`. `Transition()` method. `computeNextAction()` with 20-25h jitter using `math/rand`.

**Commit:** `feat(domain): add prospect aggregate with state machine`

---

#### Sub-task 1.2.3 — Domain interfaces
**Test file:** `internal/repository_test.go`, `internal/llm_test.go`, `internal/browser_test.go`

```go
// Tests (compile-time interface satisfaction only — no logic):
// - verify that a hand-written fake struct satisfies ProspectRepository
// - verify that a hand-written fake struct satisfies LLMClient
// - verify that a hand-written fake struct satisfies BrowserClient
// - verify that a hand-written fake struct satisfies RateLimiter
// - verify that a hand-written fake struct satisfies EmbeddingClient
// - verify that a hand-written fake struct satisfies EmbeddingStore
//
// Pattern: var _ ProspectRepository = (*fakeRepo)(nil)
```

**Implementation:** Four interface files:

`internal/repository.go`:
```go
type ProspectRepository interface {
    Save(ctx context.Context, p *Prospect) error
    Get(ctx context.Context, profileURL string) (*Prospect, error)
    InsertIfNew(ctx context.Context, p *Prospect) (inserted bool, err error)
    ListByState(ctx context.Context, state State, dueBy time.Time) ([]*Prospect, error)
    ListByStateOrderedByScore(ctx context.Context, state State, limit int) ([]*Prospect, error)
}

type RateLimiter interface {
    Acquire(ctx context.Context, scope string) error  // returns ErrRateLimitExceeded if daily cap hit
    Current(ctx context.Context, scope string) (int, error)
}
```

`internal/llm.go`:
```go
type ScoreResult struct {
    Score      int    // 1-10
    Reasoning  string // ≤50 words
    Message    string // 150-300 chars
    CritiqueScore int // 3-15 (sum of specificity+relevance+tone)
}

type LLMClient interface {
    ScoreAndDraft(ctx context.Context, p *Prospect, examples []Prospect) (ScoreResult, error)
    Critique(ctx context.Context, message string) (int, error) // returns composite score 3-15
}
```

`internal/embedding.go`:
```go
type EmbeddingClient interface {
    Embed(ctx context.Context, text string) ([]float32, error)
}

type EmbeddingStore interface {
    Upsert(ctx context.Context, id string, vector []float32, payload map[string]any) error
    SearchSimilar(ctx context.Context, vector []float32, filter map[string]any, topK int) ([]*Prospect, error)
}
```

`internal/browser.go`:
```go
type BlockType int
const (
    BlockNone BlockType = iota
    BlockAuthwall
    BlockChallenge
    BlockSoftEmpty
)

type ProfileData struct {
    URL        string
    Name       string
    Headline   string
    Location   string
    About      string
    RecentPosts []string
    Slug       string
}

type BrowserClient interface {
    VisitProfile(ctx context.Context, profileURL string) (ProfileData, error)
    LikeRecentPost(ctx context.Context, profileURL string) error
    SearchProfiles(ctx context.Context, query, location string, limit int) ([]string, error) // returns profile URLs
    CheckBlock(ctx context.Context) (BlockType, error)
    Close() error
}
```

**Commit:** `feat(domain): add repository, llm, browser, and embedding interfaces`

---

#### Sub-task 1.2.4 — Shared test doubles (in-memory fakes)
**Test file:** `internal/testdouble/fakes_test.go`

```go
// Tests for FakeProspectRepository:
// - Save then Get returns same prospect
// - InsertIfNew returns true on first call, false on duplicate
// - ListByState returns only prospects matching state and dueBy
// - ListByStateOrderedByScore returns prospects sorted by score descending
// Tests for FakeRateLimiter:
// - Acquire succeeds until max is reached
// - Acquire returns ErrRateLimitExceeded at max
// - Current returns correct count
```

**Implementation:** `internal/testdouble/fakes.go`

In-memory implementations of all domain interfaces. Used by usecase tests in every subsequent stage. Never used in production code.

**Commit:** `test(testdouble): add shared in-memory fakes for all domain interfaces`

---

### Story 1.3 — Configuration

**Goal:** All configuration loaded from environment variables at startup via a single `Config` struct. The binary fails fast with a descriptive error if required vars are missing.

#### Sub-task 1.3.1 — Config struct and loader
**Test file:** `internal/config/config_test.go`

```go
// Tests:
// TestMustLoad_AllPresent: sets all required env vars, MustLoad() succeeds
// TestMustLoad_MissingRequired: each required var absent → panic with var name in message
// TestMustLoad_DefaultValues: optional vars have correct defaults when absent
// TestConfig_ChromeBinDefault: empty CHROME_BIN is valid (auto-detect)
```

**Implementation:** `internal/config/config.go`

```go
type Config struct {
    // AWS
    AWSRegion       string // AWS_REGION, required
    DynamoTableName string // DYNAMO_TABLE, required
    BedrockModelID  string // BEDROCK_MODEL_ID, default: anthropic.claude-haiku-4-5-20251001-v1:0
    BedrockRegion   string // BEDROCK_REGION, default: AWSRegion

    // LinkedIn
    LinkedInCookiesSecret string // LINKEDIN_COOKIES_SECRET, required (Secrets Manager ARN)

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

    // Operational
    MaxProfileViewsPerDay     int // MAX_PROFILE_VIEWS_PER_DAY, default: 40
    MaxConnectionReqsPerDay   int // MAX_CONNECTION_REQS_PER_DAY, default: 10
    MaxProspectsPerRun        int // MAX_PROSPECTS_PER_RUN, default: 20
    AnalyzeConcurrency        int // ANALYZE_CONCURRENCY, default: 3
    PromptConfigPath          string // PROMPT_CONFIG_PATH, default: prompts/scoring.json
}
```

Use `os.Getenv` + explicit panic with variable name for required fields. No reflection magic.

**Commit:** `feat(infrastructure): add config struct loaded from environment variables`

---

### Story 1.4 — CLI & Graceful Shutdown

**Goal:** A working binary with four subcommands. Each subcommand currently just prints "not yet implemented". Signal handling shuts down cleanly.

#### Sub-task 1.4.1 — CLI router with os.Args dispatch
**Test file:** `cmd/in-network-explorer/main_test.go`

```go
// Tests (via exec.Command or os.Args injection):
// TestCLI_NoArgs: exits with code 1, prints usage to stderr
// TestCLI_UnknownCommand: exits with code 1, prints "unknown command: foo"
// TestCLI_KnownCommands: scrape/analyze/report/calibrate exit 0 with stub runners
```

**Implementation:** `cmd/in-network-explorer/main.go`

```go
func main() {
    cfg := infrastructure.MustLoad()
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    if len(os.Args) < 2 {
        fmt.Fprintln(os.Stderr, "usage: explorer [scrape|analyze|report|calibrate]")
        os.Exit(1)
    }
    log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    slog.SetDefault(log)

    switch os.Args[1] {
    case "scrape":    runScrape(ctx, cfg)
    case "analyze":   runAnalyze(ctx, cfg)
    case "report":    runReport(ctx, cfg)
    case "calibrate": runCalibrate(ctx, cfg)
    default:
        fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
        os.Exit(1)
    }
}
```

Each `run*` function in its own file, returning immediately with `slog.Info("not yet implemented")`.

**Commit:** `feat(cmd): add CLI router with os.Args dispatch and graceful shutdown`

---

#### Sub-task 1.4.2 — Structured logging setup
**Test file:** `cmd/in-network-explorer/logging_test.go`

```go
// Tests:
// TestLogging_JSONFormat: log output is valid JSON
// TestLogging_RunIDField: every log line includes run_id field
// TestLogging_CorrelationLogger: With("prospect_url", url) propagates to sub-calls
```

**Implementation:** Add `newLogger(runID string) *slog.Logger` to `cmd/in-network-explorer/main.go`. Every subcommand receives a pre-configured logger with `run_id` field injected. Logger is threaded via context using `slog.NewLogLogger` pattern — no global state.

**Commit:** `feat(cmd): add structured JSON logging with run_id correlation field`

---

#### Sub-task 1.4.3 — AWS client constructors
**Test file:** `internal/config/aws_test.go`

```go
// Tests (build-tag: integration):
// TestNewDynamoClient_Connect: creates client without error (requires AWS creds)
// TestNewBedrockClient_Connect: creates client without error
// Unit test (no AWS required):
// TestNewDynamoClient_UsesRegion: client configured with region from config
```

**Implementation:** `internal/config/aws.go`

```go
func NewDynamoClient(ctx context.Context, cfg Config) (*dynamodb.Client, error)
func NewBedrockClient(ctx context.Context, cfg Config) (*bedrockruntime.Client, error)
func LoadLinkedInCookies(ctx context.Context, cfg Config) (string, error) // reads from Secrets Manager
```

**Commit:** `feat(infrastructure): add AWS client constructors`

---

## Stage 2 — Browser Core

**Branch:** `feat/stage-2-browser-core`  
**PR title:** `feat(browser): human behavior simulation, stealth browser factory, session management`  
**Goal:** A fully-configured stealthy browser with realistic human mouse, scroll, and typing simulation. No LinkedIn-specific logic yet — just the reusable behavior layer.  
**Prerequisite:** Stage 1 merged to `main`.

---

### Story 2.1 — Timing & Jitter Primitives

**Goal:** Statistical sampling utilities that produce human-realistic timing across the entire system.

#### Sub-task 2.1.1 — Core timing distributions
**Test file:** `internal/jitter/jitter_test.go`

```go
// Tests:
// TestJitter_Range: Jitter(100ms) always returns value in [70ms, 130ms] over 1000 calls
// TestLogNormalSample_Positive: always returns positive value
// TestLogNormalSample_Distribution: mean and stddev within expected range over 10000 samples
// TestGammaSample_Positive: always returns positive value
// TestIKIDuration_CommonBigram: result is meaningfully shorter than rare pair (avg over 100 calls)
// TestIKIDuration_NeverZero: never returns 0 duration
```

**Implementation:** `internal/jitter/jitter.go`

```go
func Jitter(d time.Duration) time.Duration                          // ±30% uniform variance
func LogNormalDuration(mean time.Duration, sigma float64) time.Duration
func LogNormalSample(mean, sigma float64) float64
func GammaSample(shape, rate float64) float64                       // for inter-profile delays
func IKIDuration(baseWPM float64, isCommonBigram bool) time.Duration
func DwellDuration(wordCount int) time.Duration                     // profile reading time
```

**Commit:** `feat(jitter): add timing distribution utilities`

---

#### Sub-task 2.1.2 — WindMouse algorithm
**Test file:** `internal/jitter/windmouse_test.go`

```go
// Tests:
// TestWindMouseGuide_ReachesTarget: final point is within 2px of destination
// TestWindMouseGuide_ProducesMultiplePoints: guide function returns >5 points for 200px distance
// TestWindMouseGuide_VelocityClipped: no single step moves >20px (max velocity respected)
// TestWindMouseGuide_NonLinear: path deviates from straight line by >5px on average
// TestWindMouseGuide_Terminates: guide eventually returns done=true (no infinite loop)
```

**Implementation:** `internal/jitter/windmouse.go`

Port WindMouse algorithm (gravity + stochastic wind) to Go as a `func() (proto.Point, bool)` guide function compatible with `rod.Mouse.MoveAlong`.

Parameters: `G0=9, W0=3, M0=15, D0=12`. All as unexported constants.

**Commit:** `feat(jitter): add WindMouse algorithm for human-like mouse movement`

---

#### Sub-task 2.1.3 — Bezier curve mouse movement
**Test file:** `internal/jitter/bezier_test.go`

```go
// Tests:
// TestBezierGuide_StartsAtStart: first point equals start
// TestBezierGuide_EndsAtEnd: last point equals end (within 1px)
// TestBezierGuide_CorrectStepCount: produces steps = int(distance/3) points, clamped [10,80]
// TestBezierGuide_OneSideOfLine: all control points on same side of start-end line
// TestBezierGuide_SmoothVelocity: velocity (point-to-point distance) bell-curves: peaks in middle
```

**Implementation:** `internal/jitter/bezier.go`

Cubic Bézier with control points placed on one side of the start-end line. Steps = `int(distance/3.0)`, clamped to `[10, 80]`.

**Commit:** `feat(jitter): add Bezier curve mouse movement algorithm`

---

### Story 2.2 — Human Interaction Simulation

**Goal:** Scroll and typing simulation that produces human-realistic interaction patterns.

#### Sub-task 2.2.1 — Scroll simulation
**Test file:** `internal/jitter/scroll_test.go`

```go
// Tests (using a mock page that records scroll calls):
// TestHumanScroll_TotalPixels: sum of all scroll calls equals totalPixels ±5%
// TestHumanScroll_ChunkSizes: each chunk is between 50px and 250px
// TestHumanScroll_PauseBetweenChunks: at least one pause between chunks (mock sleep tracking)
// TestHumanScroll_ContextCancellation: returns ctx.Err() when context cancelled
// TestHumanScroll_DirectionReversals: ~20% of calls include an upward scroll segment
```

**Implementation:** `internal/jitter/scroll.go`

```go
// HumanScroll simulates a human scrolling totalPixels downward on a rod page.
// Scroll gestures: 80-200px chunks, LogNormal(120, 0.4).
// Reading pauses: LogNormal(400ms, 0.5) between chunks.
// Direction reversals: 20% probability of one upward re-read per profile.
func HumanScroll(ctx context.Context, mouse ScrollMouse, totalPixels float64) error
```

`ScrollMouse` is a small interface (`Scroll(x, y float64, steps int) error`) so it can be faked in tests without a real browser.

**Commit:** `feat(jitter): add human scroll simulation`

---

#### Sub-task 2.2.2 — Typing simulation
**Test file:** `internal/jitter/typing_test.go`

```go
// Tests (using a mock keyboard):
// TestHumanType_AllCharactersTyped: all characters of input text are typed
// TestHumanType_NoInstantKeystrokes: no two consecutive keystrokes within 10ms
// TestHumanType_ErrorRate: ~4% of calls produce a backspace-then-retype sequence (stochastic, tolerance ±2%)
// TestHumanType_WPMRange: effective WPM is between 40 and 90 for a 50-char string (timing mock)
// TestHumanType_ContextCancellation: returns ctx.Err() when context cancelled
```

**Implementation:** `internal/jitter/typing.go`

```go
// HumanType simulates character-by-character typing with realistic IKI, error rate, and corrections.
// Never uses paste events. Always dispatches individual key events via the Keyboard interface.
func HumanType(ctx context.Context, kb Keyboard, text string) error

type Keyboard interface {
    Press(key rune) error
    Backspace() error
}
```

**Commit:** `feat(jitter): add human typing simulation with error rate and IKI`

---

### Story 2.3 — Browser Factory

**Goal:** A production-ready go-rod browser configured with all anti-detection measures from the research.

#### Sub-task 2.3.1 — Browser factory (launcher configuration)
**Test file:** `internal/config/browser_test.go`

```go
// Tests:
// TestNewBrowser_LaunchesWithoutError: (build tag: integration) browser starts and connects
// TestBrowserConfig_ProxySet: launcher includes --proxy-server when ProxyAddr is non-empty
// TestBrowserConfig_NoProxyWhenEmpty: launcher omits --proxy-server when ProxyAddr is empty
// TestBrowserConfig_UserDataDir: launcher sets correct UserDataDir from config
// TestBrowserConfig_KeepsUserDataDir: KeepUserDataDir is always set (profile persistence)
// TestBrowserConfig_NoSandbox: NoSandbox flag is set (required on Linux)
// TestBrowserConfig_DisableWebRTC: --disable-webrtc flag is set (prevent LAN IP leak)
```

**Implementation:** `internal/config/browser.go`

Complete browser factory implementing all flags from the research:
- `--disable-blink-features=AutomationControlled`
- `--window-size=1920,1080`
- `--disable-webrtc`
- `--use-gl=angle --use-angle=swiftshader` (GPU-less EC2)
- `--lang=en-US`
- `--disable-dev-shm-usage`
- `NoSandbox(true)`
- `UserDataDir + KeepUserDataDir`
- Proxy + `MustHandleAuth` goroutine if proxy configured
- Rotate browser every `N` navigations (configurable, default 50)

**Commit:** `feat(infrastructure): add go-rod browser factory with anti-detection launcher flags`

---

#### Sub-task 2.3.2 — go-rod/stealth integration and custom patches
**Test file:** `internal/config/browser_stealth_test.go`

```go
// Tests (build tag: integration):
// TestStealth_NavigatorWebdriverFalsy: page.Eval("navigator.webdriver") returns undefined/falsy
// TestStealth_PluginsNonEmpty: navigator.plugins.length > 0
// TestStealth_OuterDimensions: outerWidth equals innerWidth
// TestCustomPatches_ScreenWidth: screen.width returns 1920
// TestCustomPatches_ColorDepth: screen.colorDepth returns 24
```

**Implementation:** Add to `internal/config/browser.go`:

1. Apply `stealth.MustPage(browser)` for every new page (return a `newPage()` helper)
2. Apply custom `EvalOnNewDocument` patches (screen properties, AudioContext noise)
3. Return a browser wrapper that ensures `stealth.Page()` is always used

**Commit:** `feat(infrastructure): apply stealth patches and custom EvalOnNewDocument evasions`

---

#### Sub-task 2.3.3 — BrowserGate extension hijack
**Test file:** `internal/config/browser_hijack_test.go`

```go
// Tests (build tag: integration):
// TestExtensionHijack_KnownExtensionReturns200: fetch to known extension ID returns 200
// TestExtensionHijack_UnknownExtensionRefused: fetch to unknown extension ID fails with network error
// TestExtensionHijack_ManifestJSON: known extension returns valid JSON manifest
// TestExtensionHijack_SimulatesExactly5: exactly 5 extensions are configured
```

**Implementation:** `SetupExtensionHijack(browser *rod.Browser) *rod.HijackRouter` in `internal/config/browser.go`

Extension IDs: uBlock Origin, Google Docs Offline, Grammarly, 1Password, LastPass.

**Commit:** `feat(infrastructure): add BrowserGate extension hijack to simulate installed extensions`

---

### Story 2.4 — Session Management

**Goal:** Cookie injection, session health checking, and block detection. The browser client is usable before LinkedIn-specific scraping logic is added.

#### Sub-task 2.4.1 — Session health checker
**Test file:** `internal/config/session_test.go`

```go
// Tests (using fake page URL responses):
// TestDetectBlock_None: profile page URL returns BlockNone
// TestDetectBlock_Authwall: /authwall URL returns BlockAuthwall
// TestDetectBlock_Challenge: /checkpoint/challenge URL returns BlockChallenge
// TestDetectBlock_SoftEmpty: body with <100 chars returns BlockSoftEmpty
// TestDetectBlock_Loginwall: loginwall URL returns BlockLoginwall (maps to BlockChallenge)
```

**Implementation:** `internal/config/session.go`

```go
func DetectBlock(page *rod.Page) BlockType
func InjectCookies(browser *rod.Browser, cookiesJSON string) error // parses li_at + JSESSIONID
func ExtractSessionCookies(browser *rod.Browser) (liAt, jsessionID string, err error)
```

**Commit:** `feat(infrastructure): add session cookie injection and block detection`

---

## Stage 3 — Data Extraction

**Branch:** `feat/stage-3-data-extraction`  
**PR title:** `feat(extraction): Voyager API client, JSON-LD parser, post extractor, goquery HTML parsing`  
**Goal:** All LinkedIn data extraction logic — Voyager API calls, JSON-LD fallback, and post extraction. No scraping orchestration yet, just the extraction primitives.  
**Prerequisite:** Stage 2 merged.

---

### Story 3.1 — Voyager API Client

**Goal:** Authenticated HTTP client for LinkedIn's internal Voyager API. Returns structured Go types.

#### Sub-task 3.1.1 — Voyager HTTP client with required headers
**Test file:** `internal/linkedin/voyager_client_test.go`

```go
// Tests (using httptest.NewServer):
// TestVoyagerClient_SetsAcceptHeader: every request includes accept: application/vnd.linkedin.normalized+json+2.1
// TestVoyagerClient_SetsCSRFToken: csrf-token header derived from JSESSIONID (quotes stripped)
// TestVoyagerClient_SetsRestliProtocol: x-restli-protocol-version: 2.0.0 header present
// TestVoyagerClient_SetsUserAgent: user-agent matches expected Chrome UA string
// TestVoyagerClient_SendsCookies: li_at and JSESSIONID cookies attached to request
```

**Implementation:** `internal/linkedin/voyager_client.go`

```go
type VoyagerClient struct {
    httpClient *http.Client
    liAt       string
    jsessionID string // without surrounding quotes
}

func NewVoyagerClient(liAt, jsessionID string) *VoyagerClient
func (c *VoyagerClient) do(ctx context.Context, path string) (*http.Response, error)
```

**Commit:** `feat(linkedin): add Voyager API HTTP client with required authentication headers`

---

#### Sub-task 3.1.2 — Profile data extraction from Voyager response
**Test file:** `internal/linkedin/profile_parser_test.go`

```go
// Tests (using fixture JSON files in testdata/):
// TestParseProfileResponse_Name: extracts correct firstName + lastName
// TestParseProfileResponse_Headline: extracts headline
// TestParseProfileResponse_Location: extracts location string
// TestParseProfileResponse_About: extracts summary/about text
// TestParseProfileResponse_MissingFields: returns empty strings for absent fields (no panic)
// TestParseProfileResponse_InvalidJSON: returns descriptive error
```

**Implementation:** Add `parseProfileResponse(body []byte) (ProfileData, error)` to `internal/linkedin/voyager_client.go`

Parse the Voyager `included` array: find the node where `entityUrn` contains `"fsd_profile:"` and `firstName` is present. Extract all required fields.

Add fixture file: `internal/linkedin/testdata/profile_response.json` — sanitized sample Voyager response.

**Commit:** `feat(linkedin): add Voyager profile response parser`

---

#### Sub-task 3.1.3 — Voyager FetchProfile method
**Test file:** `internal/linkedin/fetch_profile_test.go`

```go
// Tests (httptest server returning fixture):
// TestFetchProfile_Success: returns populated ProfileData for valid slug
// TestFetchProfile_404: returns ErrNotFound
// TestFetchProfile_429: returns ErrRateLimitExceeded
// TestFetchProfile_ContextCancellation: returns context error when cancelled
```

**Implementation:** `func (c *VoyagerClient) FetchProfile(ctx context.Context, slug string) (ProfileData, error)`

Profile endpoint: `GET /voyager/api/identity/dash/profiles?q=memberIdentity&memberIdentity={slug}&decorationId=com.linkedin.voyager.dash.deco.identity.profile.FullProfileWithEntities-93`

**Commit:** `feat(linkedin): add VoyagerClient.FetchProfile method`

---

### Story 3.2 — HTML Parsing & JSON-LD Fallback

**Goal:** Parse LinkedIn profile HTML for JSON-LD structured data when Voyager is unavailable or returns incomplete data.

#### Sub-task 3.2.1 — JSON-LD extractor
**Test file:** `internal/linkedin/jsonld_test.go`

```go
// Tests:
// TestExtractJSONLD_Person_Name: extracts name from @type:Person node
// TestExtractJSONLD_Person_Headline: extracts headline
// TestExtractJSONLD_Person_Location: extracts location
// TestExtractJSONLD_NoPersonNode: returns zero ProfileData without error (not a panic)
// TestExtractJSONLD_InvalidJSON: returns error
// TestExtractJSONLD_MultipleGraphNodes: correctly finds Person among other @type values
```

**Implementation:** `internal/linkedin/jsonld.go`

```go
// ExtractJSONLD parses all <script type="application/ld+json"> tags from HTML
// and extracts a ProfileData from the @type:Person node if present.
func ExtractJSONLD(html string) (ProfileData, error)
```

Uses `goquery` to find script tags, then `encoding/json` to parse the `@graph` array.

Add fixture: `internal/linkedin/testdata/profile.html` — sanitized sample LinkedIn profile page.

**Commit:** `feat(linkedin): add JSON-LD profile extractor as Voyager fallback`

---

#### Sub-task 3.2.2 — Post extractor
**Test file:** `internal/linkedin/posts_test.go`

```go
// Tests (using httptest + fixture HTML):
// TestExtractPosts_ReturnsUpTo3: returns at most 3 posts even if more are present
// TestExtractPosts_EmptyActivity: returns empty slice without error
// TestExtractPosts_StripsHTML: returned strings are plain text (no HTML tags)
// TestExtractPosts_MaxLength: each post text truncated to 1000 chars
```

**Implementation:** `internal/linkedin/posts.go`

```go
// ExtractPosts parses the /recent-activity/shares/ page HTML and returns
// up to maxPosts most recent post texts.
func ExtractPosts(html string, maxPosts int) ([]string, error)
```

Uses `goquery` to find article elements and extract text content.

**Commit:** `feat(linkedin): add post extractor from recent-activity page`

---

### Story 3.3 — Browser Client Implementation

**Goal:** Implement `BrowserClient` using go-rod to navigate LinkedIn and call the extraction primitives.

#### Sub-task 3.3.1 — BrowserClient struct and VisitProfile
**Test file:** `internal/linkedin/browser_client_test.go`

```go
// Tests (build tag: integration):
// TestBrowserClient_VisitProfile_ReturnsData: visits a real or mock profile, returns ProfileData
// TestBrowserClient_VisitProfile_DetectsBlock: returns ErrBlockDetected when LinkedIn blocks
// TestBrowserClient_VisitProfile_DwellTime: navigation takes ≥20s (dwell time enforced)
// TestBrowserClient_implements_BrowserClient: compile-time interface check
```

**Implementation:** `internal/linkedin/browser_client.go`

```go
type BrowserClient struct {
    browser    *rod.Browser
    voyager    *VoyagerClient
    mouse      MouseController
    navCount   int
    maxNavs    int // restart browser at this count (default 50)
}

// VisitProfile:
// 1. Navigate to profileURL
// 2. Check for block (DetectBlock)
// 3. Simulate human scroll (HumanScroll — full profile)
// 4. Extract data: try Voyager API first, fall back to JSON-LD
// 5. Enforce dwell time: LogNormalDuration(wordCount)
// 6. Return ProfileData
func (c *BrowserClient) VisitProfile(ctx context.Context, profileURL string) (ProfileData, error)
```

**Commit:** `feat(linkedin): implement BrowserClient.VisitProfile with Voyager + JSON-LD extraction`

---

#### Sub-task 3.3.2 — LikeRecentPost and SearchProfiles
**Test file:** `internal/linkedin/browser_client_actions_test.go`

```go
// Tests (build tag: integration):
// TestBrowserClient_LikeRecentPost_Success: like action executes without error
// TestBrowserClient_LikeRecentPost_NoPost: returns error when no recent post found
// TestBrowserClient_SearchProfiles_ReturnsURLs: returns ≥1 profile URL for known query
// TestBrowserClient_SearchProfiles_RespectsLimit: never returns more than limit URLs
```

**Implementation:** Add to `internal/linkedin/browser_client.go`

`LikeRecentPost`: navigate to `/recent-activity/shares/`, find first post like button, WindMouse to it, click. Includes human dwell and scroll before clicking.

`SearchProfiles`: navigate to LinkedIn People search with Boolean query + location filter, paginate up to `limit` profile URLs.

**Commit:** `feat(linkedin): implement BrowserClient.LikeRecentPost and SearchProfiles`

---

### Story 3.4 — Retry & Backoff

**Goal:** Robust retry logic wrapping all external calls.

#### Sub-task 3.4.1 — Exponential backoff with jitter
**Test file:** `internal/jitter/backoff_test.go`

```go
// Tests:
// TestBackoffDuration_Attempt0: returns ~30s (base)
// TestBackoffDuration_Attempt1: returns ~60s
// TestBackoffDuration_AttemptMax: capped at 10min
// TestBackoffDuration_HasJitter: 100 calls for same attempt produce different durations
// TestRetryWithBackoff_SucceedsOnFirstTry: calls fn once
// TestRetryWithBackoff_RetriesOnError: retries up to maxAttempts
// TestRetryWithBackoff_DoesNotRetryNonRetryable: ErrBlockDetected is not retried
```

**Implementation:** `internal/jitter/backoff.go`

```go
func BackoffDuration(attempt int) time.Duration  // base=30s, 2x, max=10min, ±30% jitter

// RetryWithBackoff calls fn up to maxAttempts times.
// Stops retrying if ctx is done or fn returns a non-retryable error.
// Non-retryable: ErrBlockDetected, ErrSessionExpired, ErrInvalidTransition.
func RetryWithBackoff(ctx context.Context, maxAttempts int, fn func() error) error
```

**Commit:** `feat(jitter): add exponential backoff with jitter for external call retries`

---

## Stage 4 — Scrape Phase

**Branch:** `feat/stage-4-scrape`  
**PR title:** `feat(scrape): rate-limited prospect discovery, deduplication, DynamoDB adapter, scrape cron`  
**Goal:** The `scrape` subcommand discovers 20 new prospects per run, deduplicates against DynamoDB, and stores them in the `Scanned` state. Rate limiting prevents exceeding daily profile-view limits.  
**Prerequisite:** Stage 3 merged.

---

### Story 4.1 — DynamoDB Adapter

**Goal:** Implement `ProspectRepository` and `RateLimiter` against DynamoDB.

#### Sub-task 4.1.1 — DynamoDB ProspectRepository: Save and Get
**Test file:** `internal/dynamo/prospect_repo_test.go`

```go
// Unit tests (with fake DynamoDB using testify/mock or hand-rolled):
// TestProspectRepo_Save_Succeeds: marshals and puts item without error
// TestProspectRepo_Save_UpdatesGSI1: GSI1PK and GSI1SK attributes are set
// TestProspectRepo_Get_Found: GetItem returns correct Prospect
// TestProspectRepo_Get_NotFound: returns ErrNotFound when item missing
// TestProspectRepo_Get_ContextTimeout: returns error on context cancellation

// Integration test (build tag: integration — requires real DynamoDB local):
// TestProspectRepo_SaveThenGet: save then get round-trip produces identical struct
```

**Implementation:** `internal/dynamo/prospect_repo.go`

Use `aws-sdk-go-v2/feature/dynamodb/expression` for all expressions (never raw strings).
Use `aws-sdk-go-v2/feature/dynamodb/attributevalue` for marshal/unmarshal.
Context with 5s timeout on every DynamoDB call.

**Commit:** `feat(dynamo): implement ProspectRepository Save and Get`

---

#### Sub-task 4.1.2 — ProspectRepository: InsertIfNew and ListByState
**Test file:** add to `internal/dynamo/prospect_repo_test.go`

```go
// TestInsertIfNew_NewProspect: returns inserted=true
// TestInsertIfNew_Duplicate: returns inserted=false, no error
// TestListByState_ReturnsOnlyDueItems: only items with NextActionAt <= asOf
// TestListByState_ReturnsOnlyMatchingState: no cross-state contamination
// TestListByState_EmptyTable: returns empty slice, no error
// TestListByStateOrderedByScore_Descending: highest score first
```

**Implementation:** Add methods to `internal/dynamo/prospect_repo.go`

`InsertIfNew`: `PutItem` with `ConditionExpression: attribute_not_exists(PK)`.

`ListByState`: GSI1 query `GSI1PK = state AND GSI1SK <= now`.

`ListByStateOrderedByScore`: GSI2 query `GSI2PK = state` with `ScanIndexForward=false`.

**Commit:** `feat(dynamo): implement ProspectRepository InsertIfNew and ListByState`

---

#### Sub-task 4.1.3 — DynamoDB RateLimiter
**Test file:** `internal/dynamo/rate_limiter_test.go`

```go
// Unit tests:
// TestRateLimiter_AcquireUnderLimit: returns nil error
// TestRateLimiter_AcquireAtLimit: returns ErrRateLimitExceeded
// TestRateLimiter_UsesToday: correct PK format uses today's UTC date
// TestRateLimiter_TTLSetCorrectly: TTL is tomorrow at UTC midnight + buffer

// Integration tests (build tag: integration):
// TestRateLimiter_AtomicIncrement: concurrent Acquire from N goroutines, exactly max succeed
```

**Implementation:** `internal/dynamo/rate_limiter.go`

`UpdateItem` with `if_not_exists(Count, 0) + 1` and `ConditionExpression: attribute_not_exists(Count) OR Count < :max`.

Scopes: `"profile_views"` (max from config), `"connection_requests"` (max from config).

**Commit:** `feat(dynamo): implement DynamoDB-backed RateLimiter with atomic token bucket`

---

### Story 4.2 — Scrape Use Case

**Goal:** Orchestrate the discovery of new prospects.

#### Sub-task 4.2.1 — ScrapePhase core logic
**Test file:** `internal/scrape_test.go`

```go
// All tests use fake repository and fake BrowserClient:
// TestScrapePhase_RunsSearchAndVisits: calls SearchProfiles then VisitProfile for each result
// TestScrapePhase_RespectsMaxProspects: stops at MaxProspectsPerRun even if more available
// TestScrapePhase_SkipsDuplicates: InsertIfNew=false means VisitProfile not called again
// TestScrapePhase_RespectsRateLimit: stops when RateLimiter.Acquire returns ErrRateLimitExceeded
// TestScrapePhase_HandlesBrowserError: logs error and continues with next profile (non-fatal)
// TestScrapePhase_SetsCorrectInitialState: prospects saved with StateScanned and NextActionAt in future
// TestScrapePhase_ContextCancellation: exits cleanly when ctx cancelled
```

**Implementation:** `internal/scrape.go`

```go
type ScrapePhase struct {
    browser   BrowserClient
    repo      ProspectRepository
    limiter   RateLimiter
    maxPerRun int
    searchQ   string
    location  string
    logger    *slog.Logger
}

func (s *ScrapePhase) Run(ctx context.Context) error
```

**Commit:** `feat(usecase): implement ScrapePhase with rate-limited prospect discovery`

---

#### Sub-task 4.2.2 — Wire scrape command
**Test file:** `cmd/in-network-explorer/scrape_test.go`

```go
// TestRunScrape_WiresCorrectly: (build tag: integration) command runs without panic
// TestRunScrape_RespectsContext: exits when context cancelled
```

**Implementation:** `cmd/in-network-explorer/scrape.go`

Fully wired `runScrape(ctx, cfg)` — creates all dependencies (DynamoDB client, browser, rate limiter), constructs `ScrapePhase`, calls `.Run(ctx)`, logs outcome.

**Commit:** `feat(cmd): wire scrape subcommand with all dependencies`

---

### Story 4.3 — Warmup Use Case

**Goal:** Advance `Scanned` prospects to `Liked` on Day 2.

#### Sub-task 4.3.1 — WarmupPhase core logic
**Test file:** `internal/warmup_test.go`

```go
// TestWarmupPhase_AdvancesScannedToLiked: calls LikeRecentPost and transitions state
// TestWarmupPhase_RespectsSchedule: skips prospects not yet due (NextActionAt in future)
// TestWarmupPhase_HandlesNoRecentPost: transitions to StateSkipped when no post found
// TestWarmupPhase_RespectsRateLimit: stops when profile view limit reached
// TestWarmupPhase_SavesTransitionedState: saves updated prospect after Transition()
// TestWarmupPhase_ContextCancellation: exits cleanly
```

**Implementation:** `internal/warmup.go`

```go
type WarmupPhase struct {
    browser BrowserClient
    repo    ProspectRepository
    limiter RateLimiter
    logger  *slog.Logger
}

func (w *WarmupPhase) Run(ctx context.Context) error
```

Queries GSI1 for `StateScanned` prospects due today. For each: `VisitProfile`, `LikeRecentPost`, `Transition(StateLiked)`, `Save`.

**Commit:** `feat(usecase): implement WarmupPhase for Day 2 post-liking`

---

#### Sub-task 4.3.2 — Wire warmup into scrape command (same cron)
**Implementation:** `cmd/in-network-explorer/scrape.go` — `runScrape` runs `ScrapePhase` then `WarmupPhase` sequentially. The single cron job at `09:00` handles both new discovery and warm-up advancement.

**Commit:** `feat(cmd): run WarmupPhase after ScrapePhase in scrape subcommand`

---

## Stage 5 — Warm-Up Pipeline (Draft Queue)

**Branch:** `feat/stage-5-draft-queue`  
**PR title:** `feat(warmup): Day 3 draft queue, prospect advancement to Drafted state`  
**Goal:** Prospects in `Liked` state that are due on Day 3 are transitioned to `Drafted`. The human-facing pipeline is now complete — every prospect reaching `Drafted` has a prepared message waiting for the human to send.  
**Prerequisite:** Stage 4 merged.

---

### Story 5.1 — Draft Advancement

#### Sub-task 5.1.1 — Advance Liked → Drafted
**Test file:** `internal/warmup_draft_test.go`

```go
// TestDraftPhase_AdvancesLikedToDrafted: queries GSI1 for StateLiked due today, transitions each
// TestDraftPhase_RequiresDraftedMessage: does not transition if DraftedMessage is empty
// TestDraftPhase_SavesState: Save called after each successful transition
// TestDraftPhase_RespectsLimit: stops at MaxConnectionReqsPerDay
```

**Implementation:** Extend `internal/warmup.go` with a `DraftPhase` struct (or `AdvanceToDraft` method on `WarmupPhase`).

This phase does NOT send any connection request — it only updates state to `Drafted` so the human can review and send. The AI analysis (Stage 6) populates `DraftedMessage` before this phase runs.

**Commit:** `feat(usecase): implement draft advancement phase (Liked → Drafted)`

---

#### Sub-task 5.1.2 — Warmup cron subcommand
**Implementation:** `cmd/in-network-explorer/scrape.go` already runs scrape + warmup. This sub-task separates it if needed, or adds `--phase` flag. No-op commit unless there is wiring to add.

Review whether `analyze` should run before `warmup`'s draft advancement (yes — analyze must run first to populate `DraftedMessage`). Document the correct cron order in `cmd/in-network-explorer/main.go` comments.

**Commit:** `docs(cmd): document correct cron phase order (scrape → analyze → warmup-draft → report)`

---

## Stage 6 — AI Analysis Engine

**Branch:** `feat/stage-6-analysis`  
**PR title:** `feat(analysis): Bedrock LLM scoring, self-critique, prompt injection prevention, analyze cron`  
**Goal:** The `analyze` subcommand scores `Scanned` prospects 1-10 and drafts personalized 150-300 char connection messages using Claude Haiku 4.5. Includes prompt injection prevention and self-critique gating.  
**Prerequisite:** Stage 5 merged.

---

### Story 6.1 — Bedrock LLM Client

#### Sub-task 6.1.1 — Prompt templates and versioning
**Test file:** `internal/bedrock/prompt_test.go`

```go
// Tests:
// TestBuildScoringPayload_StructuredCorrectly: system field separate from user content
// TestBuildScoringPayload_ProfileDataInXMLTags: content wrapped in <profile_data> tags
// TestBuildScoringPayload_MaxTokens: max_tokens=512 (blast radius limit)
// TestBuildScoringPayload_AnthropicVersion: anthropic_version field present
```

**Implementation:** `internal/bedrock/prompts.go`

Scoring system prompt, XML-structured user content builder, critique prompt. Prompts are also loadable from `prompts/scoring.json` (path from config) to support DSPy-compiled prompt updates without recompilation.

```go
type PromptConfig struct {
    ScoringSystem string `json:"scoring_system"`
    CritiqueUser  string `json:"critique_user"`
    Version       string `json:"version"`
}

func LoadPromptConfig(path string) (*PromptConfig, error)
func DefaultPromptConfig() *PromptConfig
```

**Commit:** `feat(bedrock): add prompt templates with version-controlled config file`

---

#### Sub-task 6.1.2 — Input sanitization against prompt injection
**Test file:** `internal/bedrock/sanitize_test.go`

```go
// Tests:
// TestSanitizeProfileField_StripsBOM: strips \ufeff
// TestSanitizeProfileField_StripsDirectionOverrides: strips \u202e, \u200b etc.
// TestSanitizeProfileField_TruncatesAt2000: input of 5000 chars returns exactly 2000 runes
// TestSanitizeProfileField_PreservesNormalText: typical profile text unchanged
// TestSanitizeProfileField_PreservesGermanChars: ä ö ü ß preserved
```

**Implementation:** `internal/bedrock/sanitize.go`

```go
// SanitizeProfileField strips Unicode control/direction characters and
// enforces a 2000-rune hard cap. Returns cleaned string.
func SanitizeProfileField(raw string) string
```

**Commit:** `feat(bedrock): add profile field sanitization to prevent prompt injection`

---

#### Sub-task 6.1.3 — Bedrock InvokeModel client
**Test file:** `internal/bedrock/llm_client_test.go`

```go
// Tests (using httptest server or Bedrock stub):
// TestLLMClient_ParsesScoreResult: valid JSON response unmarshals correctly
// TestLLMClient_InvalidJSON: returns descriptive error (not a panic)
// TestLLMClient_ScoreOutOfRange: returns error if score > 10 or < 1
// TestLLMClient_EmptyMessage: returns error if message is empty string
// TestLLMClient_ContextTimeout: returns context error
// TestLLMClient_implements_LLMClient: compile-time interface check
```

**Implementation:** `internal/bedrock/llm_client.go`

```go
type LLMClient struct {
    client   *bedrockruntime.Client
    modelID  string
    prompts  *PromptConfig
}

func (c *LLMClient) ScoreAndDraft(ctx context.Context, p *Prospect, examples []Prospect) (ScoreResult, error)
func (c *LLMClient) Critique(ctx context.Context, message string) (int, error)
```

Always validate response with `json.Unmarshal` into strict struct. Discard raw LLM output if parsing fails.

**Commit:** `feat(bedrock): implement Bedrock LLMClient with ScoreAndDraft and Critique`

---

### Story 6.2 — Self-Critique Loop

#### Sub-task 6.2.1 — Scoring with critique gating
**Test file:** `internal/analyze_test.go`

```go
// All tests use fake LLMClient and fake ProspectRepository:
// TestAnalyzePhase_ScoresAndDrafts: ScoreAndDraft called for each Scanned prospect
// TestAnalyzePhase_CritiqueGating: low critique score triggers one regeneration
// TestAnalyzePhase_MaxRegenerations: no more than 2 regeneration rounds
// TestAnalyzePhase_SavesScoreAndMessage: Save called with score and message set
// TestAnalyzePhase_ConcurrencyLimit: at most 3 concurrent LLM calls (fake semaphore tracking)
// TestAnalyzePhase_NonFatalLLMError: LLM error for one prospect doesn't abort batch
// TestAnalyzePhase_ContextCancellation: exits cleanly
```

**Implementation:** `internal/analyze.go`

```go
type AnalyzePhase struct {
    llm     LLMClient
    repo    ProspectRepository
    embeds  EmbeddingStore  // may be nil if RAG not yet set up
    concurrency int
    logger  *slog.Logger
}

func (a *AnalyzePhase) Run(ctx context.Context) error
```

`errgroup.SetLimit(concurrency)` for batch. Per-prospect: `ScoreAndDraft` → `Critique` → if score < 9 and attempts < 2, regenerate with critique feedback → `Save`.

**Commit:** `feat(usecase): implement AnalyzePhase with self-critique gating`

---

#### Sub-task 6.2.2 — Wire analyze command
**Test file:** `cmd/in-network-explorer/analyze_test.go`

```go
// TestRunAnalyze_WiresCorrectly: (build tag: integration) command runs without panic
```

**Implementation:** `cmd/in-network-explorer/analyze.go`

Fully wired `runAnalyze(ctx, cfg)`. Loads prompt config from file (with fallback to default). Wires `LLMClient`, `ProspectRepository`, constructs `AnalyzePhase`, runs it.

**Commit:** `feat(cmd): wire analyze subcommand with Bedrock LLM client`

---

## Stage 7 — RAG & Embeddings

**Branch:** `feat/stage-7-rag`  
**PR title:** `feat(rag): Qdrant vector store, Bedrock embeddings, few-shot retrieval for message drafting`  
**Goal:** Profile embeddings stored in Qdrant. At analysis time, retrieve the top-3 most similar past accepted profiles as few-shot examples to improve message personalization.  
**Prerequisite:** Stage 6 merged.

---

### Story 7.1 — Bedrock Embedding Client

#### Sub-task 7.1.1 — Bedrock Titan Embeddings adapter
**Test file:** `internal/bedrock/embedding_client_test.go`

```go
// TestEmbeddingClient_ReturnsVector: valid text returns float32 slice of length > 0
// TestEmbeddingClient_ConsistentLength: two different texts return same-length vectors
// TestEmbeddingClient_EmptyInput: returns error for empty string
// TestEmbeddingClient_implements_EmbeddingClient: compile-time interface check
```

**Implementation:** `internal/bedrock/embedding_client.go`

Model: `amazon.titan-embed-text-v2:0`. Input: text string. Output: `[]float32`.

**Commit:** `feat(bedrock): implement Bedrock Titan Embeddings EmbeddingClient`

---

### Story 7.2 — Qdrant Store

#### Sub-task 7.2.1 — Qdrant EmbeddingStore adapter
**Test file:** `internal/qdrant/embedding_store_test.go`

```go
// TestEmbeddingStore_UpsertAndSearch: upsert a vector, search returns it in top-1 (build tag: integration)
// TestEmbeddingStore_FilterByOutcome: search with filter outcome=ACCEPTED only returns accepted items
// TestEmbeddingStore_EmptyCollection: search on empty collection returns empty slice without error
// TestEmbeddingStore_implements_EmbeddingStore: compile-time interface check
```

**Implementation:** `internal/qdrant/embedding_store.go`

Uses `github.com/qdrant/go-client` (gRPC). Creates collection if not exists with cosine distance. Payload fields: `profile_url`, `outcome`, `critique_score`, `message_sent`.

**Commit:** `feat(qdrant): implement Qdrant EmbeddingStore adapter`

---

### Story 7.3 — RAG Pipeline Integration

#### Sub-task 7.3.1 — RAG retrieval in AnalyzePhase
**Test file:** Add to `internal/analyze_test.go`

```go
// TestAnalyzePhase_InjectsFewShot: when EmbeddingStore returns results,
//   they are passed as examples to ScoreAndDraft
// TestAnalyzePhase_WorksWithoutRAG: when embeds is nil, ScoreAndDraft called with empty examples
// TestAnalyzePhase_StoresEmbeddingAfterScoring: EmbeddingStore.Upsert called after successful score
```

**Implementation:** Update `internal/analyze.go`

After scoring: `EmbeddingClient.Embed(profileText)` → `EmbeddingStore.Upsert(...)`. Before scoring: `EmbeddingStore.SearchSimilar(vector, filter{outcome:ACCEPTED}, topK=3)` → inject as `examples` to `ScoreAndDraft`.

**Commit:** `feat(usecase): integrate RAG few-shot retrieval into AnalyzePhase`

---

#### Sub-task 7.3.2 — Outcome feedback recording
**Test file:** `internal/feedback_test.go`

```go
// TestRecordOutcome_Accepted: updates Prospect with ConnectionAccepted=true, saves
// TestRecordOutcome_Edited: sets MessageEdited=true, stores MessageSent
// TestRecordOutcome_UpdatesQdrantPayload: EmbeddingStore.Upsert called with updated outcome
```

**Implementation:** `internal/feedback.go`

```go
// RecordOutcome records a human decision (accepted/rejected/edited) back to
// DynamoDB and updates the Qdrant payload for future RAG retrieval.
func RecordOutcome(ctx context.Context, repo ProspectRepository, store EmbeddingStore,
    profileURL string, accepted bool, messageSent string, edited bool) error
```

Called from `report` subcommand when human provides feedback via the report interface.

**Commit:** `feat(usecase): add RecordOutcome for human feedback loop`

---

## Stage 8 — Calibration

**Branch:** `feat/stage-8-calibration`  
**PR title:** `feat(calibration): Platt scaling, Bayesian bucket scoring, EWMA drift detection, calibrate cron`  
**Goal:** The LLM's raw 1-10 score is calibrated against actual human decisions. A weekly calibration cron job updates the calibration parameters stored in DynamoDB.  
**Prerequisite:** Stage 7 merged.

---

### Story 8.1 — Bayesian Calibration

#### Sub-task 8.1.1 — Beta distribution bucket tracking
**Test file:** `internal/calibrate_test.go`

```go
// TestBetaBucket_UpdateOnAccept: alpha increments
// TestBetaBucket_UpdateOnReject: beta increments
// TestBetaBucket_ExpectedRate: alpha/(alpha+beta) converges to true rate over 100 outcomes
// TestBetaBucket_CredibleInterval_WidthDecreases: CI narrows as N grows
// TestScoreBucket: score 7 maps to bucket "7-8"
```

**Implementation:** `internal/calibrate.go` (partial)

```go
type BetaBucket struct {
    Bucket string  // "1-2", "3-4", "5-6", "7-8", "9-10"
    Alpha  float64
    Beta   float64
}

func (b *BetaBucket) Update(accepted bool)
func (b *BetaBucket) ExpectedRate() float64
func ScoreBucket(score int) string
```

Stored per-bucket in DynamoDB as a calibration item (`PK: "CALIBRATION#bucket#7-8"`).

**Commit:** `feat(calibrate): add Bayesian Beta bucket calibration tracking`

---

#### Sub-task 8.1.2 — Platt scaling (logistic regression)
**Test file:** `internal/platt_test.go`

```go
// TestPlattScale_PerfectPositive: all accepted → calibrated score near 1.0 for high LLM score
// TestPlattScale_PerfectNegative: all rejected → calibrated score near 0.0
// TestPlattScale_MonotonicallyIncreasing: higher LLM score → higher calibrated score
// TestFitPlatt_RequiresMinimum30Samples: returns error if fewer than 30 outcomes
// TestFitPlatt_ProducesReasonableParameters: |a| and |b| are bounded
```

**Implementation:** Add to `internal/calibrate.go`

```go
type PlattParams struct {
    A float64
    B float64
}

// FitPlatt fits logistic regression to (llm_score, accepted) pairs.
// Requires at least 30 outcomes. Uses gradient descent (stdlib math, no ML library).
func FitPlatt(outcomes []CalibrationOutcome) (PlattParams, error)

// PlattScore converts raw LLM score to acceptance probability using fitted params.
func PlattScore(score int, p PlattParams) float64
```

Gradient descent implementation using only `math` stdlib. No external ML library.

**Commit:** `feat(calibrate): add Platt scaling for LLM score calibration`

---

#### Sub-task 8.1.3 — EWMA drift detection
**Test file:** `internal/drift_test.go`

```go
// TestEWMA_Update: EWMA converges toward new rate over time
// TestEWMA_DriftDetected: returns true when rate drops >1.5σ from historical mean
// TestEWMA_NoDriftInStableSystem: returns false for stable acceptance rate
```

**Implementation:** Add to `internal/calibrate.go`

```go
func UpdateEWMA(previous, newRate, alpha float64) float64  // alpha=0.2
func IsDriftDetected(ewma, historicalMean, historicalStdDev float64) bool
```

**Commit:** `feat(calibrate): add EWMA drift detection for score calibration monitoring`

---

### Story 8.2 — Calibrate Phase & Command

#### Sub-task 8.2.1 — CalibratePhase orchestration
**Test file:** Add to `internal/calibrate_test.go`

```go
// TestCalibratePhase_CollectsOutcomes: reads all ACCEPTED/REJECTED prospects from repo
// TestCalibratePhase_UpdatesBetaBuckets: updates DynamoDB calibration items
// TestCalibratePhase_FitsPlattWhen30Outcomes: calls FitPlatt and stores params
// TestCalibratePhase_SkipsPlattBelow30: logs warning, does not call FitPlatt
// TestCalibratePhase_LogsDriftWarning: logs warning when drift detected
```

**Implementation:** `internal/calibrate.go` (complete)

Reads all terminal-state prospects, updates Beta buckets, fits Platt if ≥30 outcomes, computes EWMA, stores all results in DynamoDB calibration items.

**Commit:** `feat(usecase): implement CalibratePhase orchestration`

---

#### Sub-task 8.2.2 — Wire calibrate command
**Implementation:** `cmd/in-network-explorer/calibrate.go`

Fully wired `runCalibrate(ctx, cfg)`. Scheduled weekly via separate cron entry.

**Commit:** `feat(cmd): wire calibrate subcommand`

---

## Stage 9 — Report

**Branch:** `feat/stage-9-report`  
**PR title:** `feat(report): 48-hour prospect report with HTML output and copy-to-clipboard messages`  
**Goal:** The `report` subcommand generates a human-readable HTML report of all `Drafted` prospects ranked by worthiness score, with copy-to-clipboard message buttons.  
**Prerequisite:** Stage 8 merged.

---

### Story 9.1 — Report Data Model & Logic

#### Sub-task 9.1.1 — ReportPhase use case
**Test file:** `internal/report_test.go`

```go
// TestReportPhase_QueriesDrafted: calls ListByStateOrderedByScore with StateDrafted
// TestReportPhase_ProducesReportData: returns correctly shaped ProspectReport struct
// TestReportPhase_RanksByScore: highest score appears first
// TestReportPhase_IncludesAllFields: profile URL, name, score, drafted message all present
// TestReportPhase_EmptyPipeline: returns empty report without error
```

**Implementation:** `internal/report.go`

```go
type ProspectReport struct {
    GeneratedAt time.Time
    Prospects   []ReportItem
}

type ReportItem struct {
    ProfileURL     string
    Name           string
    Headline       string
    Location       string
    WorthinessScore int
    CalibratedProb  float64
    DraftedMessage  string
    RecentPost      string // post that was liked on Day 2
    ScannedAt       time.Time
}

func (r *ReportPhase) Run(ctx context.Context) (*ProspectReport, error)
```

**Commit:** `feat(usecase): implement ReportPhase to produce ranked prospect report`

---

### Story 9.2 — HTML Renderer

#### Sub-task 9.2.1 — JSON report output
**Test file:** `internal/report/render_test.go`

```go
// TestRenderJSON_ValidJSON: output parses as valid JSON
// TestRenderJSON_ContainsAllProspects: all items in report appear in output
// TestRenderJSON_CorrectFieldNames: uses snake_case JSON field names
```

**Implementation:** `internal/report/render.go`

```go
func RenderJSON(w io.Writer, r *usecase.ProspectReport) error
```

**Commit:** `feat(report): add JSON report renderer`

---

#### Sub-task 9.2.2 — HTML report with copy-to-clipboard
**Test file:** add to `internal/report/render_test.go`

```go
// TestRenderHTML_ValidHTML: output parses without goquery error
// TestRenderHTML_HasCopyButtons: each item has a data-message attribute for copy-to-clipboard
// TestRenderHTML_ShowsScore: each item displays worthiness score
// TestRenderHTML_ProfileLinks: each item links to the LinkedIn profile URL
// TestRenderHTML_ScoreColor: score ≥8 has green indicator, score ≤4 has red
```

**Implementation:** Add `RenderHTML(w io.Writer, r *usecase.ProspectReport) error` to `internal/report/render.go`

Inline CSS only (no external stylesheet). Each prospect card shows: name (link to LinkedIn), headline, location, score with color indicator, drafted message with "Copy" button (clipboard API via `onclick`). Scores displayed as `score/10 (~calibratedPct% acceptance probability)`.

**Commit:** `feat(report): add HTML report renderer with copy-to-clipboard messages`

---

#### Sub-task 9.2.3 — Wire report command
**Implementation:** `cmd/in-network-explorer/report.go`

Writes `report-{timestamp}.json` and `report-{timestamp}.html` to the working directory. Logs file paths.

**Commit:** `feat(cmd): wire report subcommand with JSON and HTML output`

---

## Stage 10 — Production Hardening

**Branch:** `feat/stage-10-hardening`  
**PR title:** `feat(hardening): Chromium crash recovery, observability, Docker Compose, EC2 deployment`  
**Goal:** The system is production-ready. Chromium crashes are recovered automatically. Observability is complete. Docker Compose for local dev is functional and parity-validated against EC2.  
**Prerequisite:** Stage 9 merged.

---

### Story 10.1 — Chromium Reliability

#### Sub-task 10.1.1 — Browser rotation and crash recovery
**Test file:** `internal/config/browser_rotation_test.go`

```go
// TestBrowserRotation_RotatesAtMaxNavs: a new browser is created after maxNavs navigations
// TestBrowserRotation_RecoversFromCrash: browser.Connect error triggers restart
// TestBrowserRotation_ReInjectsProxy: new browser has proxy configured
// TestBrowserRotation_ClosesOldBrowser: old browser.Close() called before creating new one
```

**Implementation:** Add browser rotation to `internal/config/browser.go`

Wrap `*rod.Browser` in a `RotatingBrowser` struct that counts navigations and transparently restarts Chromium at `maxNavs`. Re-applies all launch flags, proxy auth, stealth, and extension hijack after restart.

**Commit:** `feat(infrastructure): add browser rotation and crash recovery`

---

### Story 10.2 — Observability

#### Sub-task 10.2.1 — Health check command
**Test file:** `cmd/in-network-explorer/health_test.go`

```go
// TestHealthCheck_DynamoReachable: exits 0 when DynamoDB responds
// TestHealthCheck_BedrockReachable: exits 0 when Bedrock responds
// TestHealthCheck_QdrantReachable: exits 0 when Qdrant responds
// TestHealthCheck_PrintsJSON: output is valid JSON health summary
```

**Implementation:** `cmd/in-network-explorer/health.go` — `health` subcommand (add to `cmd/in-network-explorer/main.go` switch).

Pings each external service, prints JSON summary: `{"dynamo":"ok","bedrock":"ok","qdrant":"ok","chrome":"ok"}`. Used by EC2 monitoring to detect degraded state.

**Commit:** `feat(cmd): add health check subcommand for observability`

---

#### Sub-task 10.2.2 — Block alert in report
**Implementation:** Add `BlockAlert bool` and `BlockDetectedAt *time.Time` to `ProspectReport`. When any cron run records a `BlockChallenge` in DynamoDB, the next report flags it prominently in HTML (red banner: "LinkedIn challenge detected on [date] — human intervention required").

**Test file:** add to `internal/report_test.go`

```go
// TestReportPhase_IncludesBlockAlert: when block item present in DynamoDB, BlockAlert=true
```

**Commit:** `feat(report): add block alert flag to prospect report`

---

### Story 10.3 — Deployment

#### Sub-task 10.3.1 — Docker Compose for local dev
**Test:** Manual verification — `docker compose up` starts both `rod` sidecar and `explorer` container. `docker compose run explorer scrape` completes without error (assuming `.env.local` has valid credentials).

**Implementation:** `docker-compose.yml` + `Dockerfile` + `.env.local.example`

```yaml
# docker-compose.yml (final, production-ready version)
services:
  rod:
    image: ghcr.io/go-rod/rod
    shm_size: "512m"
    ports: ["7317:7317"]
    restart: unless-stopped
  qdrant:
    image: qdrant/qdrant:latest
    ports: ["6333:6333", "6334:6334"]
    volumes: ["qdrant-data:/qdrant/storage"]
    restart: unless-stopped
  explorer:
    build: .
    env_file: .env.local
    environment:
      ROD_MANAGER_URL: ws://rod:7317
      QDRANT_ADDR: qdrant:6334
    volumes: ["chrome-profile:/chrome-profile"]
    depends_on: [rod, qdrant]
volumes:
  qdrant-data:
  chrome-profile:
```

**Commit:** `chore(deploy): add Docker Compose for local development`

---

#### Sub-task 10.3.2 — EC2 deployment script
**Implementation:** `scripts/setup-ec2.sh`

Bash script for initial EC2 setup:
- Install Google Chrome stable via `dnf`
- Install required font packages (Liberation, DejaVu, Noto)
- Install Qdrant as a systemd service (Docker)
- Set up cron entries (see below)
- Create `/var/log/explorer/` directory
- Set correct file permissions

Cron entries documented in script comments:
```cron
0 9  * * 1-5  /usr/local/bin/explorer scrape    >> /var/log/explorer/scrape.log  2>&1
0 10 * * 1-5  /usr/local/bin/explorer analyze   >> /var/log/explorer/analyze.log 2>&1
30 10 * * 1-5 /usr/local/bin/explorer scrape    >> /var/log/explorer/warmup.log  2>&1
0 6  */2 * *  /usr/local/bin/explorer report    >> /var/log/explorer/report.log  2>&1
0 8  * * 1    /usr/local/bin/explorer calibrate >> /var/log/explorer/calibrate.log 2>&1
```

**Commit:** `chore(deploy): add EC2 setup script with cron configuration`

---

#### Sub-task 10.3.3 — Architecture documentation
**Implementation:** `docs/architecture.md`

Mermaid diagrams:
1. **Class diagram**: domain entities + interfaces
2. **Sequence diagram**: full 3-day prospect lifecycle (Scanned → Liked → Drafted → human sends → Accepted → feedback recorded)
3. **Deployment diagram**: EC2 + DynamoDB + Qdrant + Bedrock + Residential Proxy + LinkedIn

**Commit:** `docs(architecture): add Mermaid diagrams for domain model, lifecycle, and deployment`

---

## Appendix A — Cron Schedule

```
# Berlin working hours only (09:00-18:00 CET)

# Phase 1: Discover new prospects + warm up Scanned→Liked
0 9  * * 1-5  explorer scrape

# Phase 2: Score + draft messages for Scanned prospects
0 10 * * 1-5  explorer analyze

# Phase 3: Advance Liked→Drafted (after analyze populates messages)
30 10 * * 1-5 explorer scrape   # WarmupPhase runs inside scrape command

# Phase 4: Human-review report every 48 hours
0 6  */2 * *  explorer report

# Phase 5: Weekly calibration (Monday morning)
0 8  * * 1    explorer calibrate
```

---

## Appendix B — Environment Variables Reference

| Variable | Required | Default | Description |
|---|---|---|---|
| `AWS_REGION` | Yes | — | AWS region for DynamoDB and Secrets Manager |
| `DYNAMO_TABLE` | Yes | — | DynamoDB table name |
| `BEDROCK_MODEL_ID` | No | `anthropic.claude-haiku-4-5-20251001-v1:0` | Bedrock model |
| `BEDROCK_REGION` | No | `AWS_REGION` | Bedrock region (may differ) |
| `LINKEDIN_COOKIES_SECRET` | Yes | — | Secrets Manager ARN for LinkedIn session cookies |
| `PROXY_ADDR` | No | — | Residential proxy host:port |
| `PROXY_USER` | No | — | Proxy username |
| `PROXY_PASS` | No | — | Proxy password |
| `CHROME_BIN` | No | auto | Path to Chrome binary (`/usr/bin/google-chrome-stable` on EC2) |
| `CHROME_PROFILE_DIR` | Yes | — | Path to persistent Chrome user data directory |
| `QDRANT_ADDR` | No | `localhost:6334` | Qdrant gRPC address |
| `QDRANT_COLLECTION` | No | `prospects` | Qdrant collection name |
| `MAX_PROFILE_VIEWS_PER_DAY` | No | `40` | Daily profile view cap |
| `MAX_CONNECTION_REQS_PER_DAY` | No | `10` | Daily connection request cap |
| `MAX_PROSPECTS_PER_RUN` | No | `20` | Max new prospects discovered per scrape run |
| `ANALYZE_CONCURRENCY` | No | `3` | Concurrent Bedrock calls during analyze phase |
| `PROMPT_CONFIG_PATH` | No | `prompts/scoring.json` | Path to DSPy-compiled prompt config |

---

## Appendix C — Key Library Versions (go.mod)

```
github.com/go-rod/rod          v0.116.x
github.com/go-rod/stealth      latest
github.com/aws/aws-sdk-go-v2   v1.x (latest)
github.com/qdrant/go-client    v1.x (latest)
github.com/PuerkitoBio/goquery v1.9.x
golang.org/x/sync              v0.x (latest)
```

All other functionality uses Go stdlib. No mocking frameworks. No ORM. No logger beyond `log/slog`.

---

## Appendix D — Test Commands

```bash
# Run all unit tests
go test ./...

# Run with race detector (required before every PR)
go test -race ./...

# Run integration tests (requires real AWS + LinkedIn session)
go test -tags integration ./...

# Run single test
go test -run TestProspect_Transition_Valid ./internal/...

# Vet
go vet ./...

# Build
go build ./cmd/...
```
