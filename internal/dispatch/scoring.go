package dispatch

import (
	"math"
	"sort"

	"smart-hotel-dispatch/internal/location"
	"smart-hotel-dispatch/internal/models"
)

type Weights struct {
	Time     float64
	Workload float64
	Skill    float64
}

func DefaultWeights() Weights {
	return Weights{Time: 0.6, Workload: 0.25, Skill: 0.15}
}

var priorityFactor = map[models.Priority]float64{
	models.PriorityLow:    0.8,
	models.PriorityNormal: 1.0,
	models.PriorityHigh:   1.3,
	models.PriorityUrgent: 1.6,
}

type Candidate struct {
	Staff         models.Staff
	Distance      float64
	TravelTime    float64 // seconds
	ETA           float64 // seconds: travel + task duration
	TimeScore     float64
	WorkloadScore float64
	SkillScore    float64
	Score         float64
}

func newCandidate(task models.Task, s models.Staff) Candidate {
	d := location.Distance(location.StaffPoint(s), location.RoomPoint(task.RoomNumber))
	tt := location.TravelTime(d, s.Speed)
	return Candidate{
		Staff:      s,
		Distance:   d,
		TravelTime: tt,
		ETA:        tt + float64(task.EstimatedDuration)*60,
	}
}

// lowIsBetter maps the smallest value to 1 and the largest to 0.
func lowIsBetter(v, min, max float64) float64 {
	if max-min < 1e-9 {
		return 1
	}
	return 1 - (v-min)/(max-min)
}

func ScoreCandidates(task models.Task, eligible []models.Staff, w Weights) []Candidate {
	cands := make([]Candidate, len(eligible))
	for i, s := range eligible {
		cands[i] = newCandidate(task, s)
	}
	if len(cands) == 0 {
		return cands
	}

	minT, maxT := cands[0].TravelTime, cands[0].TravelTime
	minW, maxW := float64(cands[0].Staff.TasksHandled), float64(cands[0].Staff.TasksHandled)
	for _, c := range cands[1:] {
		minT, maxT = math.Min(minT, c.TravelTime), math.Max(maxT, c.TravelTime)
		h := float64(c.Staff.TasksHandled)
		minW, maxW = math.Min(minW, h), math.Max(maxW, h)
	}

	pf, ok := priorityFactor[task.Priority]
	if !ok {
		pf = 1
	}
	wt := w.Time * pf
	total := wt + w.Workload + w.Skill
	if total <= 0 {
		total = 1
	}

	for i := range cands {
		c := &cands[i]
		c.TimeScore = lowIsBetter(c.TravelTime, minT, maxT)
		c.WorkloadScore = lowIsBetter(float64(c.Staff.TasksHandled), minW, maxW)
		c.SkillScore = 1 / float64(len(c.Staff.Skills))
		c.Score = (wt*c.TimeScore + w.Workload*c.WorkloadScore + w.Skill*c.SkillScore) / total
	}

	sort.SliceStable(cands, func(i, j int) bool {
		if math.Abs(cands[i].Score-cands[j].Score) > 1e-9 {
			return cands[i].Score > cands[j].Score
		}
		return cands[i].Staff.ID < cands[j].Staff.ID
	})
	return cands
}