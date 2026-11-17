package location

import "math"

// CorridorY is the horizontal line that all walking routes follow.
const CorridorY = ElevatorY

// Route returns the waypoints from a to b: walk to the corridor,
// follow it, and use the elevator when the floors differ.
func Route(a, b Point) []Point {
	if a == b {
		return []Point{a}
	}
	pts := []Point{a}
	add := func(p Point) {
		if pts[len(pts)-1] != p {
			pts = append(pts, p)
		}
	}
	if a.Floor == b.Floor {
		add(Point{Floor: a.Floor, X: a.X, Y: CorridorY})
		add(Point{Floor: a.Floor, X: b.X, Y: CorridorY})
		add(b)
		return pts
	}
	add(Point{Floor: a.Floor, X: a.X, Y: CorridorY})
	add(Point{Floor: a.Floor, X: ElevatorX, Y: ElevatorY})
	add(Point{Floor: b.Floor, X: ElevatorX, Y: ElevatorY})
	add(Point{Floor: b.Floor, X: b.X, Y: CorridorY})
	add(b)
	return pts
}

// PathLength sums walking distance; each floor change costs FloorChangeCost.
func PathLength(pts []Point) float64 {
	total := 0.0
	for i := 1; i < len(pts); i++ {
		p, q := pts[i-1], pts[i]
		if p.Floor == q.Floor {
			total += Euclidean(p.X, p.Y, q.X, q.Y)
		} else {
			total += FloorChangeCost * math.Abs(float64(p.Floor-q.Floor))
		}
	}
	return total
}