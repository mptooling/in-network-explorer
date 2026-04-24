// Package qdrant implements the EmbeddingStore interface using Qdrant's
// REST API. No third-party client library — uses net/http from stdlib.
package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

var _ explorer.EmbeddingStore = (*Store)(nil)

// Store persists and searches prospect embeddings in Qdrant.
type Store struct {
	baseURL    string // e.g. "http://localhost:6333"
	collection string
	client     *http.Client
}

// NewStore creates a Qdrant embedding store.
func NewStore(restAddr, collection string) *Store {
	return &Store{
		baseURL:    "http://" + restAddr,
		collection: collection,
		client:     &http.Client{},
	}
}

// Upsert inserts or updates a vector point.
func (s *Store) Upsert(ctx context.Context, id string, vector []float32, payload map[string]any) error {
	body := upsertRequest{
		Points: []point{{
			ID:      id,
			Vector:  vector,
			Payload: payload,
		}},
	}
	return s.put(ctx, fmt.Sprintf("/collections/%s/points", s.collection), body)
}

// SearchSimilar returns the top-K prospects nearest to vector that match the
// filter predicates. Filter keys map to Qdrant payload field conditions.
func (s *Store) SearchSimilar(ctx context.Context, vector []float32, filter map[string]any, topK int) ([]*explorer.Prospect, error) {
	req := searchRequest{
		Vector:      vector,
		Limit:       topK,
		WithPayload: true,
	}
	if len(filter) > 0 {
		req.Filter = buildFilter(filter)
	}

	var resp searchResponse
	if err := s.post(ctx, fmt.Sprintf("/collections/%s/points/search", s.collection), req, &resp); err != nil {
		return nil, err
	}

	prospects := make([]*explorer.Prospect, 0, len(resp.Result))
	for _, hit := range resp.Result {
		p := payloadToProspect(hit.Payload)
		prospects = append(prospects, p)
	}
	return prospects, nil
}

// ── HTTP helpers ────────────────────────────────────────────────────────────

func (s *Store) put(ctx context.Context, path string, body any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, s.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("qdrant PUT %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qdrant PUT %s: %d %s", path, resp.StatusCode, string(b))
	}
	return nil
}

func (s *Store) post(ctx context.Context, path string, body any, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("qdrant POST %s: %w", path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qdrant POST %s: %d %s", path, resp.StatusCode, string(b))
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

// ── request/response types ──────────────────────────────────────────────────

type point struct {
	ID      string         `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload,omitempty"`
}

type upsertRequest struct {
	Points []point `json:"points"`
}

type searchRequest struct {
	Vector      []float32 `json:"vector"`
	Limit       int       `json:"limit"`
	WithPayload bool      `json:"with_payload"`
	Filter      any       `json:"filter,omitempty"`
}

type searchResponse struct {
	Result []searchHit `json:"result"`
}

type searchHit struct {
	ID      string         `json:"id"`
	Score   float64        `json:"score"`
	Payload map[string]any `json:"payload"`
}

// buildFilter converts a flat key→value map into Qdrant's filter format.
func buildFilter(kv map[string]any) map[string]any {
	must := make([]map[string]any, 0, len(kv))
	for k, v := range kv {
		must = append(must, map[string]any{
			"key":   k,
			"match": map[string]any{"value": v},
		})
	}
	return map[string]any{"must": must}
}

func payloadToProspect(payload map[string]any) *explorer.Prospect {
	str := func(key string) string {
		v, _ := payload[key].(string)
		return v
	}
	num := func(key string) int {
		v, _ := payload[key].(float64)
		return int(v)
	}
	return &explorer.Prospect{
		ProfileURL:      str("profile_url"),
		Name:            str("name"),
		Headline:        str("headline"),
		Location:        str("location"),
		WorthinessScore: num("worthiness_score"),
	}
}
