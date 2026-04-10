package report_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/report"
)

func testReport() *explorer.ProspectReport {
	return &explorer.ProspectReport{
		GeneratedAt: time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC),
		Prospects: []explorer.ReportItem{
			{
				ProfileURL:      "https://linkedin.com/in/test-user",
				Name:            "Test User",
				Headline:        "Engineer",
				Location:        "Berlin",
				WorthinessScore: 8,
				CalibratedProb:  0.75,
				DraftedMessage:  "Hello Test!",
				RecentPost:      "Great post",
				ScannedAt:       time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC),
			},
		},
	}
}

func get(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", url, err)
	}
	return resp
}

func TestPreviewHandler_HTMLEndpoint(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := report.NewPreviewHandler(testReport(), log)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp := get(t, srv.URL+"/")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type: got %q, want text/html", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Test User") {
		t.Error("HTML response should contain prospect name")
	}
}

func TestPreviewHandler_JSONEndpoint(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := report.NewPreviewHandler(testReport(), log)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp := get(t, srv.URL+"/json")
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want 200", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("content-type: got %q, want application/json", ct)
	}

	var decoded explorer.ProspectReport
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		t.Fatalf("decode JSON: %v", err)
	}
	if len(decoded.Prospects) != 1 {
		t.Errorf("prospect count: got %d, want 1", len(decoded.Prospects))
	}
}

func TestPreviewServer_ShutdownOnContextCancel(t *testing.T) {
	t.Parallel()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	r := testReport()

	ctx, cancel := context.WithCancel(context.Background())

	srv := report.NewPreviewServer(r, "127.0.0.1:0", log)

	done := make(chan error, 1)
	go func() { done <- srv.ListenAndServe(ctx) }()

	// Give server time to start, then cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Server stopped — success.
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down within 2s after context cancel")
	}
}
