// Package domain contains the core business entities, value objects, and
// interfaces for the in-network-explorer prospect pipeline.
package domain

import "errors"

// Sentinel errors for domain-level conditions. Callers use errors.Is to
// distinguish them; never compare by string.
var (
	// ErrNotFound is returned when a prospect record does not exist in the store.
	ErrNotFound = errors.New("prospect not found")

	// ErrInvalidTransition is returned when a requested state transition is not
	// permitted by the prospect state machine.
	ErrInvalidTransition = errors.New("invalid state transition")

	// ErrRateLimitExceeded is returned when the daily cap for a given action
	// scope (profile views, connection requests) has been reached.
	ErrRateLimitExceeded = errors.New("daily rate limit exceeded")

	// ErrBlockDetected is returned when LinkedIn signals a soft or hard block
	// (CAPTCHA, auth-wall, empty challenge page).
	ErrBlockDetected = errors.New("linkedin block detected")

	// ErrSessionExpired is returned when the stored session cookie is no longer
	// valid and reauthentication is required.
	ErrSessionExpired = errors.New("linkedin session expired")
)
