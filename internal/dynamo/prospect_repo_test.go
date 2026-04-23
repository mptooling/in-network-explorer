package dynamo

import (
	"testing"
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

func TestParseState_AllStates(t *testing.T) {
	cases := []struct {
		input string
		want  explorer.State
	}{
		{"SCANNED", explorer.StateScanned},
		{"LIKED", explorer.StateLiked},
		{"DRAFTED", explorer.StateDrafted},
		{"SENT", explorer.StateSent},
		{"ACCEPTED", explorer.StateAccepted},
		{"REJECTED", explorer.StateRejected},
		{"SKIPPED", explorer.StateSkipped},
	}
	for _, tc := range cases {
		got, err := parseState(tc.input)
		if err != nil {
			t.Fatalf("parseState(%q) error = %v", tc.input, err)
		}
		if got != tc.want {
			t.Fatalf("parseState(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
}

func TestParseState_Unknown(t *testing.T) {
	_, err := parseState("BOGUS")
	if err == nil {
		t.Fatal("expected error for unknown state")
	}
}

func TestFormatTime_Zero(t *testing.T) {
	if got := formatTime(time.Time{}); got != "" {
		t.Fatalf("formatTime(zero) = %q, want empty", got)
	}
}

func TestParseTime_Empty(t *testing.T) {
	if got := parseTime(""); !got.IsZero() {
		t.Fatalf("parseTime(\"\") = %v, want zero", got)
	}
}

func TestRoundtripMarshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	p := &explorer.Prospect{
		ProfileURL:      "https://linkedin.com/in/alice",
		Slug:            "alice",
		Name:            "Alice",
		Headline:        "Engineer",
		Location:        "Berlin",
		About:           "About text",
		RecentPosts:     []string{"post1", "post2"},
		WorthinessScore: 8,
		ScoreReasoning:  "Good fit",
		DraftedMessage:  "Hi Alice",
		CritiqueScore:   12,
		EmbeddingID:     "emb-123",
		State:           explorer.StateDrafted,
		LastActionAt:    now,
		NextActionAt:    now.Add(24 * time.Hour),
		CreatedAt:       now.Add(-48 * time.Hour),
	}

	av, err := marshalProspect(p)
	if err != nil {
		t.Fatalf("marshalProspect: %v", err)
	}

	got, err := unmarshalProspect(av)
	if err != nil {
		t.Fatalf("unmarshalProspect: %v", err)
	}

	if got.ProfileURL != p.ProfileURL {
		t.Errorf("ProfileURL = %q, want %q", got.ProfileURL, p.ProfileURL)
	}
	if got.Slug != p.Slug {
		t.Errorf("Slug = %q, want %q", got.Slug, p.Slug)
	}
	if got.Name != p.Name {
		t.Errorf("Name = %q, want %q", got.Name, p.Name)
	}
	if got.State != p.State {
		t.Errorf("State = %v, want %v", got.State, p.State)
	}
	if got.WorthinessScore != p.WorthinessScore {
		t.Errorf("WorthinessScore = %d, want %d", got.WorthinessScore, p.WorthinessScore)
	}
	if got.CritiqueScore != p.CritiqueScore {
		t.Errorf("CritiqueScore = %d, want %d", got.CritiqueScore, p.CritiqueScore)
	}
	if len(got.RecentPosts) != 2 {
		t.Errorf("RecentPosts len = %d, want 2", len(got.RecentPosts))
	}
	if !got.LastActionAt.Equal(p.LastActionAt) {
		t.Errorf("LastActionAt = %v, want %v", got.LastActionAt, p.LastActionAt)
	}
	if !got.NextActionAt.Equal(p.NextActionAt) {
		t.Errorf("NextActionAt = %v, want %v", got.NextActionAt, p.NextActionAt)
	}
	if !got.CreatedAt.Equal(p.CreatedAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, p.CreatedAt)
	}
}

func TestRoundtripMarshal_ZeroTimes(t *testing.T) {
	p := &explorer.Prospect{
		ProfileURL: "https://linkedin.com/in/bob",
		State:      explorer.StateAccepted,
	}
	av, err := marshalProspect(p)
	if err != nil {
		t.Fatalf("marshalProspect: %v", err)
	}
	got, err := unmarshalProspect(av)
	if err != nil {
		t.Fatalf("unmarshalProspect: %v", err)
	}
	if !got.NextActionAt.IsZero() {
		t.Errorf("NextActionAt = %v, want zero", got.NextActionAt)
	}
}
