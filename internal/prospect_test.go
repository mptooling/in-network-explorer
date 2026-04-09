package explorer_test

import (
	"errors"
	"testing"
	"time"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

func TestProspect_State_String(t *testing.T) {
	cases := []struct {
		state explorer.State
		want  string
	}{
		{explorer.StateScanned, "SCANNED"},
		{explorer.StateLiked, "LIKED"},
		{explorer.StateDrafted, "DRAFTED"},
		{explorer.StateSent, "SENT"},
		{explorer.StateAccepted, "ACCEPTED"},
		{explorer.StateRejected, "REJECTED"},
		{explorer.StateSkipped, "SKIPPED"},
	}
	for _, tc := range cases {
		if got := tc.state.String(); got != tc.want {
			t.Errorf("State(%d).String() = %q, want %q", tc.state, got, tc.want)
		}
	}
}

func TestProspect_Transition_Valid(t *testing.T) {
	cases := []struct {
		name string
		from explorer.State
		to   explorer.State
	}{
		{"Scanned→Liked", explorer.StateScanned, explorer.StateLiked},
		{"Scanned→Skipped", explorer.StateScanned, explorer.StateSkipped},
		{"Liked→Drafted", explorer.StateLiked, explorer.StateDrafted},
		{"Drafted→Sent", explorer.StateDrafted, explorer.StateSent},
		{"Sent→Accepted", explorer.StateSent, explorer.StateAccepted},
		{"Sent→Rejected", explorer.StateSent, explorer.StateRejected},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &explorer.Prospect{State: tc.from}
			if err := p.Transition(tc.to); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.State != tc.to {
				t.Fatalf("state = %s, want %s", p.State, tc.to)
			}
		})
	}
}

func TestProspect_Transition_Invalid(t *testing.T) {
	cases := []struct {
		name string
		from explorer.State
		to   explorer.State
	}{
		{"Scanned→Drafted", explorer.StateScanned, explorer.StateDrafted},
		{"Liked→Accepted", explorer.StateLiked, explorer.StateAccepted},
		{"Accepted→Sent", explorer.StateAccepted, explorer.StateSent},
		{"Rejected→Sent", explorer.StateRejected, explorer.StateSent},
		{"Skipped→Liked", explorer.StateSkipped, explorer.StateLiked},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &explorer.Prospect{State: tc.from}
			err := p.Transition(tc.to)
			if !errors.Is(err, explorer.ErrInvalidTransition) {
				t.Fatalf("expected ErrInvalidTransition, got %v", err)
			}
		})
	}
}

func TestProspect_Transition_UpdatesFields(t *testing.T) {
	t.Run("LastActionAt changes after transition", func(t *testing.T) {
		p := &explorer.Prospect{State: explorer.StateScanned}
		before := time.Now()
		if err := p.Transition(explorer.StateLiked); err != nil {
			t.Fatal(err)
		}
		if p.LastActionAt.Before(before) {
			t.Fatal("LastActionAt not updated after transition")
		}
	})

	t.Run("NextActionAt is in future (≥20h) after StateScanned", func(t *testing.T) {
		p := &explorer.Prospect{State: explorer.StateScanned}
		// We need to call Transition to a non-terminal state from a state that
		// leads to StateScanned being the origin. Actually StateScanned is the
		// initial state — test NextActionAt after inserting as new prospect.
		// Reconstruct: fresh prospect transitions to Liked, check NextActionAt.
		if err := p.Transition(explorer.StateLiked); err != nil {
			t.Fatal(err)
		}
		minNext := time.Now().Add(20 * time.Hour)
		if p.NextActionAt.Before(minNext) {
			t.Fatalf("NextActionAt %v is less than 20h from now", p.NextActionAt)
		}
	})

	t.Run("NextActionAt is in future (≥20h) after StateLiked", func(t *testing.T) {
		p := &explorer.Prospect{State: explorer.StateLiked}
		if err := p.Transition(explorer.StateDrafted); err != nil {
			t.Fatal(err)
		}
		minNext := time.Now().Add(20 * time.Hour)
		if p.NextActionAt.Before(minNext) {
			t.Fatalf("NextActionAt %v is less than 20h from now", p.NextActionAt)
		}
	})

	t.Run("NextActionAt is zero for terminal states", func(t *testing.T) {
		cases := []struct {
			from explorer.State
			to   explorer.State
		}{
			{explorer.StateSent, explorer.StateAccepted},
			{explorer.StateSent, explorer.StateRejected},
			{explorer.StateScanned, explorer.StateSkipped},
		}
		for _, tc := range cases {
			p := &explorer.Prospect{State: tc.from}
			if err := p.Transition(tc.to); err != nil {
				t.Fatal(err)
			}
			if !p.NextActionAt.IsZero() {
				t.Fatalf("NextActionAt should be zero for terminal state %s, got %v", tc.to, p.NextActionAt)
			}
		}
	})
}
