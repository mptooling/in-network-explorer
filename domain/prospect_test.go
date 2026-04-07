package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/pavlomaksymov/in-network-explorer/domain"
)

func TestProspect_State_String(t *testing.T) {
	cases := []struct {
		state domain.State
		want  string
	}{
		{domain.StateScanned, "SCANNED"},
		{domain.StateLiked, "LIKED"},
		{domain.StateDrafted, "DRAFTED"},
		{domain.StateSent, "SENT"},
		{domain.StateAccepted, "ACCEPTED"},
		{domain.StateRejected, "REJECTED"},
		{domain.StateSkipped, "SKIPPED"},
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
		from domain.State
		to   domain.State
	}{
		{"Scanned→Liked", domain.StateScanned, domain.StateLiked},
		{"Scanned→Skipped", domain.StateScanned, domain.StateSkipped},
		{"Liked→Drafted", domain.StateLiked, domain.StateDrafted},
		{"Drafted→Sent", domain.StateDrafted, domain.StateSent},
		{"Sent→Accepted", domain.StateSent, domain.StateAccepted},
		{"Sent→Rejected", domain.StateSent, domain.StateRejected},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &domain.Prospect{State: tc.from}
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
		from domain.State
		to   domain.State
	}{
		{"Scanned→Drafted", domain.StateScanned, domain.StateDrafted},
		{"Liked→Accepted", domain.StateLiked, domain.StateAccepted},
		{"Accepted→Sent", domain.StateAccepted, domain.StateSent},
		{"Rejected→Sent", domain.StateRejected, domain.StateSent},
		{"Skipped→Liked", domain.StateSkipped, domain.StateLiked},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := &domain.Prospect{State: tc.from}
			err := p.Transition(tc.to)
			if !errors.Is(err, domain.ErrInvalidTransition) {
				t.Fatalf("expected ErrInvalidTransition, got %v", err)
			}
		})
	}
}

func TestProspect_Transition_UpdatesFields(t *testing.T) {
	t.Run("LastActionAt changes after transition", func(t *testing.T) {
		p := &domain.Prospect{State: domain.StateScanned}
		before := time.Now()
		if err := p.Transition(domain.StateLiked); err != nil {
			t.Fatal(err)
		}
		if p.LastActionAt.Before(before) {
			t.Fatal("LastActionAt not updated after transition")
		}
	})

	t.Run("NextActionAt is in future (≥20h) after StateScanned", func(t *testing.T) {
		p := &domain.Prospect{State: domain.StateScanned}
		// We need to call Transition to a non-terminal state from a state that
		// leads to StateScanned being the origin. Actually StateScanned is the
		// initial state — test NextActionAt after inserting as new prospect.
		// Reconstruct: fresh prospect transitions to Liked, check NextActionAt.
		if err := p.Transition(domain.StateLiked); err != nil {
			t.Fatal(err)
		}
		minNext := time.Now().Add(20 * time.Hour)
		if p.NextActionAt.Before(minNext) {
			t.Fatalf("NextActionAt %v is less than 20h from now", p.NextActionAt)
		}
	})

	t.Run("NextActionAt is in future (≥20h) after StateLiked", func(t *testing.T) {
		p := &domain.Prospect{State: domain.StateLiked}
		if err := p.Transition(domain.StateDrafted); err != nil {
			t.Fatal(err)
		}
		minNext := time.Now().Add(20 * time.Hour)
		if p.NextActionAt.Before(minNext) {
			t.Fatalf("NextActionAt %v is less than 20h from now", p.NextActionAt)
		}
	})

	t.Run("NextActionAt is zero for terminal states", func(t *testing.T) {
		cases := []struct {
			from domain.State
			to   domain.State
		}{
			{domain.StateSent, domain.StateAccepted},
			{domain.StateSent, domain.StateRejected},
			{domain.StateScanned, domain.StateSkipped},
		}
		for _, tc := range cases {
			p := &domain.Prospect{State: tc.from}
			if err := p.Transition(tc.to); err != nil {
				t.Fatal(err)
			}
			if !p.NextActionAt.IsZero() {
				t.Fatalf("NextActionAt should be zero for terminal state %s, got %v", tc.to, p.NextActionAt)
			}
		}
	})
}
