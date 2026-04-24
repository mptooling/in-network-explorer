package explorer

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// AnalyzeUseCase scores Liked prospects via LLM and drafts connection messages.
// Prospects transition from StateLiked to StateDrafted. When embedding deps
// are provided, it also upserts profile vectors for semantic search.
type AnalyzeUseCase struct {
	repo    ProspectRepository
	llm     LLMClient
	embedCl EmbeddingClient // nil = skip embedding
	embedSt EmbeddingStore  // nil = skip embedding
	log     *slog.Logger
	fewShot int
}

// NewAnalyzeUseCase creates an AnalyzeUseCase. fewShot controls how many
// top-scored existing prospects are used as few-shot examples (0 = none).
// embedClient and embedStore are optional — pass nil to skip vector indexing.
func NewAnalyzeUseCase(
	repo ProspectRepository,
	llm LLMClient,
	log *slog.Logger,
	fewShot int,
	embedClient EmbeddingClient,
	embedStore EmbeddingStore,
) *AnalyzeUseCase {
	return &AnalyzeUseCase{
		repo: repo, llm: llm, log: log, fewShot: fewShot,
		embedCl: embedClient, embedSt: embedStore,
	}
}

// Run scores all Liked prospects whose NextActionAt is due.
func (uc *AnalyzeUseCase) Run(ctx context.Context) error {
	prospects, err := uc.repo.ListByState(ctx, StateLiked, time.Now())
	if err != nil {
		return fmt.Errorf("list liked prospects: %w", err)
	}
	if len(prospects) == 0 {
		uc.log.InfoContext(ctx, "no prospects ready for analysis")
		return nil
	}

	examples, err := uc.loadExamples(ctx)
	if err != nil {
		return err
	}
	uc.log.InfoContext(ctx, "analyzing prospects", "count", len(prospects), "examples", len(examples))

	for _, p := range prospects {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := uc.scoreOne(ctx, p, examples); err != nil {
			uc.log.WarnContext(ctx, "score failed", "url", p.ProfileURL, "error", err)
			continue
		}
		uc.embedOne(ctx, p)
	}
	return nil
}

func (uc *AnalyzeUseCase) loadExamples(ctx context.Context) ([]Prospect, error) {
	if uc.fewShot <= 0 {
		return nil, nil
	}
	top, err := uc.repo.ListByStateOrderedByScore(ctx, StateDrafted, uc.fewShot)
	if err != nil {
		return nil, fmt.Errorf("load few-shot examples: %w", err)
	}
	examples := make([]Prospect, len(top))
	for i, p := range top {
		examples[i] = *p
	}
	return examples, nil
}

func (uc *AnalyzeUseCase) scoreOne(ctx context.Context, p *Prospect, examples []Prospect) error {
	result, err := uc.llm.ScoreAndDraft(ctx, p, examples)
	if err != nil {
		return fmt.Errorf("score %s: %w", p.ProfileURL, err)
	}

	p.WorthinessScore = result.Score
	p.ScoreReasoning = result.Reasoning
	p.DraftedMessage = result.Message
	p.CritiqueScore = result.CritiqueScore

	if err := p.Transition(StateDrafted); err != nil {
		return fmt.Errorf("transition %s: %w", p.ProfileURL, err)
	}
	if err := uc.repo.Save(ctx, p); err != nil {
		return fmt.Errorf("save %s: %w", p.ProfileURL, err)
	}

	uc.log.InfoContext(ctx, "scored prospect",
		"url", p.ProfileURL,
		"score", result.Score,
		"critique", result.CritiqueScore,
	)
	return nil
}

// embedOne generates and stores a vector for the prospect's profile text.
// Failures are logged but do not halt the pipeline.
func (uc *AnalyzeUseCase) embedOne(ctx context.Context, p *Prospect) {
	if uc.embedCl == nil || uc.embedSt == nil {
		return
	}

	text := strings.Join([]string{p.Name, p.Headline, p.Location, p.About}, " — ")
	vec, err := uc.embedCl.Embed(ctx, text)
	if err != nil {
		uc.log.WarnContext(ctx, "embed failed", "url", p.ProfileURL, "error", err)
		return
	}

	payload := map[string]any{
		"profile_url":      p.ProfileURL,
		"name":             p.Name,
		"headline":         p.Headline,
		"location":         p.Location,
		"worthiness_score": p.WorthinessScore,
	}
	if err := uc.embedSt.Upsert(ctx, p.Slug, vec, payload); err != nil {
		uc.log.WarnContext(ctx, "upsert embedding failed", "url", p.ProfileURL, "error", err)
	}
}
