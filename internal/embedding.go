package explorer

import "context"

// EmbeddingClient converts text to a dense vector representation.
// Implementations live in internal/bedrock.
type EmbeddingClient interface {
	// Embed returns the embedding vector for the given text.
	Embed(ctx context.Context, text string) ([]float32, error)
}

// EmbeddingStore is the vector database abstraction for semantic search over
// prospect profiles. Implementations live in internal/qdrant.
type EmbeddingStore interface {
	// Upsert inserts or updates a vector point identified by id with the given
	// payload metadata.
	Upsert(ctx context.Context, id string, vector []float32, payload map[string]any) error

	// SearchSimilar returns the top-K prospects nearest to vector that also
	// match the filter predicates.
	SearchSimilar(ctx context.Context, vector []float32, filter map[string]any, topK int) ([]*Prospect, error)
}
