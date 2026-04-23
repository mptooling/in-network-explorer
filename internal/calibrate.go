package explorer

import (
	"context"
	"fmt"
	"log/slog"
	"math"
)

// CalibrateUseCase re-scores a sample of previously drafted prospects and
// reports the variance between original and new scores. This validates LLM
// scoring consistency over time.
type CalibrateUseCase struct {
	repo       ProspectRepository
	llm        LLMClient
	log        *slog.Logger
	sampleSize int
}

// NewCalibrateUseCase creates a CalibrateUseCase. sampleSize controls how many
// previously drafted prospects are re-scored.
func NewCalibrateUseCase(repo ProspectRepository, llm LLMClient, log *slog.Logger, sampleSize int) *CalibrateUseCase {
	return &CalibrateUseCase{repo: repo, llm: llm, log: log, sampleSize: sampleSize}
}

// CalibrationResult holds the outcome of a calibration run.
type CalibrationResult struct {
	Checked  int
	MeanDiff float64 // average absolute difference between old and new scores
	MaxDiff  int     // worst-case single-prospect score drift
}

// Run re-scores a sample of Drafted prospects and returns the variance report.
func (uc *CalibrateUseCase) Run(ctx context.Context) (*CalibrationResult, error) {
	prospects, err := uc.repo.ListByStateOrderedByScore(ctx, StateDrafted, uc.sampleSize)
	if err != nil {
		return nil, fmt.Errorf("list drafted prospects: %w", err)
	}
	if len(prospects) == 0 {
		uc.log.InfoContext(ctx, "no drafted prospects to calibrate")
		return &CalibrationResult{}, nil
	}

	var totalDiff float64
	var maxDiff int

	for _, p := range prospects {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		result, err := uc.llm.ScoreAndDraft(ctx, p, nil)
		if err != nil {
			uc.log.WarnContext(ctx, "re-score failed", "url", p.ProfileURL, "error", err)
			continue
		}

		diff := abs(result.Score - p.WorthinessScore)
		totalDiff += float64(diff)
		if diff > maxDiff {
			maxDiff = diff
		}

		uc.log.InfoContext(ctx, "calibration sample",
			"url", p.ProfileURL,
			"original", p.WorthinessScore,
			"new", result.Score,
			"diff", diff,
		)
	}

	mean := 0.0
	if len(prospects) > 0 {
		mean = math.Round(totalDiff/float64(len(prospects))*100) / 100
	}

	res := &CalibrationResult{
		Checked:  len(prospects),
		MeanDiff: mean,
		MaxDiff:  maxDiff,
	}
	uc.log.InfoContext(ctx, "calibration complete",
		"checked", res.Checked,
		"mean_diff", res.MeanDiff,
		"max_diff", res.MaxDiff,
	)
	return res, nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
