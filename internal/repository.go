package explorer

import (
	"context"
	"time"
)

// ProspectRepository is the persistence abstraction for Prospect aggregates.
// Implementations live in internal/dynamo.
type ProspectRepository interface {
	// Save creates or fully replaces the prospect record.
	Save(ctx context.Context, p *Prospect) error

	// Get returns the prospect for profileURL or ErrNotFound.
	Get(ctx context.Context, profileURL string) (*Prospect, error)

	// InsertIfNew writes the prospect only when no record exists for the URL.
	// Returns true if a new record was created, false if it already existed.
	InsertIfNew(ctx context.Context, p *Prospect) (inserted bool, err error)

	// ListByState returns all prospects in the given state whose NextActionAt
	// is at or before dueBy.
	ListByState(ctx context.Context, state State, dueBy time.Time) ([]*Prospect, error)

	// ListByStateOrderedByScore returns up to limit prospects in the given state,
	// ordered by WorthinessScore descending.
	ListByStateOrderedByScore(ctx context.Context, state State, limit int) ([]*Prospect, error)
}

// RateLimiter enforces daily action caps per scope (e.g. "profile_views",
// "connection_requests"). Implementations live in internal/dynamo.
type RateLimiter interface {
	// Acquire increments the counter for scope. Returns ErrRateLimitExceeded
	// if the daily cap has already been reached.
	Acquire(ctx context.Context, scope string) error

	// Current returns the number of actions already taken today for scope.
	Current(ctx context.Context, scope string) (int, error)
}
