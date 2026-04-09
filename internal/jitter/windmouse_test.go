package jitter_test

import (
	"math"
	"testing"

	"github.com/pavlomaksymov/in-network-explorer/internal/jitter"
)

// drainGuide collects all points from a guide function up to maxIter.
func drainGuide(guide jitter.GuideFunc, maxIter int) []jitter.Point {
	var points []jitter.Point
	for i := 0; i < maxIter; i++ {
		p, done := guide()
		points = append(points, p)
		if done {
			return points
		}
	}
	return points
}

func dist(a, b jitter.Point) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}

func TestWindMouseGuide_ReachesTarget(t *testing.T) {
	start := jitter.Point{X: 100, Y: 100}
	dst := jitter.Point{X: 300, Y: 300}
	guide := jitter.WindMouseGuide(start, dst)

	points := drainGuide(guide, 10000)
	last := points[len(points)-1]
	if d := dist(last, dst); d > 2 {
		t.Fatalf("final point %v is %.2fpx from target %v, want <= 2px", last, d, dst)
	}
}

func TestWindMouseGuide_ProducesMultiplePoints(t *testing.T) {
	start := jitter.Point{X: 0, Y: 0}
	dst := jitter.Point{X: 200, Y: 0}
	guide := jitter.WindMouseGuide(start, dst)

	points := drainGuide(guide, 10000)
	if len(points) <= 5 {
		t.Fatalf("got %d points for 200px distance, want > 5", len(points))
	}
}

func TestWindMouseGuide_VelocityClipped(t *testing.T) {
	start := jitter.Point{X: 0, Y: 0}
	dst := jitter.Point{X: 500, Y: 500}
	guide := jitter.WindMouseGuide(start, dst)

	points := drainGuide(guide, 10000)
	prev := start
	for i, p := range points {
		step := dist(prev, p)
		if step > 20 {
			t.Fatalf("step %d: moved %.2fpx, want <= 20px", i, step)
		}
		prev = p
	}
}

func TestWindMouseGuide_NonLinear(t *testing.T) {
	start := jitter.Point{X: 0, Y: 0}
	dst := jitter.Point{X: 800, Y: 0}
	guide := jitter.WindMouseGuide(start, dst)

	points := drainGuide(guide, 10000)
	maxDeviation := 0.0
	for _, p := range points {
		if d := math.Abs(p.Y); d > maxDeviation {
			maxDeviation = d
		}
	}
	if maxDeviation <= 3 {
		t.Fatalf("max Y deviation = %.2fpx, want > 3px (path too linear)", maxDeviation)
	}
}

func TestWindMouseGuide_Terminates(t *testing.T) {
	start := jitter.Point{X: 50, Y: 50}
	dst := jitter.Point{X: 250, Y: 250}
	guide := jitter.WindMouseGuide(start, dst)

	for i := 0; i < 10000; i++ {
		_, done := guide()
		if done {
			return
		}
	}
	t.Fatal("guide did not terminate within 10000 iterations")
}
