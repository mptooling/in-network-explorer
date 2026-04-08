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

		chunk := LogNormalSample(120, 0.4)
		chunk = math.Min(chunk, 250) // cap at 250px
		chunk = math.Min(chunk, totalPixels-scrolled)
		chunk = math.Max(chunk, 1) // at least 1px
		steps := int(chunk/15) + 1

		if err := mouse.Scroll(0, chunk, steps); err != nil {
			return err
		}
		scrolled += chunk

		// One upward re-read reversal per session.
		if doReversal && !reversed && scrolled > totalPixels*0.3 {
			reversed = true
			reversal := LogNormalSample(80, 0.3)
			reverseSteps := int(reversal/15) + 1
			if err := mouse.Scroll(0, -reversal, reverseSteps); err != nil {
				return err
			}
			// Pause while "re-reading".
			if err := sleep(ctx, LogNormalDuration(400*time.Millisecond, 0.5)); err != nil {
				return err
			}
		}

		// Reading pause between chunks.
		if scrolled < totalPixels {
			if err := sleep(ctx, LogNormalDuration(400*time.Millisecond, 0.5)); err != nil {
				return err
			}
		}
	}
	return nil
}
