package explorer_test

import (
	"context"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

// Compile-time checks: fakeEmbeddingClient and fakeEmbeddingStore must satisfy
// the respective interfaces.

var _ explorer.EmbeddingClient = (*fakeEmbeddingClient)(nil)
var _ explorer.EmbeddingStore = (*fakeEmbeddingStore)(nil)

type fakeEmbeddingClient struct{}

func (f *fakeEmbeddingClient) Embed(_ context.Context, _ string) ([]float32, error) {
	return nil, nil
}

type fakeEmbeddingStore struct{}

func (f *fakeEmbeddingStore) Upsert(_ context.Context, _ string, _ []float32, _ map[string]any) error {
	return nil
}
func (f *fakeEmbeddingStore) SearchSimilar(_ context.Context, _ []float32, _ map[string]any, _ int) ([]*explorer.Prospect, error) {
	return nil, nil
}
