package dispatch

import (
	"fmt"
	"sort"

	"smart-hotel-dispatch/internal/models"
)

const noCandidate = "no available employee with the required skill"

var priorityRank = map[models.Priority]int{
	models.PriorityLow:    0,
	models.PriorityNormal: 1,
	models.PriorityHigh:   2,
	models.PriorityUrgent: 3,
}

type Decision struct {
	TaskID     string
	Strategy   string
	Chosen     *Candidate // nil = task could not be assigned
	Candidates []Candidate
	Reason     string
}

type Strategy interface {
	Name() string
	Select(task models.Task, staff []models.Staff) Decision
}

// Eligible keeps only available employees who have the skill and can move.
func Eligible(task models.Task, staff []models.Staff) []models.Staff {
	var out []models.Staff
	for _, s := range staff {
		if s.Status == models.StatusAvailable && s.Speed > 0 && s.HasSkill(task.Type) {
			out = append(out, s)
		}
	}
	return out
}

// PendingByPriority returns pending tasks: most urgent first, then oldest first.
func PendingByPriority(tasks []models.Task) []models.Task {
	var out []models.Task
	for _, t := range tasks {
		if t.Status == models.StatusPending {
			out = append(out, t)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		ri, rj := priorityRank[out[i].Priority], priorityRank[out[j].Priority]
		if ri != rj {
			return ri > rj
		}
		return out[i].CreatedAt.Before(out[j].CreatedAt)
	})
	return out
}

// ---- First Available ----

type FirstAvailable struct{}

func (FirstAvailable) Name() string { return "First Available" }

func (f FirstAvailable) Select(task models.Task, staff []models.Staff) Decision {
	d := Decision{TaskID: task.ID, Strategy: f.Name()}
	el := Eligible(task, staff)
	if len(el) == 0 {
		d.Reason = noCandidate
		return d
	}
	c := newCandidate(task, el[0])
	d.Chosen = &c
	d.Candidates = []Candidate{c}
	d.Reason = fmt.Sprintf("first eligible employee in list order (travel %.0fs)", c.TravelTime)
	return d
}

// ---- Weighted Scoring ----

type Weighted struct{ W Weights }

func (Weighted) Name() string { return "Weighted Scoring" }

func (w Weighted) Select(task models.Task, staff []models.Staff) Decision {
	cands := ScoreCandidates(task, Eligible(task, staff), w.W)
	d := Decision{TaskID: task.ID, Strategy: w.Name(), Candidates: cands}
	if len(cands) == 0 {
		d.Reason = noCandidate
		return d
	}
	best := cands[0]
	d.Chosen = &best
	d.Reason = fmt.Sprintf("score %.2f: travel %.0fs, distance %.1f, handled %d tasks, %d skills",
		best.Score, best.TravelTime, best.Distance, best.Staff.TasksHandled, len(best.Staff.Skills))
	if len(cands) > 1 {
		d.Reason += fmt.Sprintf("; runner-up %s (%.2f)", cands[1].Staff.Name, cands[1].Score)
	}
	return d
}

// Dispatch assigns all pending tasks on a COPY of the staff list,
// so nobody is double-booked and the real data is not changed.
func Dispatch(s Strategy, tasks []models.Task, staff []models.Staff) []Decision {
	work := make([]models.Staff, len(staff))
	copy(work, staff)

	var out []Decision
	for _, t := range PendingByPriority(tasks) {
		d := s.Select(t, work)
		out = append(out, d)
		if d.Chosen == nil {
			continue
		}
		for i := range work {
			if work[i].ID == d.Chosen.Staff.ID {
				work[i].Status = models.StatusBusy
				work[i].CurrentTaskID = t.ID
				work[i].TasksHandled++
			}
		}
	}
	return out
}