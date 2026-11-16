package data

import (
	"testing"
	"time"

	"smart-hotel-dispatch/internal/models"
)

func TestAssignAndComplete(t *testing.T) {
	ds, _ := LoadSample()
	s := NewStore(ds)
	now := time.Now()

	if err := s.AssignTask("T1", "S1", 12.5, now); err != nil {
		t.Fatal(err)
	}
	if err := s.AssignTask("T1", "S2", 1, now); err == nil {
		t.Error("an already assigned task must be rejected")
	}
	if err := s.AssignTask("T2", "S1", 1, now); err == nil {
		t.Error("a busy employee must be rejected")
	}
	if err := s.CompleteTask("T1", now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	for _, st := range s.Staff() {
		if st.ID == "S1" && st.Status != models.StatusAvailable {
			t.Error("employee must be available after completing the task")
		}
	}
}

func TestSubscribeNotifies(t *testing.T) {
	ds, _ := LoadSample()
	s := NewStore(ds)
	calls := 0
	s.Subscribe(func() {
		_ = s.Staff() // must not deadlock
		calls++
	})
	if _, err := s.AddTask(101, models.RoomCleaning, models.PriorityNormal, 10); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Errorf("listener calls: got %d, want 1", calls)
	}
	if _, err := s.AddTask(999, models.RoomCleaning, models.PriorityNormal, 10); err == nil {
		t.Error("invalid room must fail")
	}
	if calls != 1 {
		t.Error("a failed change must not notify")
	}
}