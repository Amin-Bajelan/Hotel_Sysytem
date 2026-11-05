package data

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"smart-hotel-dispatch/internal/models"
)

type Store struct {
	mu    sync.RWMutex
	staff []models.Staff
	tasks []models.Task
}

func NewStore(ds *Dataset) *Store {
	s := &Store{}
	s.staff = append(s.staff, ds.Staff...)
	s.tasks = append(s.tasks, ds.Tasks...)
	return s
}

func (s *Store) Staff() []models.Staff {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Staff, len(s.staff))
	copy(out, s.staff)
	return out
}

func (s *Store) Tasks() []models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Task, len(s.tasks))
	copy(out, s.tasks)
	return out
}

func (s *Store) hasStaffID(id string) bool {
	for _, st := range s.staff {
		if st.ID == id {
			return true
		}
	}
	return false
}

// AddStaff validates the employee, assigns an ID and stores it.
func (s *Store) AddStaff(st models.Staff) (models.Staff, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(st.Name) == "" {
		return st, errors.New("name is required")
	}
	if len(st.Skills) == 0 {
		return st, errors.New("select at least one skill")
	}
	if st.Speed <= 0 {
		return st, errors.New("speed must be greater than zero")
	}

	n := len(s.staff) + 1
	for s.hasStaffID(fmt.Sprintf("S%d", n)) {
		n++
	}
	st.ID = fmt.Sprintf("S%d", n)
	st.Status = models.StatusAvailable
	st.CurrentTaskID = ""

	s.staff = append(s.staff, st)
	return st, nil
}
func (s *Store) hasTaskID(id string) bool {
	for _, t := range s.tasks {
		if t.ID == id {
			return true
		}
	}
	return false
}

func (s *Store) AddTask(room int, tt models.TaskType, p models.Priority, minutes int) (models.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	floor, n := room/100, room%100
	if floor < 1 || floor > 3 || n < 1 || n > 7 {
		return models.Task{}, errors.New("room must be 101-107, 201-207 or 301-307")
	}
	if minutes <= 0 {
		return models.Task{}, errors.New("duration must be greater than zero")
	}

	id := len(s.tasks) + 1
	for s.hasTaskID(fmt.Sprintf("T%d", id)) {
		id++
	}
	t := models.Task{
		ID:                fmt.Sprintf("T%d", id),
		RoomNumber:        room,
		Type:              tt,
		Priority:          p,
		Status:            models.StatusPending,
		CreatedAt:         time.Now(),
		EstimatedDuration: minutes,
	}
	s.tasks = append(s.tasks, t)
	return t, nil
}