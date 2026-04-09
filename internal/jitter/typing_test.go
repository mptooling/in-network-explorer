package jitter_test

import (
	"context"
	"testing"
	"time"

	"github.com/pavlomaksymov/in-network-explorer/internal/jitter"
)

type keyEvent struct {
	kind rune // 0 = backspace, otherwise the pressed rune
}

type spyKeyboard struct {
	events []keyEvent
}

func (k *spyKeyboard) Press(key rune) error {
	k.events = append(k.events, keyEvent{kind: key})
	return nil
}

func (k *spyKeyboard) Backspace() error {
	k.events = append(k.events, keyEvent{kind: 0})
	return nil
}

// extractTypedText replays the keyboard events and returns the final text,
// accounting for backspace corrections.
func extractTypedText(events []keyEvent) string {
	var buf []rune
	for _, e := range events {
		if e.kind == 0 {
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
			}
		} else {
			buf = append(buf, e.kind)
		}
	}
	return string(buf)
}

func TestHumanType_AllCharactersTyped(t *testing.T) {
	kb := &spyKeyboard{}
	sleeper := &spySleeper{}
	input := "Hello, this is a test of the typing simulation."

	err := jitter.HumanType(context.Background(), kb, input, sleeper.sleep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := extractTypedText(kb.events)
	if got != input {
		t.Fatalf("typed text = %q, want %q", got, input)
	}
}

func TestHumanType_NoInstantKeystrokes(t *testing.T) {
	kb := &spyKeyboard{}
	sleeper := &spySleeper{}
	input := "Quick brown fox jumps over the lazy dog."

	err := jitter.HumanType(context.Background(), kb, input, sleeper.sleep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// There should be a sleep before each keystroke (including corrections).
	// Total sleeps should be at least len(input) - 1 (first char may not need sleep).
	if len(sleeper.durations) < len(input)-1 {
		t.Fatalf("got %d sleeps for %d char input, want at least %d",
			len(sleeper.durations), len(input), len(input)-1)
	}

	for i, d := range sleeper.durations {
		if d < 10*time.Millisecond {
			t.Fatalf("sleep %d: duration %v, want >= 10ms", i, d)
		}
	}
}

func TestHumanType_ErrorRate(t *testing.T) {
	const runs = 50
	input := "The quick brown fox jumps over the lazy dog. This is a test string that is exactly one hundred chars!!"

	totalBackspaces := 0
	totalKeys := 0

	for range runs {
		kb := &spyKeyboard{}
		sleeper := &spySleeper{}

		err := jitter.HumanType(context.Background(), kb, input, sleeper.sleep)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		for _, e := range kb.events {
			if e.kind == 0 {
				totalBackspaces++
			}
			totalKeys++
		}
	}

	errorRate := float64(totalBackspaces) / float64(totalKeys) * 100
	if errorRate < 1 || errorRate > 10 {
		t.Fatalf("error rate = %.1f%%, want 1-10%% (%d backspaces / %d total keys over %d runs)",
			errorRate, totalBackspaces, totalKeys, runs)
	}
}

func TestHumanType_WPMRange(t *testing.T) {
	kb := &spyKeyboard{}
	sleeper := &spySleeper{}
	input := "This is a fifty character test string for WPM chk!" // 50 chars

	err := jitter.HumanType(context.Background(), kb, input, sleeper.sleep)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var totalSleep time.Duration
	for _, d := range sleeper.durations {
		totalSleep += d
	}

	// WPM = (chars / 5) / (minutes)
	minutes := totalSleep.Minutes()
	words := float64(len(input)) / 5.0
	wpm := words / minutes

	if wpm < 30 || wpm > 120 {
		t.Fatalf("effective WPM = %.0f (total sleep %v for %d chars), want 30-120", wpm, totalSleep, len(input))
	}
}

func TestHumanType_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	kb := &spyKeyboard{}
	sleeper := &spySleeper{}

	err := jitter.HumanType(ctx, kb, "hello world", sleeper.sleep)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
