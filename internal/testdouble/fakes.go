// Package testdouble provides in-memory implementations of all domain
// interfaces for use in unit tests. Never import this package from production
// code.
package testdouble

import (
	"context"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/pavlomaksymov/in-network-explorer/domain"
)

// ── FakeProspectRepository ──────────────────────────────────────────────────

// FakeProspectRepository is a thread-safe in-memory prospect store.
type FakeProspectRepository struct {
	mu      sync.RWMutex
	records map[string]*domain.Prospect
}

// NewFakeProspectRepository returns an initialised FakeProspectRepository.
func NewFakeProspectRepository() *FakeProspectRepository {
	return &FakeProspectRepository{records: make(map[string]*domain.Prospect)}
}

// Ensure compile-time interface satisfaction.
var _ domain.ProspectRepository = (*FakeProspectRepository)(nil)

// Save creates or replaces the prospect keyed by ProfileURL.
func (r *FakeProspectRepository) Save(_ context.Context, p *domain.Prospect) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	clone := *p
	r.records[p.ProfileURL] = &clone
	return nil
}

// Get returns the prospect or domain.ErrNotFound.
func (r *FakeProspectRepository) Get(_ context.Context, profileURL string) (*domain.Prospect, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.records[profileURL]
	if !ok {
		return nil, domain.ErrNotFound
	}
	clone := *p
	return &clone, nil
}

// InsertIfNew writes the prospect only when no record exists for the URL.
func (r *FakeProspectRepository) InsertIfNew(_ context.Context, p *domain.Prospect) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.records[p.ProfileURL]; exists {
		return false, nil
	}
	clone := *p
	r.records[p.ProfileURL] = &clone
	return true, nil
}

// ListByState returns prospects in state whose NextActionAt is at or before dueBy.
func (r *FakeProspectRepository) ListByState(_ context.Context, state domain.State, dueBy time.Time) ([]*domain.Prospect, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Prospect
	for _, p := range r.records {
		if p.State == state && !p.NextActionAt.After(dueBy) {
			clone := *p
			out = append(out, &clone)
		}
	}
	return out, nil
}

// ListByStateOrderedByScore returns up to limit prospects in state, sorted by
// WorthinessScore descending.
func (r *FakeProspectRepository) ListByStateOrderedByScore(_ context.Context, state domain.State, limit int) ([]*domain.Prospect, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []*domain.Prospect
	for _, p := range r.records {
		if p.State == state {
			clone := *p
			out = append(out, &clone)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].WorthinessScore > out[j].WorthinessScore
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

// ── FakeRateLimiter ─────────────────────────────────────────────────────────

// FakeRateLimiter enforces a configurable per-scope daily cap in memory.
type FakeRateLimiter struct {
	mu     sync.Mutex
	max    int
	counts map[string]int
}

// NewFakeRateLimiter returns a FakeRateLimiter that allows max acquisitions per scope.
func NewFakeRateLimiter(max int) *FakeRateLimiter {
	return &FakeRateLimiter{max: max, counts: make(map[string]int)}
}

// Ensure compile-time interface satisfaction.
var _ domain.RateLimiter = (*FakeRateLimiter)(nil)

// Acquire increments the counter for scope or returns ErrRateLimitExceeded.
func (rl *FakeRateLimiter) Acquire(_ context.Context, scope string) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	if rl.counts[scope] >= rl.max {
		return domain.ErrRateLimitExceeded
	}
	rl.counts[scope]++
	return nil
}

// Current returns the number of acquisitions so far for scope.
func (rl *FakeRateLimiter) Current(_ context.Context, scope string) (int, error) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return rl.counts[scope], nil
}

// ── FakeLLMClient ────────────────────────────────────────────────────────────

// FakeLLMClient returns configurable canned responses for LLM calls.
type FakeLLMClient struct {
	ScoreResult domain.ScoreResult
	CritiqueVal int
	Err         error
}

// Ensure compile-time interface satisfaction.
var _ domain.LLMClient = (*FakeLLMClient)(nil)

// ScoreAndDraft returns the configured ScoreResult or Err.
func (f *FakeLLMClient) ScoreAndDraft(_ context.Context, _ *domain.Prospect, _ []domain.Prospect) (domain.ScoreResult, error) {
	return f.ScoreResult, f.Err
}

// Critique returns the configured CritiqueVal or Err.
func (f *FakeLLMClient) Critique(_ context.Context, _ string) (int, error) {
	return f.CritiqueVal, f.Err
}

// ── FakeBrowserClient ────────────────────────────────────────────────────────

// FakeBrowserClient records calls and returns configurable responses.
type FakeBrowserClient struct {
	ProfileDataByURL map[string]domain.ProfileData
	SearchURLs       []string
	Block            domain.BlockType
	Err              error

	VisitedURLs []string
	LikedURLs   []string
}

// Ensure compile-time interface satisfaction.
var _ domain.BrowserClient = (*FakeBrowserClient)(nil)

// VisitProfile returns the configured ProfileData for the URL.
func (f *FakeBrowserClient) VisitProfile(_ context.Context, profileURL string) (domain.ProfileData, error) {
	f.VisitedURLs = append(f.VisitedURLs, profileURL)
	return f.ProfileDataByURL[profileURL], f.Err
}

// LikeRecentPost records the liked URL.
func (f *FakeBrowserClient) LikeRecentPost(_ context.Context, profileURL string) error {
	f.LikedURLs = append(f.LikedURLs, profileURL)
	return f.Err
}

// SearchProfiles returns the configured URL list.
func (f *FakeBrowserClient) SearchProfiles(_ context.Context, _, _ string, _ int) ([]string, error) {
	return f.SearchURLs, f.Err
}

// CheckBlock returns the configured BlockType.
func (f *FakeBrowserClient) CheckBlock(_ context.Context) (domain.BlockType, error) {
	return f.Block, f.Err
}

// Close is a no-op.
func (f *FakeBrowserClient) Close() error { return nil }

// ── FakeEmbeddingClient ──────────────────────────────────────────────────────

// FakeEmbeddingClient returns a fixed vector for every input.
type FakeEmbeddingClient struct {
	Vector []float32
	Err    error
}

// Ensure compile-time interface satisfaction.
var _ domain.EmbeddingClient = (*FakeEmbeddingClient)(nil)

// Embed returns the configured vector.
func (f *FakeEmbeddingClient) Embed(_ context.Context, _ string) ([]float32, error) {
	return f.Vector, f.Err
}

// ── FakeEmbeddingStore ───────────────────────────────────────────────────────

// FakeEmbeddingStore is an in-memory embedding store.
type FakeEmbeddingStore struct {
	mu      sync.Mutex
	points  map[string]fakePoint
	Results []*domain.Prospect
	Err     error
}

type fakePoint struct {
	vector  []float32
	payload map[string]any
}

// NewFakeEmbeddingStore returns an initialised FakeEmbeddingStore.
func NewFakeEmbeddingStore() *FakeEmbeddingStore {
	return &FakeEmbeddingStore{points: make(map[string]fakePoint)}
}

// Ensure compile-time interface satisfaction.
var _ domain.EmbeddingStore = (*FakeEmbeddingStore)(nil)

// Upsert stores the vector and payload for id.
func (s *FakeEmbeddingStore) Upsert(_ context.Context, id string, vector []float32, payload map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.Err != nil {
		return s.Err
	}
	s.points[id] = fakePoint{vector: vector, payload: payload}
	return nil
}

// SearchSimilar returns the pre-configured Results slice.
func (s *FakeEmbeddingStore) SearchSimilar(_ context.Context, _ []float32, _ map[string]any, topK int) ([]*domain.Prospect, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	out := s.Results
	if topK > 0 && len(out) > topK {
		out = out[:topK]
	}
	return out, nil
}

// UpsertedIDs returns the set of IDs that have been upserted, useful in tests.
func (s *FakeEmbeddingStore) UpsertedIDs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, 0, len(s.points))
	for id := range s.points {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}

