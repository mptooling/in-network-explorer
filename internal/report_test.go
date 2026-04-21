package explorer_test

import (
	"encoding/json"
	"testing"
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

func TestReportItem_JSONFieldNames(t *testing.T) {
	t.Parallel()

	item := explorer.ReportItem{
		ProfileURL:      "https://linkedin.com/in/test",
		Name:            "Test User",
		Headline:        "Engineer",
		Location:        "Berlin",
		WorthinessScore: 8,
		CalibratedProb:  0.75,
		DraftedMessage:  "Hello!",
		RecentPost:      "Exciting news",
		ScannedAt:       time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	wantKeys := []string{
		"profile_url",
		"name",
		"headline",
		"location",
		"worthiness_score",
		"calibrated_prob",
		"drafted_message",
		"recent_post",
		"scanned_at",
	}

	for _, key := range wantKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("missing JSON key %q, got keys: %v", key, keys(raw))
		}
	}

	if len(raw) != len(wantKeys) {
		t.Errorf("unexpected key count: got %d, want %d", len(raw), len(wantKeys))
	}
}

func TestProspectReport_JSONRoundTrip(t *testing.T) {
	t.Parallel()

	report := explorer.ProspectReport{
		GeneratedAt: time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC),
		Prospects: []explorer.ReportItem{
			{
				ProfileURL:      "https://linkedin.com/in/alice",
				Name:            "Alice",
				WorthinessScore: 9,
			},
			{
				ProfileURL:      "https://linkedin.com/in/bob",
				Name:            "Bob",
				WorthinessScore: 4,
			},
		},
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded explorer.ProspectReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if len(decoded.Prospects) != 2 {
		t.Fatalf("prospects count: got %d, want 2", len(decoded.Prospects))
	}
	if decoded.Prospects[0].Name != "Alice" {
		t.Errorf("first prospect name: got %q, want %q", decoded.Prospects[0].Name, "Alice")
	}
	if decoded.GeneratedAt.IsZero() {
		t.Error("generated_at should not be zero")
	}
}

func keys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
