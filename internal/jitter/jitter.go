// Package jitter provides human-realistic timing distributions and mouse
// movement algorithms for anti-detection browser automation.
package jitter

import (
	"math"
	"math/rand"
	"time"
)

// Jitter applies uniform +/-30% variance to a duration.
func Jitter(d time.Duration) time.Duration {
	const pct = 0.30
	factor := 1.0 - pct + rand.Float64()*2*pct
	return time.Duration(float64(d) * factor)
}

// LogNormalSample returns a log-normally distributed value with the given mean
// and sigma. The mean parameter is the expected value of the distribution (not
// the mu parameter of the underlying normal).
func LogNormalSample(mean, sigma float64) float64 {
	mu := math.Log(mean) - sigma*sigma/2
	return math.Exp(mu + sigma*rand.NormFloat64())
}

// LogNormalDuration returns a log-normally distributed duration with the given
// mean duration and sigma.
func LogNormalDuration(mean time.Duration, sigma float64) time.Duration {
	return time.Duration(LogNormalSample(float64(mean), sigma))
}

// GammaSample returns a Gamma-distributed value with the given shape and rate
// parameters using the Marsaglia-Tsang method.
func GammaSample(shape, rate float64) float64 {
	if shape < 1 {
		// Gamma(a) = Gamma(a+1) * U^(1/a)
		return GammaSample(shape+1, rate) * math.Pow(rand.Float64(), 1.0/shape)
	}

	d := shape - 1.0/3.0
	c := 1.0 / math.Sqrt(9.0*d)

	for {
		var x, v float64
		for {
			x = rand.NormFloat64()
			v = 1.0 + c*x
			if v > 0 {
				break
			}
		}
		v = v * v * v
		u := rand.Float64()

		// Fast acceptance
		if u < 1.0-0.0331*(x*x)*(x*x) {
			return d * v / rate
		}
		// Slow acceptance
		if math.Log(u) < 0.5*x*x+d*(1.0-v+math.Log(v)) {
			return d * v / rate
		}
	}
}

// IKIDuration returns a realistic inter-keystroke interval duration.
// baseWPM controls the average typing speed (60 WPM ~ 200ms per character).
// Common bigrams are typed ~60% faster than rare pairs.
func IKIDuration(baseWPM float64, isCommonBigram bool) time.Duration {
	baseMs := 60000.0 / (baseWPM * 5) // 5 chars per word
	modifier := 1.0
	if isCommonBigram {
		modifier = 0.4
	}
	jitter := math.Exp(rand.NormFloat64() * 0.4)
	ms := baseMs * modifier * jitter
	if ms < 10 {
		ms = 10
	}
	return time.Duration(ms * float64(time.Millisecond))
}

// DwellDuration returns a realistic reading dwell time for a profile with the
// given word count. Based on 250 WPM reading speed with log-normal variance,
// clamped to [20s, 120s].
func DwellDuration(wordCount int) time.Duration {
	readTimeSec := float64(wordCount) / 250.0 * 60.0
	sample := LogNormalSample(readTimeSec, 0.5)
	if sample < 20 {
		sample = 20
	}
	if sample > 120 {
		sample = 120
	}
	return time.Duration(sample * float64(time.Second))
}
