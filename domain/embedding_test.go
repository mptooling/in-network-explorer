package domain_test

import (
	"context"

	"github.com/pavlomaksymov/in-network-explorer/domain"
)

// Compile-time checks: fakeEmbeddingClient and fakeEmbeddingStore must satisfy
// the respective interfaces.

var _ domain.EmbeddingClient = (*fakeEmbeddingClient)(nil)
var _ domain.EmbeddingStore = (*fakeEmbeddingStore)(nil)

type fakeEmbeddingClient struct{}

func (f *fakeEmbeddingClient) Embed(_ context.Context, _ string) ([]float32, error) {
	return nil, nil
}

type fakeEmbeddingStore struct{}

func (f *fakeEmbeddingStore) Upsert(_ context.Context, _ string, _ []float32, _ map[string]any) error {
	return nil
}
func (f *fakeEmbeddingStore) SearchSimilar(_ context.Context, _ []float32, _ map[string]any, _ int) ([]*domain.Prospect, error) {
	return nil, nil
}
