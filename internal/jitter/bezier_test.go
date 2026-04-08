package jitter_test

import (
	"testing"

	"github.com/pavlomaksymov/in-network-explorer/internal/jitter"
)

func TestBezierGuide_StartsAtStart(t *testing.T) {
	start := jitter.Point{X: 50, Y: 100}
	end := jitter.Point{X: 400, Y: 300}
	guide := jitter.BezierGuide(start, end)

	first, _ := guide()
	if first.X != start.X || first.Y != start.Y {
		t.Fatalf("first point = %v, want %v", first, start)
	}
}

func TestBezierGuide_EndsAtEnd(t *testing.T) {
	start := jitter.Point{X: 50, Y: 100}
	end := jitter.Point{X: 400, Y: 300}
	guide := jitter.BezierGuide(start, end)

	points := drainGuide(guide, 10000)
	last := points[len(points)-1]
	if d := dist(last, end); d > 1 {
		t.Fatalf("last point %v is %.2fpx from end %v, want <= 1px", last, d, end)
	}
}

func TestBezierGuide_CorrectStepCount(t *testing.T) {
	cases := []struct {
		name       string
		start, end jitter.Point
		wantPoints int // steps + 1 (includes start and end)
	}{
		{
			name:       "short distance, clamped to 10 steps",
			start:      jitter.Point{X: 0, Y: 0},
			end:        jitter.Point{X: 15, Y: 0},
			wantPoints: 11,
		},
		{
			name:       "medium distance, 40 steps",
			start:      jitter.Point{X: 0, Y: 0},
			end:        jitter.Point{X: 120, Y: 0},
			wantPoints: 41,
		},
		{
			name:       "long distance, clamped to 80 steps",
			start:      jitter.Point{X: 0, Y: 0},
			end:        jitter.Point{X: 300, Y: 0},
			wantPoints: 81,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			guide := jitter.BezierGuide(tc.start, tc.end)
			points := drainGuide(guide, 10000)
			if len(points) != tc.wantPoints {
				t.Fatalf("got %d points, want %d", len(points), tc.wantPoints)
			}
		})
	}
}

func TestBezierGuide_OneSideOfLine(t *testing.T) {
	start := jitter.Point{X: 0, Y: 0}
	end := jitter.Point{X: 300, Y: 0}
	guide := jitter.BezierGuide(start, end)

	points := drainGuide(guide, 10000)

	// Cross-product: (end-start) x (p-start). For horizontal line, this is just p.Y.
	// All intermediate points should have the same sign.
	dx := end.X - start.X
	dy := end.Y - start.Y

	var positive, negative int
	for i := 1; i < len(points)-1; i++ {
		cross := dx*(points[i].Y-start.Y) - dy*(points[i].X-start.X)
		if cross > 1e-9 {
			positive++
		} else if cross < -1e-9 {
			negative++
		}
	}

	if positive > 0 && negative > 0 {
		t.Fatalf("points on both sides of the line: %d positive, %d negative cross products", positive, negative)
	}
}

func TestBezierGuide_SmoothVelocity(t *testing.T) {
	start := jitter.Point{X: 0, Y: 0}
	end := jitter.Point{X: 300, Y: 0}
	guide := jitter.BezierGuide(start, end)

	points := drainGuide(guide, 10000)
	if len(points) < 3 {
		t.Fatal("too few points to check velocity profile")
	}

	// Compute velocity (distance between consecutive points) for each step.
	velocities := make([]float64, len(points)-1)
	for i := 1; i < len(points); i++ {
		velocities[i-1] = dist(points[i-1], points[i])
	}

	// Velocity at start and end should be lower than somewhere in the middle.
	// The Bezier curve accelerates away from start and decelerates into end.
	startVel := velocities[0]
	endVel := velocities[len(velocities)-1]

	midStart := len(velocities) / 4
	midEnd := 3 * len(velocities) / 4
	maxMidVel := 0.0
	for i := midStart; i <= midEnd; i++ {
		if velocities[i] > maxMidVel {
			maxMidVel = velocities[i]
		}
	}

	if maxMidVel <= startVel {
		t.Fatalf("mid velocity %.2f should exceed start velocity %.2f", maxMidVel, startVel)
	}
	if maxMidVel <= endVel {
		t.Fatalf("mid velocity %.2f should exceed end velocity %.2f", maxMidVel, endVel)
	}
}

// helpers: dist and drainGuide are defined in windmouse_test.go
