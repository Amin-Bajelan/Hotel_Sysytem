package dispatch

import (
	"testing"
	"time"

	"smart-hotel-dispatch/internal/models"
)

func cleaner(id string, floor int, x, y float64) models.Staff {
	return models.Staff{
		ID: id, Name: id, Floor: floor, X: x, Y: y, Speed: 1.5,
		Skills: []models.TaskType{models.RoomCleaning},
		Status: models.StatusAvailable,
	}
}

func cleaningTask(id string, p models.Priority, created time.Time) models.Task {
	return models.Task{
		ID: id, RoomNumber: 101, Type: models.RoomCleaning, Priority: p,
		Status: models.StatusPending, CreatedAt: created, EstimatedDuration: 20,
	}
}

func TestEligibleFilters(t *testing.T) {
	busy := cleaner("busy", 0, 0, 0)
	busy.Status = models.StatusBusy
	tech := cleaner("tech", 0, 0, 0)
	tech.Skills = []models.TaskType{models.Maintenance}
	ok := cleaner("ok", 0, 0, 0)

	got := Eligible(cleaningTask("T1", models.PriorityNormal, time.Now()),
		[]models.Staff{busy, tech, ok})
	if len(got) != 1 || got[0].ID != "ok" {
		t.Errorf("expected only 'ok', got %v", got)
	}
}

func TestWeightedPicksNearest(t *testing.T) {
	staff := []models.Staff{cleaner("far", 0, 80, 5), cleaner("near", 0, 10, 35)}
	d := Weighted{W: DefaultWeights()}.Select(
		cleaningTask("T1", models.PriorityNormal, time.Now()), staff)
	if d.Chosen == nil || d.Chosen.Staff.ID != "near" {
		t.Errorf("expected 'near', got %+v", d.Chosen)
	}
}

func TestWeightedBalancesWorkload(t *testing.T) {
	a := cleaner("a", 0, 10, 35)
	a.TasksHandled = 5
	b := cleaner("b", 0, 10, 35)
	d := Weighted{W: DefaultWeights()}.Select(
		cleaningTask("T1", models.PriorityNormal, time.Now()), []models.Staff{a, b})
	if d.Chosen == nil || d.Chosen.Staff.ID != "b" {
		t.Errorf("expected 'b' (less workload), got %+v", d.Chosen)
	}
}

func TestFirstAvailableUsesListOrder(t *testing.T) {
	staff := []models.Staff{cleaner("far", 0, 80, 5), cleaner("near", 0, 10, 35)}
	d := FirstAvailable{}.Select(
		cleaningTask("T1", models.PriorityNormal, time.Now()), staff)
	if d.Chosen == nil || d.Chosen.Staff.ID != "far" {
		t.Errorf("expected 'far' (first in list), got %+v", d.Chosen)
	}
}

func TestNoEligibleStaff(t *testing.T) {
	d := Weighted{W: DefaultWeights()}.Select(
		cleaningTask("T1", models.PriorityNormal, time.Now()), nil)
	if d.Chosen != nil {
		t.Error("expected no assignment")
	}
}

func TestPendingByPriority(t *testing.T) {
	now := time.Now()
	tasks := []models.Task{
		cleaningTask("low", models.PriorityLow, now),
		cleaningTask("urgent", models.PriorityUrgent, now.Add(time.Minute)),
	}
	got := PendingByPriority(tasks)
	if len(got) != 2 || got[0].ID != "urgent" {
		t.Errorf("urgent task must come first, got %v", got)
	}
}

func TestDispatchNoDoubleBooking(t *testing.T) {
	now := time.Now()
	tasks := []models.Task{
		cleaningTask("T1", models.PriorityNormal, now),
		cleaningTask("T2", models.PriorityNormal, now.Add(time.Minute)),
	}
	ds := Dispatch(Weighted{W: DefaultWeights()}, tasks, []models.Staff{cleaner("only", 0, 10, 35)})
	if len(ds) != 2 {
		t.Fatalf("expected 2 decisions, got %d", len(ds))
	}
	if ds[0].Chosen == nil {
		t.Error("first task should be assigned")
	}
	if ds[1].Chosen != nil {
		t.Error("second task must stay unassigned (employee is busy)")
	}
}