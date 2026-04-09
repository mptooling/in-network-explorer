package explorer_test

import (
	"errors"
	"testing"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

func TestSentinelErrors_NonNil(t *testing.T) {
	errs := []error{
		explorer.ErrNotFound,
		explorer.ErrInvalidTransition,
		explorer.ErrRateLimitExceeded,
		explorer.ErrBlockDetected,
		explorer.ErrSessionExpired,
	}
	for _, err := range errs {
		if err == nil {
			t.Fatalf("expected non-nil sentinel error, got nil")
		}
	}
}

func TestSentinelErrors_Distinct(t *testing.T) {
	if errors.Is(explorer.ErrInvalidTransition, explorer.ErrRateLimitExceeded) {
		t.Fatal("ErrInvalidTransition and ErrRateLimitExceeded must be distinct")
	}
	if errors.Is(explorer.ErrNotFound, explorer.ErrBlockDetected) {
		t.Fatal("ErrNotFound and ErrBlockDetected must be distinct")
	}
}

func TestSentinelErrors_Wrap(t *testing.T) {
	wrapped := errors.Join(explorer.ErrInvalidTransition, errors.New("Scanned → Drafted"))
	if !errors.Is(wrapped, explorer.ErrInvalidTransition) {
		t.Fatal("errors.Is must unwrap to ErrInvalidTransition")
	}
}
