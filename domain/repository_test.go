package domain_test

import (
	"context"
	"time"

	"github.com/pavlomaksymov/in-network-explorer/domain"
)

// Compile-time checks: fakeRepo and fakeRateLimiter must satisfy the interfaces.

var _ domain.ProspectRepository = (*fakeRepo)(nil)
var _ domain.RateLimiter = (*fakeRateLimiter)(nil)

type fakeRepo struct{}

func (f *fakeRepo) Save(_ context.Context, _ *domain.Prospect) error { return nil }
func (f *fakeRepo) Get(_ context.Context, _ string) (*domain.Prospect, error) {
	return nil, nil
}
func (f *fakeRepo) InsertIfNew(_ context.Context, _ *domain.Prospect) (bool, error) {
	return false, nil
}
func (f *fakeRepo) ListByState(_ context.Context, _ domain.State, _ time.Time) ([]*domain.Prospect, error) {
	return nil, nil
}
func (f *fakeRepo) ListByStateOrderedByScore(_ context.Context, _ domain.State, _ int) ([]*domain.Prospect, error) {
	return nil, nil
}

type fakeRateLimiter struct{}

func (f *fakeRateLimiter) Acquire(_ context.Context, _ string) error { return nil }
func (f *fakeRateLimiter) Current(_ context.Context, _ string) (int, error) {
	return 0, nil
}
