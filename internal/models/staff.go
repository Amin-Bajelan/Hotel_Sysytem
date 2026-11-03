package models

type StaffStatus string

const (
	StatusAvailable StaffStatus = "Available"
	StatusBusy      StaffStatus = "Busy"
)

type Staff struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Role          string      `json:"role"`
	Skills        []TaskType  `json:"skills"`
	Floor         int         `json:"floor"`
	X             float64     `json:"x"`
	Y             float64     `json:"y"`
	Speed         float64     `json:"speed"`
	Status        StaffStatus `json:"status"`
	CurrentTaskID string      `json:"current_task_id"`
}

func (s Staff) HasSkill(t TaskType) bool {
	for _, sk := range s.Skills {
		if sk == t {
			return true
		}
	}
	return false
}
