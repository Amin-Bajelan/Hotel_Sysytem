package simulation

import (
	"testing"

	"smart-hotel-dispatch/internal/location"
)

func TestMoverArrivesAtDestination(t *testing.T) {
	a := location.Point{Floor: 0, X: 10, Y: 35}
	b := location.RoomPoint(104)
	route := location.Route(a, b)
	total := location.PathLength(route)

	m := NewMover(route)
	m.Advance(total / 2)
	if m.Arrived() {
		t.Fatal("must not arrive after half of the distance")
	}
	m.Advance(total)
	if !m.Arrived() {
		t.Fatal("must arrive after the full distance")
	}
	if m.Pos != b {
		t.Errorf("final position: got %+v, want %+v", m.Pos, b)
	}
}

func TestMoverCrossesFloors(t *testing.T) {
	a := location.Point{Floor: 2, X: 30, Y: 25}
	b := location.Point{Floor: 0, X: 30, Y: 25}
	route := location.Route(a, b)
	total := location.PathLength(route)

	m := NewMover(route)
	steps := 0
	for !m.Arrived() && steps < 1000 {
		m.Advance(0.5)
		steps++
	}
	if !m.Arrived() {
		t.Fatal("never arrived")
	}
	if m.Pos.Floor != 0 {
		t.Errorf("final floor: got %d, want 0", m.Pos.Floor)
	}
	if float64(steps)*0.5 < total-0.5 {
		t.Errorf("arrived too early: %d steps for a route of length %v", steps, total)
	}
}

func TestMoverAlreadyThere(t *testing.T) {
	a := location.Point{Floor: 1, X: 5, Y: 5}
	m := NewMover(location.Route(a, a))
	if !m.Arrived() {
		t.Error("a zero-length route must be arrived immediately")
	}
}