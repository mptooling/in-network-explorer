package report_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
	"github.com/pavlomaksymov/in-network-explorer/internal/report"
)

func sampleReport() *explorer.ProspectReport {
	return &explorer.ProspectReport{
		GeneratedAt: time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC),
		Prospects: []explorer.ReportItem{
			{
				ProfileURL:      "https://linkedin.com/in/anna-mueller",
				Name:            "Anna Müller",
				Headline:        "Senior Platform Engineer at Delivery Hero",
				Location:        "Berlin, Germany",
				WorthinessScore: 9,
				CalibratedProb:  0.85,
				DraftedMessage:  "Hi Anna, I noticed your post about platform engineering at Delivery Hero — really resonated with my experience scaling microservices in Berlin.",
				RecentPost:      "Just shipped our new observability stack!",
				ScannedAt:       time.Date(2026, 4, 8, 14, 30, 0, 0, time.UTC),
			},
			{
				ProfileURL:      "https://linkedin.com/in/bob-schmidt",
				Name:            "Bob Schmidt",
				Headline:        "DevOps Lead at N26",
				Location:        "Berlin, Germany",
				WorthinessScore: 5,
				CalibratedProb:  0.40,
				DraftedMessage:  "Hey Bob, your take on GitOps workflows caught my eye.",
				RecentPost:      "Terraform tips for financial services",
				ScannedAt:       time.Date(2026, 4, 7, 10, 0, 0, 0, time.UTC),
			},
			{
				ProfileURL:      "https://linkedin.com/in/clara-weber",
				Name:            "Clara Weber",
				Headline:        "Junior QA Engineer",
				Location:        "Berlin",
				WorthinessScore: 3,
				CalibratedProb:  0.12,
				DraftedMessage:  "Hi Clara, saw your post about testing.",
				RecentPost:      "",
				ScannedAt:       time.Date(2026, 4, 6, 8, 0, 0, 0, time.UTC),
			},
		},
	}
}

// ── JSON tests ──────────────────────────────────────────────────────────────

func TestRenderJSON_ValidJSON(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := report.RenderJSON(&buf, sampleReport()); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	if !json.Valid(buf.Bytes()) {
		t.Fatal("output is not valid JSON")
	}
}

func TestRenderJSON_ContainsAllProspects(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := report.RenderJSON(&buf, sampleReport()); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}

	var decoded explorer.ProspectReport
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Prospects) != 3 {
		t.Errorf("prospects count: got %d, want 3", len(decoded.Prospects))
	}
}

func TestRenderJSON_SnakeCaseFields(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := report.RenderJSON(&buf, sampleReport()); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}

	out := buf.String()
	for _, key := range []string{"generated_at", "profile_url", "worthiness_score", "drafted_message"} {
		if !strings.Contains(out, key) {
			t.Errorf("missing snake_case key %q in JSON output", key)
		}
	}
}

// ── HTML tests ──────────────────────────────────────────────────────────────

func TestRenderHTML_ValidHTML(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := report.RenderHTML(&buf, sampleReport()); err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "<!DOCTYPE html>") {
		t.Error("output should contain DOCTYPE")
	}
	if !strings.Contains(out, "</html>") {
		t.Error("output should contain closing </html> tag")
	}
}

func TestRenderHTML_ProfileLinks(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	r := sampleReport()
	if err := report.RenderHTML(&buf, r); err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}

	out := buf.String()
	for _, item := range r.Prospects {
		if !strings.Contains(out, item.ProfileURL) {
			t.Errorf("missing profile link for %s", item.Name)
		}
	}
}

func TestRenderHTML_ShowsScore(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := report.RenderHTML(&buf, sampleReport()); err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "9/10") {
		t.Error("should display score 9/10")
	}
	if !strings.Contains(out, "5/10") {
		t.Error("should display score 5/10")
	}
}

func TestRenderHTML_ScoreColor(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := report.RenderHTML(&buf, sampleReport()); err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "score-high") {
		t.Error("score >= 8 should have class score-high")
	}
	if !strings.Contains(out, "score-mid") {
		t.Error("score 5-7 should have class score-mid")
	}
	if !strings.Contains(out, "score-low") {
		t.Error("score <= 4 should have class score-low")
	}
}

func TestRenderHTML_HasCopyButtons(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	r := sampleReport()
	if err := report.RenderHTML(&buf, r); err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}

	out := buf.String()
	for _, item := range r.Prospects {
		if item.DraftedMessage == "" {
			continue
		}
		if !strings.Contains(out, "Copy") {
			t.Errorf("missing copy button for %s", item.Name)
		}
	}
	if !strings.Contains(out, "clipboard") {
		t.Error("should use clipboard API for copy functionality")
	}
}

func TestRenderHTML_EscapesSpecialChars(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := report.RenderHTML(&buf, sampleReport()); err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}

	out := buf.String()
	// Anna Müller — template should properly render unicode
	if !strings.Contains(out, "Müller") {
		t.Error("should render unicode characters (Müller)")
	}
}

func TestRenderHTML_EmptyProspects(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	empty := &explorer.ProspectReport{
		GeneratedAt: time.Now(),
		Prospects:   nil,
	}
	if err := report.RenderHTML(&buf, empty); err != nil {
		t.Fatalf("RenderHTML with empty prospects: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("should produce output even with no prospects")
	}
}
