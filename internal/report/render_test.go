package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

func sampleEntries() []explorer.ProspectReport {
	return []explorer.ProspectReport{
		{
			ProfileURL:      "https://linkedin.com/in/alice",
			Name:            "Alice",
			Headline:        "Staff Engineer",
			Location:        "Berlin",
			WorthinessScore: 9,
			ScoreReasoning:  "Great match",
			DraftedMessage:  "Hi Alice, I saw your talk on distributed systems.",
			CritiqueScore:   13,
		},
		{
			ProfileURL:      "https://linkedin.com/in/bob",
			Name:            "Bob",
			Headline:        "Tech Lead",
			Location:        "Berlin",
			WorthinessScore: 7,
			DraftedMessage:  "Hi Bob, your work on observability caught my eye.",
			CritiqueScore:   10,
		},
	}
}

func TestWriteJSON(t *testing.T) {
	r := New(sampleEntries())
	var buf bytes.Buffer
	if err := r.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}

	var decoded Report
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded.Entries) != 2 {
		t.Fatalf("entries = %d, want 2", len(decoded.Entries))
	}
	if decoded.Entries[0].Name != "Alice" {
		t.Errorf("first entry name = %q, want Alice", decoded.Entries[0].Name)
	}
}

func TestWriteHTML(t *testing.T) {
	r := New(sampleEntries())
	var buf bytes.Buffer
	if err := r.WriteHTML(&buf); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}

	html := buf.String()
	if !strings.Contains(html, "Alice") {
		t.Error("HTML missing Alice")
	}
	if !strings.Contains(html, "linkedin.com/in/alice") {
		t.Error("HTML missing profile link")
	}
	if !strings.Contains(html, "Score: 9/10") {
		t.Error("HTML missing score")
	}
	if !strings.Contains(html, "copy-btn") {
		t.Error("HTML missing copy button")
	}
	if !strings.Contains(html, "2 prospects") {
		t.Error("HTML missing prospect count")
	}
}

func TestWriteHTML_Empty(t *testing.T) {
	r := New(nil)
	var buf bytes.Buffer
	if err := r.WriteHTML(&buf); err != nil {
		t.Fatalf("WriteHTML: %v", err)
	}
	if !strings.Contains(buf.String(), "No drafted prospects") {
		t.Error("HTML missing empty state message")
	}
}

func TestWriteJSON_Empty(t *testing.T) {
	r := New(nil)
	var buf bytes.Buffer
	if err := r.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	if !strings.Contains(buf.String(), `"entries"`) {
		t.Error("JSON missing entries key")
	}
}
