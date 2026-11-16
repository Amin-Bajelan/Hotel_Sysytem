package reports

import (
	"testing"
	"time"

	"smart-hotel-dispatch/internal/models"
)

func TestCompute(t *testing.T) {
	base := time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC)
	staff := []models.Staff{
		{Status: models.StatusAvailable},
		{Status: models.StatusBusy},
		{Status: models.StatusBusy},
	}
	tasks := []models.Task{
		{Status: models.StatusPending, CreatedAt: base},
		{Status: models.StatusAssigned, CreatedAt: base, AssignedAt: base.Add(60 * time.Second), Distance: 10},
		{Status: models.StatusCompleted, CreatedAt: base, AssignedAt: base.Add(120 * time.Second), Distance: 20.5},
	}
	st := Compute(staff, tasks)

	if st.TotalStaff != 3 || st.AvailableStaff != 1 || st.BusyStaff != 2 {
		t.Errorf("staff counts wrong: %+v", st)
	}
	if st.TotalTasks != 3 || st.Pending != 1 || st.Assigned != 1 || st.Completed != 1 {
		t.Errorf("task counts wrong: %+v", st)
	}
	if st.AvgResponseSec != 90 {
		t.Errorf("avg response: got %v, want 90", st.AvgResponseSec)
	}
	if st.TotalDistance != 30.5 {
		t.Errorf("distance: got %v, want 30.5", st.TotalDistance)
	}
}

func TestComputeEmpty(t *testing.T) {
	st := Compute(nil, nil)
	if st.AvgResponseSec != 0 || st.TotalTasks != 0 {
		t.Errorf("empty input should give zeros: %+v", st)
	}
}