package simulation

import (
	"testing"

	"smart-hotel-dispatch/internal/location"
)

func TestRemainingRoute(t *testing.T) {
	e, _ := newTestEngine(t)
	e.Step(1)

	r := e.RemainingRoute("S2") // Sara walks to room 102
	if len(r) < 2 {
		t.Fatalf("expected a route with at least 2 points, got %v", r)
	}
	if r[len(r)-1] != location.RoomPoint(102) {
		t.Errorf("route must end at room 102, ends at %+v", r[len(r)-1])
	}
	if e.RemainingRoute("S9") != nil {
		t.Error("unknown employee must have no route")
	}
}