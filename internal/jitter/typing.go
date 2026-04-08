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

		// Inter-keystroke delay (skip before first character).
		if i > 0 {
			bigram := false
			if i < len(lower) {
				pair := lower[i-1 : i+1]
				bigram = commonBigrams[pair]
			}
			delay := IKIDuration(60, bigram)
			if err := sleep(ctx, delay); err != nil {
				return err
			}
		}

		// Simulate occasional typo: press wrong key, then backspace and correct.
		if rand.Float64() < typingErrorRate {
			wrong := neighborKey(ch)
			if err := kb.Press(wrong); err != nil {
				return err
			}
			// Brief pause before noticing the error.
			if err := sleep(ctx, IKIDuration(60, false)); err != nil {
				return err
			}
			if err := kb.Backspace(); err != nil {
				return err
			}
			// Brief pause before retyping.
			if err := sleep(ctx, IKIDuration(60, false)); err != nil {
				return err
			}
		}

		if err := kb.Press(ch); err != nil {
			return err
		}
	}
	return nil
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
