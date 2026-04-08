package jitter

import (
	"math"
	"math/rand"
)

// WindMouse parameters.
const (
	wmGravity = 9.0  // G0: pull toward target
	wmWind    = 3.0  // W0: random perturbation magnitude
	wmMaxVel  = 15.0 // M0: maximum velocity
	wmDamping = 12.0 // D0: distance below which wind dampens
)

// WindMouseGuide returns a guide function that produces human-like mouse
// movement from start to dst using the WindMouse algorithm (gravity + stochastic
// wind). Compatible with rod.Mouse.MoveAlong via adapter-layer Point conversion.
func WindMouseGuide(start, dst Point) GuideFunc {
	x, y := start.X, start.Y
	vx, vy, wx, wy := 0.0, 0.0, 0.0, 0.0

	return func() (Point, bool) {
		dx := dst.X - x
		dy := dst.Y - y
		dist := math.Sqrt(dx*dx + dy*dy)

		if dist < 1 {
			return dst, true
		}

		// Wind: random perturbations when far, just decay when close.
		if dist >= wmDamping {
			wx = wx/math.Sqrt2 + (rand.Float64()*2-1)*(wmWind/math.Sqrt(5))
			wy = wy/math.Sqrt2 + (rand.Float64()*2-1)*(wmWind/math.Sqrt(5))
		} else {
			wx /= math.Sqrt2
			wy /= math.Sqrt2
		}

		// Gravity: always pulls toward target.
		vx += wx + wmGravity*dx/dist
		vy += wy + wmGravity*dy/dist

		// Clip velocity.
		speed := math.Sqrt(vx*vx + vy*vy)
		maxV := math.Max(3, wmMaxVel*math.Min(1, dist/wmDamping))
		if speed > maxV {
			s := maxV/2 + rand.Float64()*maxV/2
			vx = vx / speed * s
			vy = vy / speed * s
		}

		x += vx
		y += vy
		return Point{X: x, Y: y}, false
	}
}
