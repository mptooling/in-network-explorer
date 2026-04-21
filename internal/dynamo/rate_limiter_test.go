package dynamo

import (
	"context"
	"regexp"
	"testing"
	"time"
)

func TestRateLimitPK_Format(t *testing.T) {
	pk := rateLimitPK("profile_views")

	// Expected: RATE#profile_views#YYYY-MM-DD
	re := regexp.MustCompile(`^RATE#profile_views#\d{4}-\d{2}-\d{2}$`)
	if !re.MatchString(pk) {
		t.Fatalf("rateLimitPK() = %q, want format RATE#scope#YYYY-MM-DD", pk)
	}

	today := time.Now().UTC().Format("2006-01-02")
	want := "RATE#profile_views#" + today
	if pk != want {
		t.Fatalf("rateLimitPK() = %q, want %q", pk, want)
	}
}

func TestEndOfDayTTL_InFuture(t *testing.T) {
	ttl := endOfDayTTL()
	now := time.Now().Unix()

	if ttl <= now {
		t.Fatalf("endOfDayTTL() = %d, want > %d (now)", ttl, now)
	}

	// TTL should be between ~48h and ~72h from now.
	minOffset := int64(48 * 3600)
	maxOffset := int64(72 * 3600)
	delta := ttl - now
	if delta < minOffset || delta > maxOffset {
		t.Fatalf("endOfDayTTL() offset = %ds, want between %ds and %ds", delta, minOffset, maxOffset)
	}
}

func TestAcquire_UnknownScope(t *testing.T) {
	rl := NewRateLimiter(nil, "t", map[string]int{"a": 1})
	err := rl.Acquire(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error for unknown scope")
	}
	want := `unknown rate limit scope: "nope"`
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}

func TestCurrent_UnknownScope(t *testing.T) {
	rl := NewRateLimiter(nil, "t", map[string]int{"a": 1})
	_, err := rl.Current(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error for unknown scope")
	}
	want := `unknown rate limit scope: "nope"`
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err.Error(), want)
	}
}
