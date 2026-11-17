package location

import (
	"math"
	"testing"

	"smart-hotel-dispatch/internal/models"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestDistanceSameFloor(t *testing.T) {
	a := Point{Floor: 0, X: 10, Y: 35}
	b := Point{Floor: 0, X: 10, Y: 40}
	// 10 up to the corridor + 15 down to the room
	if got := Distance(a, b); !near(got, 25) {
		t.Errorf("got %v, want 25", got)
	}
}

func TestDistanceSamePoint(t *testing.T) {
	a := Point{Floor: 1, X: 5, Y: 5}
	if got := Distance(a, a); !near(got, 0) {
		t.Errorf("got %v, want 0", got)
	}
}

func TestDistanceElevatorOnly(t *testing.T) {
	a := Point{Floor: 0, X: ElevatorX, Y: ElevatorY}
	b := Point{Floor: 1, X: ElevatorX, Y: ElevatorY}
	if got := Distance(a, b); !near(got, FloorChangeCost) {
		t.Errorf("got %v, want %v", got, FloorChangeCost)
	}
}

func TestDistanceTwoFloors(t *testing.T) {
	a := Point{Floor: 0, X: 10, Y: 25}
	b := Point{Floor: 2, X: 10, Y: 25}
	// 15 to elevator + 2 floors * 15 + 15 back
	if got := Distance(a, b); !near(got, 60) {
		t.Errorf("got %v, want 60", got)
	}
}

func TestRouteCrossFloor(t *testing.T) {
	a := Point{Floor: 0, X: 10, Y: 35}
	b := Point{Floor: 1, X: 58, Y: 40}
	r := Route(a, b)
	if r[0] != a || r[len(r)-1] != b {
		t.Fatalf("route must start at a and end at b: %v", r)
	}
	seenFloor1 := false
	for _, p := range r {
		if p.Floor == 1 {
			seenFloor1 = true
		}
	}
	if !seenFloor1 {
		t.Error("route must pass through floor 1")
	}
}

func TestDefaultHotel(t *testing.T) {
	h := DefaultHotel()
	if h.Floors != 3 {
		t.Errorf("floors: got %d, want 3", h.Floors)
	}
	perFloor := map[int]int{}
	for _, r := range h.Rooms {
		perFloor[r.Floor]++
	}
	for f := 0; f < 3; f++ {
		if perFloor[f] != 7 {
			t.Errorf("floor %d rooms: got %d, want 7", f, perFloor[f])
		}
	}

	all := append(append([]models.Location{}, h.Rooms...), h.Facilities...)
	for _, l := range all {
		if l.X-l.W/2 < 0 || l.X+l.W/2 > 100 || l.Y-l.H/2 < 0 || l.Y+l.H/2 > 52 {
			t.Errorf("%s (floor %d) is outside the map", l.Name, l.Floor)
		}
	}

	has := map[string]bool{}
	for _, f := range h.Facilities {
		has[f.Name+"/"+string(rune('0'+f.Floor))] = true
	}
	for _, want := range []string{"Lobby/0", "Reception/0", "Restaurant/0",
		"Elevator/0", "Elevator/1", "Elevator/2", "Staircase/0", "Staircase/1", "Staircase/2"} {
		if !has[want] {
			t.Errorf("missing %s", want)
		}
	}
}