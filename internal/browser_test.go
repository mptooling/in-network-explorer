package explorer_test

import (
	"context"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

// Compile-time check: fakeBrowser must satisfy BrowserClient.

var _ explorer.BrowserClient = (*fakeBrowser)(nil)

type fakeBrowser struct{}

func (f *fakeBrowser) VisitProfile(_ context.Context, _ string) (explorer.ProfileData, error) {
	return explorer.ProfileData{}, nil
}
func (f *fakeBrowser) LikeRecentPost(_ context.Context, _ string) error { return nil }
func (f *fakeBrowser) SearchProfiles(_ context.Context, _, _ string, _ int) ([]string, error) {
	return nil, nil
}
func (f *fakeBrowser) CheckBlock(_ context.Context) (explorer.BlockType, error) {
	return explorer.BlockNone, nil
}
func (f *fakeBrowser) Close() error { return nil }
