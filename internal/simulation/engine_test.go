package simulation

import (
	"testing"
	"time"

	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/internal/dispatch"
	"smart-hotel-dispatch/internal/models"
)

func newTestEngine(t *testing.T) (*Engine, *data.Store) {
	t.Helper()
	ds, err := data.LoadSample()
	if err != nil {
		t.Fatal(err)
	}
	store := data.NewStore(ds)
	return NewEngine(store, dispatch.Weighted{W: dispatch.DefaultWeights()}), store
}

func allDone(tasks []models.Task) bool {
	for _, t := range tasks {
		if t.Status != models.StatusCompleted && t.Status != models.StatusCancelled {
			return false
		}
	}
	return true
}

func TestEngineCompletesAllTasks(t *testing.T) {
	e, store := newTestEngine(t)

	finished := false
	for i := 0; i < 4000 && !finished; i++ {
		e.Step(1)
		finished = allDone(store.Tasks())
	}
	if !finished {
		t.Fatal("tasks did not finish in time")
	}

	for _, task := range store.Tasks() {
		if task.AssignedAt.IsZero() || task.CompletedAt.IsZero() {
			t.Errorf("%s: missing assignment or completion time", task.ID)
		}
		if task.CompletedAt.Before(task.AssignedAt) {
			t.Errorf("%s: completed before it was assigned", task.ID)
		}
		if task.Distance <= 0 {
			t.Errorf("%s: distance must be positive", task.ID)
		}
	}
	for _, s := range store.Staff() {
		if s.Status != models.StatusAvailable || s.CurrentTaskID != "" {
			t.Errorf("%s must be free at the end", s.Name)
		}
	}
	for _, ev := range e.Events() {
		t.Log(ev.Time.Format("15:04:05"), ev.Text)
	}
}

func TestEngineMovesEmployee(t *testing.T) {
	e, store := newTestEngine(t)
	e.Step(1)

	for _, s := range store.Staff() {
		if s.ID != "S2" { // Sara gets the urgent food delivery
			continue
		}
		if s.Status != models.StatusBusy {
			t.Errorf("Sara should be busy, got %s", s.Status)
		}
		if s.X == 50 && s.Y == 10 {
			t.Error("Sara did not move")
		}
	}
}

func TestEngineNoDoubleBooking(t *testing.T) {
	e, store := newTestEngine(t)
	for i := 0; i < 3; i++ { // only Ali can clean rooms
		if _, err := store.AddTask(102, models.RoomCleaning, models.PriorityNormal, 5); err != nil {
			t.Fatal(err)
		}
	}

	for sec := 0; sec < 8000; sec++ {
		e.Step(1)
		active := map[string]int{}
		for _, task := range store.Tasks() {
			if task.Status == models.StatusAssigned || task.Status == models.StatusInProgress {
				active[task.AssignedStaffID]++
			}
		}
		for id, n := range active {
			if n > 1 {
				t.Fatalf("employee %s has %d active tasks at once", id, n)
			}
		}
		if allDone(store.Tasks()) {
			return
		}
	}
	t.Fatal("tasks did not finish in time")
}

func TestEngineClockUsedForNewTasks(t *testing.T) {
	e, store := newTestEngine(t)
	e.Step(100)
	task, err := store.AddTask(103, models.RoomCleaning, models.PriorityLow, 5)
	if err != nil {
		t.Fatal(err)
	}
	if !task.CreatedAt.Equal(e.SimNow()) {
		t.Errorf("task created at %v, simulation clock is %v", task.CreatedAt, e.SimNow())
	}
}

func TestEngineReset(t *testing.T) {
	e, store := newTestEngine(t)
	e.Step(120)
	if len(e.Events()) == 0 {
		t.Fatal("expected some events before reset")
	}

	ds, err := data.LoadSample()
	if err != nil {
		t.Fatal(err)
	}
	e.Reset(ds)

	for _, task := range store.Tasks() {
		if task.Status != models.StatusPending {
			t.Errorf("%s must be pending after reset, got %s", task.ID, task.Status)
		}
	}
	for _, s := range store.Staff() {
		if s.Status != models.StatusAvailable {
			t.Errorf("%s must be available after reset", s.Name)
		}
	}
	if len(e.Events()) != 0 {
		t.Error("event log must be empty after reset")
	}
}

func TestEngineStartPause(t *testing.T) {
	e, _ := newTestEngine(t)
	e.SetSpeed(10)
	start := e.SimNow()

	e.Start()
	time.Sleep(400 * time.Millisecond)
	e.Pause()

	t1 := e.SimNow()
	if t1.Sub(start) < 5*time.Second {
		t.Errorf("simulation clock advanced only %v", t1.Sub(start))
	}
	time.Sleep(300 * time.Millisecond)
	if !e.SimNow().Equal(t1) {
		t.Error("clock must not advance while paused")
	}
	if e.Running() {
		t.Error("engine should not be running after Pause")
	}
}