package jitter_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/pavlomaksymov/in-network-explorer/internal/jitter"
)

type scrollCall struct {
	x, y  float64
	steps int
}

type spyScrollMouse struct {
	calls []scrollCall
}

func (s *spyScrollMouse) Scroll(x, y float64, steps int) error {
	s.calls = append(s.calls, scrollCall{x, y, steps})
	return nil
}

type spySleeper struct {
	durations []time.Duration
}

func (s *spySleeper) sleep(ctx context.Context, d time.Duration) error {
	s.durations = append(s.durations, d)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return nil
}

func TestHumanScroll_TotalPixels(t *testing.T) {
	mouse := &spyScrollMouse{}
	sleeper := &spySleeper{}

	err := jitter.HumanScroll(context.Background(), mouse, 1000, sleeper.sleep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	totalDown := 0.0
	for _, c := range mouse.calls {
		if c.y > 0 {
			totalDown += c.y
		}
	}

	tolerance := 1000.0 * 0.05
	if math.Abs(totalDown-1000) > tolerance {
		t.Fatalf("total downward scroll = %.0fpx, want 1000px +/-5%%", totalDown)
	}
}

func TestHumanScroll_ChunkSizes(t *testing.T) {
	mouse := &spyScrollMouse{}
	sleeper := &spySleeper{}

	err := jitter.HumanScroll(context.Background(), mouse, 2000, sleeper.sleep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, c := range mouse.calls {
		absY := math.Abs(c.y)
		if absY < 1 {
			continue // skip tiny remainders or reversal scrolls
		}
		// Last chunk can be smaller (remainder), skip it.
		if i == len(mouse.calls)-1 && c.y > 0 && absY < 50 {
			continue
		}
		if absY > 250 {
			t.Fatalf("chunk %d: %.0fpx exceeds 250px max", i, absY)
		}
	}
}

func TestHumanScroll_PauseBetweenChunks(t *testing.T) {
	mouse := &spyScrollMouse{}
	sleeper := &spySleeper{}

	err := jitter.HumanScroll(context.Background(), mouse, 500, sleeper.sleep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sleeper.durations) == 0 {
		t.Fatal("expected at least one sleep call between scroll chunks")
	}
	for i, d := range sleeper.durations {
		if d <= 0 {
			t.Fatalf("sleep %d: duration %v, want > 0", i, d)
		}
	}
}

func TestHumanScroll_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	mouse := &spyScrollMouse{}
	sleeper := &spySleeper{}

	err := jitter.HumanScroll(ctx, mouse, 5000, sleeper.sleep)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestHumanScroll_DirectionReversals(t *testing.T) {
	const runs = 100
	reversalCount := 0

	for range runs {
		mouse := &spyScrollMouse{}
		sleeper := &spySleeper{}

		err := jitter.HumanScroll(context.Background(), mouse, 1000, sleeper.sleep)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, c := range mouse.calls {
			if c.y < 0 {
				reversalCount++
				break
			}
		}
	}

	pct := float64(reversalCount) / runs * 100
	if pct < 5 || pct > 45 {
		t.Fatalf("reversal rate = %.0f%%, want 5-45%% over %d runs", pct, runs)
	}
}
