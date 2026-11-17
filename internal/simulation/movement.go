package simulation

import (
	"math"

	"smart-hotel-dispatch/internal/location"
)

// Mover walks along a route (see location.Route). Walking costs 1 distance
// unit per unit of length; changing floors costs location.FloorChangeCost
// per floor, so total cost equals location.PathLength(route).
type Mover struct {
	Pos   location.Point
	Route []location.Point
	Next  int     // index of the next waypoint
	carry float64 // elevator progress already spent
}

func NewMover(route []location.Point) *Mover {
	m := &Mover{Route: route, Next: 1}
	if len(route) > 0 {
		m.Pos = route[0]
	}
	return m
}

func (m *Mover) Arrived() bool { return m.Next >= len(m.Route) }

// Advance moves along the route by up to dist units.
func (m *Mover) Advance(dist float64) {
	for dist > 1e-9 && !m.Arrived() {
		target := m.Route[m.Next]

		if m.Pos.Floor != target.Floor {
			cost := location.FloorChangeCost * math.Abs(float64(m.Pos.Floor-target.Floor))
			need := cost - m.carry
			if dist >= need {
				dist -= need
				m.carry = 0
				m.Pos = target
				m.Next++
			} else {
				m.carry += dist
				dist = 0
			}
			continue
		}

		d := location.Euclidean(m.Pos.X, m.Pos.Y, target.X, target.Y)
		if dist >= d {
			dist -= d
			m.Pos = target
			m.Next++
		} else {
			f := dist / d
			m.Pos.X += (target.X - m.Pos.X) * f
			m.Pos.Y += (target.Y - m.Pos.Y) * f
			dist = 0
		}
	}
}