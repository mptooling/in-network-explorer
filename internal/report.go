package explorer

import (
	"context"
	"fmt"
	"log/slog"
)

// ProspectReport is a read-only view of a drafted prospect for the human
// operator. It contains the data needed to decide whether to send the message.
type ProspectReport struct {
	ProfileURL      string
	Name            string
	Headline        string
	Location        string
	WorthinessScore int
	ScoreReasoning  string
	DraftedMessage  string
	CritiqueScore   int
}

// ReportUseCase generates a prospect report from the top drafted prospects.
type ReportUseCase struct {
	repo  ProspectRepository
	log   *slog.Logger
	limit int
}

// NewReportUseCase creates a ReportUseCase. limit controls how many top-scored
// Drafted prospects are included in the report.
func NewReportUseCase(repo ProspectRepository, log *slog.Logger, limit int) *ReportUseCase {
	return &ReportUseCase{repo: repo, log: log, limit: limit}
}

// Run fetches drafted prospects ordered by score and returns report entries.
func (uc *ReportUseCase) Run(ctx context.Context) ([]ProspectReport, error) {
	prospects, err := uc.repo.ListByStateOrderedByScore(ctx, StateDrafted, uc.limit)
	if err != nil {
		return nil, fmt.Errorf("list drafted prospects: %w", err)
	}
	uc.log.InfoContext(ctx, "generating report", "prospects", len(prospects))

	entries := make([]ProspectReport, len(prospects))
	for i, p := range prospects {
		entries[i] = ProspectReport{
			ProfileURL:      p.ProfileURL,
			Name:            p.Name,
			Headline:        p.Headline,
			Location:        p.Location,
			WorthinessScore: p.WorthinessScore,
			ScoreReasoning:  p.ScoreReasoning,
			DraftedMessage:  p.DraftedMessage,
			CritiqueScore:   p.CritiqueScore,
		}
	}
	return entries, nil
}
