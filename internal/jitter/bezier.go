package jitter

import (
	"math"
	"math/rand"
)

// BezierGuide returns a guide function that produces a cubic Bezier curve from
// start to end. Control points are placed on one side of the start-end line to
// avoid S-curves. Steps = int(distance/3), clamped to [10, 80]. Uses ease-in/
// ease-out parameter mapping for natural acceleration and deceleration.
func BezierGuide(start, end Point) GuideFunc {
	dx := end.X - start.X
	dy := end.Y - start.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	steps := int(dist / 3.0)
	if steps < 10 {
		steps = 10
	}
	if steps > 80 {
		steps = 80
	}

	// Perpendicular direction (unit vector).
	perpX := -dy / dist
	perpY := dx / dist

	// Both control points on the same side.
	side := 1.0
	if rand.Float64() < 0.5 {
		side = -1.0
	}

	offset1 := (0.1 + rand.Float64()*0.3) * dist * side
	offset2 := (0.1 + rand.Float64()*0.3) * dist * side

	cp1 := Point{
		X: start.X + dx/3 + perpX*offset1,
		Y: start.Y + dy/3 + perpY*offset1,
	}
	cp2 := Point{
		X: start.X + 2*dx/3 + perpX*offset2,
		Y: start.Y + 2*dy/3 + perpY*offset2,
	}

	// Pre-compute all points with ease-in/ease-out parameter mapping.
	points := make([]Point, steps+1)
	for i := 0; i <= steps; i++ {
		// Smoothstep: start slow, speed up in middle, slow down at end.
		s := float64(i) / float64(steps)
		t := s * s * (3 - 2*s)
		u := 1 - t
		points[i] = Point{
			X: u*u*u*start.X + 3*u*u*t*cp1.X + 3*u*t*t*cp2.X + t*t*t*end.X,
			Y: u*u*u*start.Y + 3*u*u*t*cp1.Y + 3*u*t*t*cp2.Y + t*t*t*end.Y,
		}
	}

	idx := 0
	return func() (Point, bool) {
		p := points[idx]
		idx++
		return p, idx > steps
	}
}
