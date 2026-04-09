package jitter_test

import (
	"math"
	"testing"
	"time"

	"github.com/pavlomaksymov/in-network-explorer/internal/jitter"
)

func TestJitter_Range(t *testing.T) {
	base := 100 * time.Millisecond
	for i := 0; i < 1000; i++ {
		got := jitter.Jitter(base)
		if got < 70*time.Millisecond || got > 130*time.Millisecond {
			t.Fatalf("iteration %d: Jitter(100ms) = %v, want [70ms, 130ms]", i, got)
		}
	}
}

func TestLogNormalSample_Positive(t *testing.T) {
	for i := 0; i < 1000; i++ {
		got := jitter.LogNormalSample(100, 0.5)
		if got <= 0 {
			t.Fatalf("iteration %d: LogNormalSample(100, 0.5) = %v, want > 0", i, got)
		}
	}
}

func TestLogNormalSample_Distribution(t *testing.T) {
	const (
		n     = 10000
		mean  = 100.0
		sigma = 0.5
	)

	// Theoretical stddev of LogNormal: mean * sqrt(exp(sigma^2) - 1)
	theoreticalStddev := mean * math.Sqrt(math.Exp(sigma*sigma)-1)

	sum := 0.0
	samples := make([]float64, n)
	for i := 0; i < n; i++ {
		s := jitter.LogNormalSample(mean, sigma)
		samples[i] = s
		sum += s
	}

	sampleMean := sum / n
	sumSqDiff := 0.0
	for _, s := range samples {
		d := s - sampleMean
		sumSqDiff += d * d
	}
	sampleStddev := math.Sqrt(sumSqDiff / float64(n-1))

	if math.Abs(sampleMean-mean)/mean > 0.10 {
		t.Errorf("sample mean = %.2f, want within 10%% of %.2f", sampleMean, mean)
	}
	if math.Abs(sampleStddev-theoreticalStddev)/theoreticalStddev > 0.30 {
		t.Errorf("sample stddev = %.2f, want within 30%% of %.2f", sampleStddev, theoreticalStddev)
	}
}

func TestGammaSample_Positive(t *testing.T) {
	for i := 0; i < 1000; i++ {
		got := jitter.GammaSample(2, 0.02)
		if got <= 0 {
			t.Fatalf("iteration %d: GammaSample(2, 0.02) = %v, want > 0", i, got)
		}
	}
}

func TestIKIDuration_CommonBigram(t *testing.T) {
	const n = 500
	var commonSum, rareSum time.Duration
	for i := 0; i < n; i++ {
		commonSum += jitter.IKIDuration(60, true)
		rareSum += jitter.IKIDuration(60, false)
	}
	commonAvg := commonSum / n
	rareAvg := rareSum / n
	if commonAvg >= rareAvg {
		t.Errorf("common bigram avg %v should be less than rare pair avg %v", commonAvg, rareAvg)
	}
}

func TestIKIDuration_NeverZero(t *testing.T) {
	for i := 0; i < 1000; i++ {
		got := jitter.IKIDuration(60, false)
		if got <= 0 {
			t.Fatalf("iteration %d: IKIDuration(60, false) = %v, want > 0", i, got)
		}
	}
}

func TestDwellDuration_Clamped(t *testing.T) {
	for i := 0; i < 1000; i++ {
		got := jitter.DwellDuration(150)
		if got < 20*time.Second || got > 120*time.Second {
			t.Fatalf("iteration %d: DwellDuration(150) = %v, want [20s, 120s]", i, got)
		}
	}
}
