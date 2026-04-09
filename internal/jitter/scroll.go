package jitter

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// ScrollMouse abstracts rod's Mouse.Scroll for testability.
type ScrollMouse interface {
	Scroll(x, y float64, steps int) error
}

// Sleeper abstracts time.Sleep for testability.
type Sleeper func(ctx context.Context, d time.Duration) error

// TimeSleeper is the production Sleeper that uses time.After.
func TimeSleeper(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// HumanScroll simulates a human scrolling totalPixels downward. It scrolls in
// log-normal-distributed chunks with reading pauses between them. There is a
// 20% chance of one upward direction reversal per call.
func HumanScroll(ctx context.Context, mouse ScrollMouse, totalPixels float64, sleep Sleeper) error {
	scrolled := 0.0
	reversed := false
	doReversal := rand.Float64() < 0.20

	for scrolled < totalPixels {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		chunk := scrollChunk(totalPixels - scrolled)
		if err := mouse.Scroll(0, chunk, int(chunk/15)+1); err != nil {
			return err
		}
		scrolled += chunk

		if doReversal && !reversed && scrolled > totalPixels*0.3 {
			reversed = true
			if err := scrollReversal(ctx, mouse, sleep); err != nil {
				return err
			}
		}

		if scrolled < totalPixels {
			if err := readingPause(ctx, sleep); err != nil {
				return err
			}
		}
	}
	return nil
}

func scrollChunk(remaining float64) float64 {
	chunk := LogNormalSample(120, 0.4)
	chunk = math.Min(chunk, 250)
	chunk = math.Min(chunk, remaining)
	return math.Max(chunk, 1)
}

func scrollReversal(ctx context.Context, mouse ScrollMouse, sleep Sleeper) error {
	reversal := LogNormalSample(80, 0.3)
	if err := mouse.Scroll(0, -reversal, int(reversal/15)+1); err != nil {
		return err
	}
	return readingPause(ctx, sleep)
}

func readingPause(ctx context.Context, sleep Sleeper) error {
	return sleep(ctx, LogNormalDuration(400*time.Millisecond, 0.5))
}
