package jitter

import (
	"context"
	"math/rand"
	"strings"
)

// Keyboard abstracts rod's keyboard for testability.
type Keyboard interface {
	Press(key rune) error
	Backspace() error
}

// commonBigrams is a set of English character pairs that are typed faster.
var commonBigrams = map[string]bool{
	"th": true, "he": true, "in": true, "er": true, "an": true,
	"re": true, "on": true, "at": true, "en": true, "nd": true,
	"ti": true, "es": true, "or": true, "te": true, "of": true,
	"ed": true, "is": true, "it": true, "al": true, "ar": true,
	"st": true, "to": true, "nt": true, "ng": true, "se": true,
}

const typingErrorRate = 0.04 // 4% chance per keystroke

// HumanType simulates character-by-character typing with realistic inter-
// keystroke intervals, occasional errors, and corrections via backspace.
func HumanType(ctx context.Context, kb Keyboard, text string, sleep Sleeper) error {
	runes := []rune(text)
	lower := strings.ToLower(text)

	for i, ch := range runes {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if i > 0 {
			if err := interKeystrokeDelay(ctx, sleep, lower, i); err != nil {
				return err
			}
		}
		if rand.Float64() < typingErrorRate {
			if err := simulateTypo(ctx, kb, sleep, ch); err != nil {
				return err
			}
		}
		if err := kb.Press(ch); err != nil {
			return err
		}
	}
	return nil
}

func interKeystrokeDelay(ctx context.Context, sleep Sleeper, lower string, i int) error {
	bigram := false
	if i < len(lower) {
		bigram = commonBigrams[lower[i-1:i+1]]
	}
	return sleep(ctx, IKIDuration(60, bigram))
}

func simulateTypo(ctx context.Context, kb Keyboard, sleep Sleeper, correct rune) error {
	if err := kb.Press(neighborKey(correct)); err != nil {
		return err
	}
	if err := sleep(ctx, IKIDuration(60, false)); err != nil {
		return err
	}
	if err := kb.Backspace(); err != nil {
		return err
	}
	return sleep(ctx, IKIDuration(60, false))
}

// neighborKey returns a plausible adjacent key for a typo. For simplicity,
// it shifts the character by +1 or -1 in the rune space, staying within the
// same case range.
func neighborKey(ch rune) rune {
	if rand.Float64() < 0.5 {
		return ch + 1
	}
	return ch - 1
}
