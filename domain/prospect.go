package domain

import (
	"fmt"
	"math/rand"
	"slices"
	"time"
)

// State represents the lifecycle stage of a prospect in the pipeline.
type State uint8

const (
	// StateScanned is the initial state: profile has been discovered and recorded.
	StateScanned State = iota
	// StateLiked means a post by this prospect has been liked as a warm-up action.
	StateLiked
	// StateDrafted means the LLM has generated a personalised connection message.
	StateDrafted
	// StateSent means the connection request has been sent by the human operator.
	StateSent
	// StateAccepted means the prospect accepted the connection request.
	StateAccepted
	// StateRejected means the prospect declined or ignored the request.
	StateRejected
	// StateSkipped means the prospect was explicitly removed from the pipeline.
	StateSkipped
)

// String returns the canonical uppercase representation of the state.
func (s State) String() string {
	switch s {
	case StateScanned:
		return "SCANNED"
	case StateLiked:
		return "LIKED"
	case StateDrafted:
		return "DRAFTED"
	case StateSent:
		return "SENT"
	case StateAccepted:
		return "ACCEPTED"
	case StateRejected:
		return "REJECTED"
	case StateSkipped:
		return "SKIPPED"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", s)
	}
}

// allowedTransitions maps each state to the set of states it may advance to.
var allowedTransitions = map[State][]State{
	StateScanned: {StateLiked, StateSkipped},
	StateLiked:   {StateDrafted},
	StateDrafted: {StateSent},
	StateSent:    {StateAccepted, StateRejected},
	// Terminal states: Accepted, Rejected, Skipped have no outgoing transitions.
}

// terminalStates is the set of states after which no further action is scheduled.
var terminalStates = map[State]bool{
	StateAccepted: true,
	StateRejected: true,
	StateSkipped:  true,
}

// Prospect is the core aggregate. It tracks a LinkedIn profile through the
// full warm-up and outreach pipeline.
type Prospect struct {
	// Identity
	ProfileURL string // LinkedIn profile URL (primary key in DynamoDB)
	Slug       string // URL slug extracted from ProfileURL

	// Profile data
	Name        string
	Headline    string
	Location    string
	About       string
	RecentPosts []string

	// Scoring
	WorthinessScore int    // 1–10, set by LLM analysis
	ScoreReasoning  string // ≤50 words from LLM
	DraftedMessage  string // 150–300 chars, ready for copy-paste
	CritiqueScore   int    // 3–15 composite score from LLM self-critique

	// Embedding
	EmbeddingID string // Qdrant point ID for semantic search

	// State machine
	State        State
	LastActionAt time.Time
	NextActionAt time.Time // zero for terminal states
	CreatedAt    time.Time
}

// Transition advances the prospect to the given state if the transition is
// permitted. It updates LastActionAt and schedules NextActionAt via
// computeNextAction. Returns ErrInvalidTransition if the move is not allowed.
func (p *Prospect) Transition(to State) error {
	if !slices.Contains(allowedTransitions[p.State], to) {
		return fmt.Errorf("%w: %s → %s", ErrInvalidTransition, p.State, to)
	}
	p.State = to
	p.LastActionAt = time.Now()
	p.NextActionAt = computeNextAction(to)
	return nil
}

// computeNextAction returns the earliest time at which the pipeline should
// act on a prospect in the given state. Terminal states return zero time.
// Non-terminal states get a 20–25 h window to spread load and add jitter.
func computeNextAction(s State) time.Time {
	if terminalStates[s] {
		return time.Time{}
	}
	// 20 h base + up to 5 h jitter (20–25% variance over a 20 h window).
	base := 20 * time.Hour
	jitter := time.Duration(rand.Int63n(int64(5 * time.Hour)))
	return time.Now().Add(base + jitter)
}
