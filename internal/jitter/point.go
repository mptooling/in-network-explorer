package jitter

// Point represents a 2D screen coordinate.
type Point struct {
	X float64
	Y float64
}

// GuideFunc is a mouse movement guide function. It returns the next point and
// true when the movement is complete. Compatible with rod.Mouse.MoveAlong
// after a trivial Point-to-proto.Point conversion in the adapter layer.
type GuideFunc func() (Point, bool)
