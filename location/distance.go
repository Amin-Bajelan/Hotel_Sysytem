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

// Distance: same floor = straight line; different floors = go to the
// elevator, pay a cost per floor, then walk to the destination.
func Distance(a, b Point) float64 {
	if a.Floor == b.Floor {
		return Euclidean(a.X, a.Y, b.X, b.Y)
	}
	floors := math.Abs(float64(a.Floor - b.Floor))
	return Euclidean(a.X, a.Y, ElevatorX, ElevatorY) +
		floors*FloorChangeCost +
		Euclidean(ElevatorX, ElevatorY, b.X, b.Y)
}