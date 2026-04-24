// Package bedrock implements the LLMClient interface using Amazon Bedrock's
// Converse API with Claude models.
package bedrock

import (
	"encoding/json"
	"fmt"
	"strings"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

// ── prompt builders ─────────────────────────────────────────────────────────

func buildScorePrompt(p *explorer.Prospect, examples []explorer.Prospect) string {
	var sb strings.Builder

	sb.WriteString(`You are an AI assistant helping identify ideal networking prospects for an IT professional based in Berlin.

Evaluate the following LinkedIn prospect and:
1. Score them 1-10 based on: location relevance (Berlin preferred), field alignment (IT/tech), activity level, and connection potential.
2. Provide a brief reasoning (≤50 words).
3. Draft a personalized connection request message (150-300 characters) that references their specific activity or background.
4. Self-critique the drafted message on specificity (1-5), relevance (1-5), and tone (1-5).

Respond ONLY with JSON, no other text:
{"score":N,"reasoning":"...","message":"...","critique":{"specificity":N,"relevance":N,"tone":N}}

## Prospect
`)
	writeProspectFields(&sb, p)

	if len(examples) > 0 {
		sb.WriteString("\n## High-scoring examples for reference\n")
		for i, ex := range examples {
			fmt.Fprintf(&sb, "\n### Example %d (score: %d)\n", i+1, ex.WorthinessScore)
			writeProspectFields(&sb, &ex)
			if ex.DraftedMessage != "" {
				fmt.Fprintf(&sb, "Drafted message: %s\n", ex.DraftedMessage)
			}
		}
	}
	return sb.String()
}

func buildCritiquePrompt(message string) string {
	return fmt.Sprintf(`Evaluate this LinkedIn connection request message on three criteria (each 1-5):
1. Specificity — Does it reference specific details about the person?
2. Relevance — Is it relevant to their work/interests?
3. Tone — Is it professional, warm, and not salesy?

Message: %q

Respond ONLY with JSON, no other text:
{"specificity":N,"relevance":N,"tone":N}`, message)
}

func writeProspectFields(sb *strings.Builder, p *explorer.Prospect) {
	fmt.Fprintf(sb, "Name: %s\n", p.Name)
	fmt.Fprintf(sb, "Headline: %s\n", p.Headline)
	fmt.Fprintf(sb, "Location: %s\n", p.Location)
	if p.About != "" {
		fmt.Fprintf(sb, "About: %s\n", p.About)
	}
	if len(p.RecentPosts) > 0 {
		sb.WriteString("Recent posts:\n")
		for _, post := range p.RecentPosts {
			fmt.Fprintf(sb, "- %s\n", truncate(post, 300))
		}
	}
}

// ── response parsers ────────────────────────────────────────────────────────

type scoreResponse struct {
	Score     int             `json:"score"`
	Reasoning string          `json:"reasoning"`
	Message   string          `json:"message"`
	Critique  critiqueSection `json:"critique"`
}

type critiqueSection struct {
	Specificity int `json:"specificity"`
	Relevance   int `json:"relevance"`
	Tone        int `json:"tone"`
}

func parseScoreResponse(raw string) (explorer.ScoreResult, error) {
	cleaned := stripCodeFences(raw)
	var resp scoreResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return explorer.ScoreResult{}, fmt.Errorf("parse score response: %w", err)
	}
	return explorer.ScoreResult{
		Score:         clamp(resp.Score, 1, 10),
		Reasoning:     resp.Reasoning,
		Message:       resp.Message,
		CritiqueScore: clamp(resp.Critique.Specificity+resp.Critique.Relevance+resp.Critique.Tone, 3, 15),
	}, nil
}

func parseCritiqueResponse(raw string) (int, error) {
	cleaned := stripCodeFences(raw)
	var resp critiqueSection
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return 0, fmt.Errorf("parse critique response: %w", err)
	}
	return clamp(resp.Specificity+resp.Relevance+resp.Tone, 3, 15), nil
}

// ── helpers ─────────────────────────────────────────────────────────────────

func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
