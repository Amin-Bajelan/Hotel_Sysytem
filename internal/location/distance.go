package location

import "math"

type Point struct {
	Floor int
	X, Y  float64
}

const (
	ElevatorX       = 25.0
	ElevatorY       = 25.0
	FloorChangeCost = 15.0
)

func Euclidean(ax, ay, bx, by float64) float64 {
	return math.Hypot(ax-bx, ay-by)
}

// Distance is the walking length of the route between two points.
func Distance(a, b Point) float64 {
	return PathLength(Route(a, b))
}