package reports

import "smart-hotel-dispatch/internal/models"

type Stats struct {
	TotalStaff     int
	AvailableStaff int
	BusyStaff      int

	TotalTasks int
	Pending    int
	Assigned   int // Assigned + In Progress
	Completed  int
	Cancelled  int

	AvgResponseSec float64 // average of (AssignedAt - CreatedAt)
	TotalDistance  float64
}

func Compute(staff []models.Staff, tasks []models.Task) Stats {
	var st Stats
	st.TotalStaff = len(staff)
	for _, s := range staff {
		switch s.Status {
		case models.StatusAvailable:
			st.AvailableStaff++
		case models.StatusBusy:
			st.BusyStaff++
		}
	}

	st.TotalTasks = len(tasks)
	var sum float64
	n := 0
	for _, t := range tasks {
		switch t.Status {
		case models.StatusPending:
			st.Pending++
		case models.StatusAssigned, models.StatusInProgress:
			st.Assigned++
		case models.StatusCompleted:
			st.Completed++
		case models.StatusCancelled:
			st.Cancelled++
		}
		if !t.AssignedAt.IsZero() {
			sum += t.AssignedAt.Sub(t.CreatedAt).Seconds()
			n++
		}
		st.TotalDistance += t.Distance
	}
	if n > 0 {
		st.AvgResponseSec = sum / float64(n)
	}
	return st
}