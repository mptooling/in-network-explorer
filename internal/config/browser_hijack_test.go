package config

import "testing"

func TestExtensionHijack_SimulatesExactly5(t *testing.T) {
	if got := len(knownExtensions); got != 5 {
		t.Errorf("knownExtensions count = %d, want 5", got)
	}
}
