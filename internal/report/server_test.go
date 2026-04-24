package report

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testServer() *httptest.Server {
	r := New(sampleEntries())
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	s := NewServer(r, log)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleHTML)
	mux.HandleFunc("GET /api/report", s.handleJSON)
	return httptest.NewServer(mux)
}

func TestServer_HTMLEndpoint(t *testing.T) {
	srv := testServer()
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Alice") {
		t.Error("HTML missing Alice")
	}
}

func TestServer_JSONEndpoint(t *testing.T) {
	srv := testServer()
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL+"/api/report", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var r Report
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(r.Entries) != 2 {
		t.Errorf("entries = %d, want 2", len(r.Entries))
	}
}
