package qdrant

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpsert_SendsCorrectPayload(t *testing.T) {
	var received upsertRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("method = %s, want PUT", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	store := &Store{baseURL: srv.URL, collection: "prospects", client: http.DefaultClient}
	err := store.Upsert(context.Background(), "point-1", []float32{0.1, 0.2, 0.3}, map[string]any{"name": "Alice"})
	if err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	if len(received.Points) != 1 {
		t.Fatalf("points = %d, want 1", len(received.Points))
	}
	if received.Points[0].ID != "point-1" {
		t.Errorf("ID = %q, want point-1", received.Points[0].ID)
	}
	if len(received.Points[0].Vector) != 3 {
		t.Errorf("vector len = %d, want 3", len(received.Points[0].Vector))
	}
}

func TestSearchSimilar_ParsesResults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := searchResponse{
			Result: []searchHit{
				{
					ID:    "p1",
					Score: 0.95,
					Payload: map[string]any{
						"profile_url":      "https://linkedin.com/in/alice",
						"name":             "Alice",
						"headline":         "Engineer",
						"location":         "Berlin",
						"worthiness_score": float64(8),
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	store := &Store{baseURL: srv.URL, collection: "prospects", client: http.DefaultClient}
	results, err := store.SearchSimilar(context.Background(), []float32{0.1, 0.2}, nil, 5)
	if err != nil {
		t.Fatalf("SearchSimilar: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	if results[0].Name != "Alice" {
		t.Errorf("Name = %q, want Alice", results[0].Name)
	}
	if results[0].WorthinessScore != 8 {
		t.Errorf("Score = %d, want 8", results[0].WorthinessScore)
	}
}

func TestSearchSimilar_WithFilter(t *testing.T) {
	var receivedBody searchRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&receivedBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(searchResponse{})
	}))
	defer srv.Close()

	store := &Store{baseURL: srv.URL, collection: "prospects", client: http.DefaultClient}
	_, err := store.SearchSimilar(context.Background(), []float32{0.1}, map[string]any{"location": "Berlin"}, 3)
	if err != nil {
		t.Fatalf("SearchSimilar: %v", err)
	}
	if receivedBody.Filter == nil {
		t.Error("expected filter in request")
	}
	if receivedBody.Limit != 3 {
		t.Errorf("limit = %d, want 3", receivedBody.Limit)
	}
}

func TestBuildFilter(t *testing.T) {
	f := buildFilter(map[string]any{"location": "Berlin"})
	must, ok := f["must"].([]map[string]any)
	if !ok || len(must) != 1 {
		t.Fatalf("filter must = %v", f)
	}
	if must[0]["key"] != "location" {
		t.Errorf("key = %v, want location", must[0]["key"])
	}
}

func TestPayloadToProspect(t *testing.T) {
	p := payloadToProspect(map[string]any{
		"profile_url":      "https://linkedin.com/in/bob",
		"name":             "Bob",
		"headline":         "Lead",
		"location":         "Berlin",
		"worthiness_score": float64(7),
	})
	if p.ProfileURL != "https://linkedin.com/in/bob" {
		t.Errorf("ProfileURL = %q", p.ProfileURL)
	}
	if p.WorthinessScore != 7 {
		t.Errorf("Score = %d, want 7", p.WorthinessScore)
	}
}

func TestUpsert_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	store := &Store{baseURL: srv.URL, collection: "test", client: http.DefaultClient}
	err := store.Upsert(context.Background(), "x", []float32{0.1}, nil)
	if err == nil {
		t.Fatal("expected error on 500")
	}
}
