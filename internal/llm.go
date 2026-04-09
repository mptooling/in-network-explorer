package explorer

import "context"

// ScoreResult holds the LLM output for a single prospect scoring pass.
type ScoreResult struct {
	// Score is the worthiness score from 1 (poor fit) to 10 (ideal prospect).
	Score int

	// Reasoning is a concise explanation of the score in ≤50 words.
	Reasoning string

	// Message is the personalised connection request draft (150–300 chars).
	Message string

	// CritiqueScore is the composite self-critique score (specificity +
	// relevance + tone), ranging from 3 (poor) to 15 (excellent).
	CritiqueScore int
}

// LLMClient abstracts the AI backend (Bedrock, OpenAI, etc.).
// Implementations live in internal/bedrock.
type LLMClient interface {
	// ScoreAndDraft scores the prospect against the target persona and drafts a
	// personalised connection message. examples are high-scoring reference
	// prospects used for few-shot prompting.
	ScoreAndDraft(ctx context.Context, p *Prospect, examples []Prospect) (ScoreResult, error)

	// Critique evaluates the drafted message and returns a composite score
	// from 3 (poor) to 15 (excellent) based on specificity, relevance, and tone.
	Critique(ctx context.Context, message string) (int, error)
}
