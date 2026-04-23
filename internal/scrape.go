package explorer

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

// ScrapeUseCase discovers new prospects and warms up previously scanned ones.
// A single run performs two phases:
//  1. Discover — search LinkedIn, visit profiles, save as StateScanned.
//  2. Warm up — like posts for due Scanned prospects, advance to StateLiked.
type ScrapeUseCase struct {
	repo    ProspectRepository
	browser BrowserClient
	limiter RateLimiter
	log     *slog.Logger
	maxRun  int
}

// NewScrapeUseCase creates a ScrapeUseCase.
func NewScrapeUseCase(repo ProspectRepository, browser BrowserClient, limiter RateLimiter, log *slog.Logger, maxPerRun int) *ScrapeUseCase {
	return &ScrapeUseCase{repo: repo, browser: browser, limiter: limiter, log: log, maxRun: maxPerRun}
}

// Run executes both discover and warm-up phases.
func (uc *ScrapeUseCase) Run(ctx context.Context, query, location string) error {
	block, err := uc.browser.CheckBlock(ctx)
	if err != nil {
		return fmt.Errorf("check block: %w", err)
	}
	if block != BlockNone {
		return fmt.Errorf("%w: type %d", ErrBlockDetected, block)
	}

	if err := uc.discover(ctx, query, location); err != nil {
		return fmt.Errorf("discover: %w", err)
	}
	if err := uc.warmUp(ctx); err != nil {
		return fmt.Errorf("warm up: %w", err)
	}
	return nil
}

func (uc *ScrapeUseCase) discover(ctx context.Context, query, location string) error {
	urls, err := uc.browser.SearchProfiles(ctx, query, location, uc.maxRun)
	if err != nil {
		return fmt.Errorf("search profiles: %w", err)
	}
	uc.log.InfoContext(ctx, "search returned", "count", len(urls))

	for _, url := range urls {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := uc.limiter.Acquire(ctx, "profile_views"); err != nil {
			if errors.Is(err, ErrRateLimitExceeded) {
				uc.log.WarnContext(ctx, "daily view limit reached")
				return nil
			}
			return fmt.Errorf("acquire rate limit: %w", err)
		}

		data, err := uc.browser.VisitProfile(ctx, url)
		if err != nil {
			uc.log.WarnContext(ctx, "visit failed", "url", url, "error", err)
			continue
		}

		p := newProspectFromProfile(data)
		inserted, err := uc.repo.InsertIfNew(ctx, p)
		if err != nil {
			return fmt.Errorf("insert prospect: %w", err)
		}
		if inserted {
			uc.log.InfoContext(ctx, "new prospect", "url", data.URL, "name", data.Name)
		}
	}
	return nil
}

func (uc *ScrapeUseCase) warmUp(ctx context.Context) error {
	due, err := uc.repo.ListByState(ctx, StateScanned, time.Now())
	if err != nil {
		return fmt.Errorf("list due prospects: %w", err)
	}
	uc.log.InfoContext(ctx, "warm-up candidates", "count", len(due))

	for _, p := range due {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := uc.limiter.Acquire(ctx, "profile_views"); err != nil {
			if errors.Is(err, ErrRateLimitExceeded) {
				uc.log.WarnContext(ctx, "daily view limit reached during warm-up")
				return nil
			}
			return fmt.Errorf("acquire rate limit: %w", err)
		}

		if err := uc.browser.LikeRecentPost(ctx, p.ProfileURL); err != nil {
			uc.log.WarnContext(ctx, "like failed", "url", p.ProfileURL, "error", err)
			continue
		}
		if err := p.Transition(StateLiked); err != nil {
			return fmt.Errorf("transition %s: %w", p.ProfileURL, err)
		}
		if err := uc.repo.Save(ctx, p); err != nil {
			return fmt.Errorf("save prospect: %w", err)
		}
		uc.log.InfoContext(ctx, "liked post", "url", p.ProfileURL)
	}
	return nil
}

func newProspectFromProfile(data ProfileData) *Prospect {
	now := time.Now()
	return &Prospect{
		ProfileURL:   data.URL,
		Slug:         data.Slug,
		Name:         data.Name,
		Headline:     data.Headline,
		Location:     data.Location,
		About:        data.About,
		RecentPosts:  data.RecentPosts,
		State:        StateScanned,
		CreatedAt:    now,
		LastActionAt: now,
		NextActionAt: computeNextAction(StateScanned),
	}
}
