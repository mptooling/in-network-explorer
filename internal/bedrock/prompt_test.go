package bedrock

import (
	"strings"
	"testing"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

func TestBuildScorePrompt_ContainsProspectData(t *testing.T) {
	p := &explorer.Prospect{
		Name:        "Alice Smith",
		Headline:    "Staff Engineer",
		Location:    "Berlin",
		About:       "Distributed systems expert",
		RecentPosts: []string{"Excited about Kubernetes 1.30"},
	}
	prompt := buildScorePrompt(p, nil)

	for _, want := range []string{"Alice Smith", "Staff Engineer", "Berlin", "Distributed systems", "Kubernetes"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}

func TestBuildScorePrompt_IncludesExamples(t *testing.T) {
	p := &explorer.Prospect{Name: "Target", Location: "Berlin"}
	examples := []explorer.Prospect{
		{Name: "Example", WorthinessScore: 9, DraftedMessage: "Hi there"},
	}
	prompt := buildScorePrompt(p, examples)

	if !strings.Contains(prompt, "Example") {
		t.Error("prompt missing example name")
	}
	if !strings.Contains(prompt, "score: 9") {
		t.Error("prompt missing example score")
	}
}

func TestBuildCritiquePrompt_ContainsMessage(t *testing.T) {
	prompt := buildCritiquePrompt("Hi Alice, love your work on distributed systems!")
	if !strings.Contains(prompt, "distributed systems") {
		t.Error("prompt missing message content")
	}
}

func TestParseScoreResponse_ValidJSON(t *testing.T) {
	raw := `{"score":8,"reasoning":"Great Berlin engineer","message":"Hi Alice!","critique":{"specificity":4,"relevance":5,"tone":4}}`
	result, err := parseScoreResponse(raw)
	if err != nil {
		t.Fatalf("parseScoreResponse: %v", err)
	}
	if result.Score != 8 {
		t.Errorf("Score = %d, want 8", result.Score)
	}
	if result.Reasoning != "Great Berlin engineer" {
		t.Errorf("Reasoning = %q", result.Reasoning)
	}
	if result.Message != "Hi Alice!" {
		t.Errorf("Message = %q", result.Message)
	}
	if result.CritiqueScore != 13 {
		t.Errorf("CritiqueScore = %d, want 13", result.CritiqueScore)
	}
}

func TestParseScoreResponse_WithCodeFences(t *testing.T) {
	raw := "```json\n{\"score\":7,\"reasoning\":\"Good fit\",\"message\":\"Hello\",\"critique\":{\"specificity\":3,\"relevance\":4,\"tone\":3}}\n```"
	result, err := parseScoreResponse(raw)
	if err != nil {
		t.Fatalf("parseScoreResponse: %v", err)
	}
	if result.Score != 7 {
		t.Errorf("Score = %d, want 7", result.Score)
	}
}

func TestParseScoreResponse_ClampsValues(t *testing.T) {
	raw := `{"score":15,"reasoning":"x","message":"m","critique":{"specificity":10,"relevance":10,"tone":10}}`
	result, err := parseScoreResponse(raw)
	if err != nil {
		t.Fatalf("parseScoreResponse: %v", err)
	}
	if result.Score != 10 {
		t.Errorf("Score = %d, want 10 (clamped)", result.Score)
	}
	if result.CritiqueScore != 15 {
		t.Errorf("CritiqueScore = %d, want 15 (clamped)", result.CritiqueScore)
	}
}

func TestParseCritiqueResponse(t *testing.T) {
	raw := `{"specificity":4,"relevance":5,"tone":3}`
	score, err := parseCritiqueResponse(raw)
	if err != nil {
		t.Fatalf("parseCritiqueResponse: %v", err)
	}
	if score != 12 {
		t.Errorf("score = %d, want 12", score)
	}
}

func TestStripCodeFences(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{`{"a":1}`, `{"a":1}`},
		{"```json\n{\"a\":1}\n```", `{"a":1}`},
		{"```\n{\"a\":1}\n```", `{"a":1}`},
		{"  ```json\n{\"a\":1}\n```  ", `{"a":1}`},
	}
	for _, tc := range cases {
		got := stripCodeFences(tc.input)
		if got != tc.want {
			t.Errorf("stripCodeFences(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 10); got != "hello" {
		t.Errorf("truncate short = %q", got)
	}
	if got := truncate("hello world", 5); got != "hello..." {
		t.Errorf("truncate long = %q", got)
	}
}
