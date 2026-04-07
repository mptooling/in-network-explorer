package domain_test

import (
	"errors"
	"testing"

	"github.com/pavlomaksymov/in-network-explorer/domain"
)

func TestSentinelErrors_NonNil(t *testing.T) {
	errs := []error{
		domain.ErrNotFound,
		domain.ErrInvalidTransition,
		domain.ErrRateLimitExceeded,
		domain.ErrBlockDetected,
		domain.ErrSessionExpired,
	}
	for _, err := range errs {
		if err == nil {
			t.Fatalf("expected non-nil sentinel error, got nil")
		}
	}
}

func TestSentinelErrors_Distinct(t *testing.T) {
	if errors.Is(domain.ErrInvalidTransition, domain.ErrRateLimitExceeded) {
		t.Fatal("ErrInvalidTransition and ErrRateLimitExceeded must be distinct")
	}
	if errors.Is(domain.ErrNotFound, domain.ErrBlockDetected) {
		t.Fatal("ErrNotFound and ErrBlockDetected must be distinct")
	}
}

func TestSentinelErrors_Wrap(t *testing.T) {
	wrapped := errors.Join(domain.ErrInvalidTransition, errors.New("Scanned → Drafted"))
	if !errors.Is(wrapped, domain.ErrInvalidTransition) {
		t.Fatal("errors.Is must unwrap to ErrInvalidTransition")
	}
}
