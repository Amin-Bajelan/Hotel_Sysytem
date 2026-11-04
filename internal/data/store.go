package data

import (
	"errors"
	"fmt"
	"strings"
	"sync"

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
