package explorer_test

import (
	"context"
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

// Compile-time checks: fakeRepo and fakeRateLimiter must satisfy the interfaces.

var _ explorer.ProspectRepository = (*fakeRepo)(nil)
var _ explorer.RateLimiter = (*fakeRateLimiter)(nil)

type fakeRepo struct{}

func (f *fakeRepo) Save(_ context.Context, _ *explorer.Prospect) error { return nil }
func (f *fakeRepo) Get(_ context.Context, _ string) (*explorer.Prospect, error) {
	return nil, nil
}
func (f *fakeRepo) InsertIfNew(_ context.Context, _ *explorer.Prospect) (bool, error) {
	return false, nil
}
func (f *fakeRepo) ListByState(_ context.Context, _ explorer.State, _ time.Time) ([]*explorer.Prospect, error) {
	return nil, nil
}
func (f *fakeRepo) ListByStateOrderedByScore(_ context.Context, _ explorer.State, _ int) ([]*explorer.Prospect, error) {
	return nil, nil
}

type fakeRateLimiter struct{}

func (f *fakeRateLimiter) Acquire(_ context.Context, _ string) error { return nil }
func (f *fakeRateLimiter) Current(_ context.Context, _ string) (int, error) {
	return 0, nil
}
