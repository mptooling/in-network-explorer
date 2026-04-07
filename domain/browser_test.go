package domain_test

import (
	"context"

	"github.com/pavlomaksymov/in-network-explorer/domain"
)

// Compile-time check: fakeBrowser must satisfy BrowserClient.

var _ domain.BrowserClient = (*fakeBrowser)(nil)

type fakeBrowser struct{}

func (f *fakeBrowser) VisitProfile(_ context.Context, _ string) (domain.ProfileData, error) {
	return domain.ProfileData{}, nil
}
func (f *fakeBrowser) LikeRecentPost(_ context.Context, _ string) error { return nil }
func (f *fakeBrowser) SearchProfiles(_ context.Context, _, _ string, _ int) ([]string, error) {
	return nil, nil
}
func (f *fakeBrowser) CheckBlock(_ context.Context) (domain.BlockType, error) {
	return domain.BlockNone, nil
}
func (f *fakeBrowser) Close() error { return nil }
