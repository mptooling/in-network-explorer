package report_test

import (
	"bytes"
	"testing"

	"github.com/pavlomaksymov/in-network-explorer/internal/report"
)

func TestSeedReport_NotEmpty(t *testing.T) {
	t.Parallel()
	r := report.SeedReport()
	if r == nil {
		t.Fatal("SeedReport returned nil")
	}
	if len(r.Prospects) == 0 {
		t.Fatal("SeedReport returned zero prospects")
	}
}

func TestSeedReport_AllFieldsPopulated(t *testing.T) {
	t.Parallel()
	for _, item := range report.SeedReport().Prospects {
		if item.ProfileURL == "" {
			t.Errorf("prospect %q: empty ProfileURL", item.Name)
		}
		if item.Name == "" {
			t.Error("prospect has empty Name")
		}
		if item.Headline == "" {
			t.Errorf("prospect %q: empty Headline", item.Name)
		}
		if item.Location == "" {
			t.Errorf("prospect %q: empty Location", item.Name)
		}
		if item.WorthinessScore == 0 {
			t.Errorf("prospect %q: zero WorthinessScore", item.Name)
		}
		if item.DraftedMessage == "" {
			t.Errorf("prospect %q: empty DraftedMessage", item.Name)
		}
		if item.ScannedAt.IsZero() {
			t.Errorf("prospect %q: zero ScannedAt", item.Name)
		}
	}
}

func TestSeedReport_ScoreRange(t *testing.T) {
	t.Parallel()
	for _, item := range report.SeedReport().Prospects {
		if item.WorthinessScore < 1 || item.WorthinessScore > 10 {
			t.Errorf("prospect %q: score %d out of range [1,10]", item.Name, item.WorthinessScore)
		}
	}
}

func TestSeedReport_RendersHTML(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	if err := report.RenderHTML(&buf, report.SeedReport()); err != nil {
		t.Fatalf("RenderHTML with seed data: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("empty HTML output from seed data")
	}
}
