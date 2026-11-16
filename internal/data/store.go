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
	mu        sync.RWMutex
	staff     []models.Staff
	tasks     []models.Task
	listeners []func()
}

func NewStore(ds *Dataset) *Store {
	s := &Store{}
	s.staff = append(s.staff, ds.Staff...)
	s.tasks = append(s.tasks, ds.Tasks...)
	return s
}

// Subscribe registers a function that runs after every data change.
func (s *Store) Subscribe(fn func()) {
	s.mu.Lock()
	s.listeners = append(s.listeners, fn)
	s.mu.Unlock()
}

// notify runs OUTSIDE the lock, so listeners may call Staff()/Tasks() safely.
func (s *Store) notify() {
	s.mu.RLock()
	ls := make([]func(), len(s.listeners))
	copy(ls, s.listeners)
	s.mu.RUnlock()
	for _, fn := range ls {
		fn()
	}
}

// mutate runs fn under the write lock, then notifies listeners if it succeeded.
func (s *Store) mutate(fn func() error) error {
	s.mu.Lock()
	err := fn()
	s.mu.Unlock()
	if err == nil {
		s.notify()
	}
	return err
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

// Reset replaces all data (used by the demo "Reset Data" button).
func (s *Store) Reset(ds *Dataset) {
	_ = s.mutate(func() error {
		s.staff = append([]models.Staff(nil), ds.Staff...)
		s.tasks = append([]models.Task(nil), ds.Tasks...)
		return nil
	})
}

// index helpers: call only while holding the lock.
func (s *Store) staffIndex(id string) int {
	for i, st := range s.staff {
		if st.ID == id {
			return i
		}
	}
	return -1
}

func (s *Store) taskIndex(id string) int {
	for i, t := range s.tasks {
		if t.ID == id {
			return i
		}
	}
	return -1
}

func (s *Store) AddStaff(st models.Staff) (models.Staff, error) {
	err := s.mutate(func() error {
		if strings.TrimSpace(st.Name) == "" {
			return errors.New("name is required")
		}
		if len(st.Skills) == 0 {
			return errors.New("select at least one skill")
		}
		if st.Speed <= 0 {
			return errors.New("speed must be greater than zero")
		}
		n := len(s.staff) + 1
		for s.staffIndex(fmt.Sprintf("S%d", n)) >= 0 {
			n++
		}
		st.ID = fmt.Sprintf("S%d", n)
		st.Status = models.StatusAvailable
		st.CurrentTaskID = ""
		s.staff = append(s.staff, st)
		return nil
	})
	return st, err
}

func (s *Store) AddTask(room int, tt models.TaskType, p models.Priority, minutes int) (models.Task, error) {
	var t models.Task
	err := s.mutate(func() error {
		floor, n := room/100, room%100
		if floor < 1 || floor > 3 || n < 1 || n > 7 {
			return errors.New("room must be 101-107, 201-207 or 301-307")
		}
		if minutes <= 0 {
			return errors.New("duration must be greater than zero")
		}
		id := len(s.tasks) + 1
		for s.taskIndex(fmt.Sprintf("T%d", id)) >= 0 {
			id++
		}
		t = models.Task{
			ID:                fmt.Sprintf("T%d", id),
			RoomNumber:        room,
			Type:              tt,
			Priority:          p,
			Status:            models.StatusPending,
			CreatedAt:         time.Now(),
			EstimatedDuration: minutes,
		}
		s.tasks = append(s.tasks, t)
		return nil
	})
	return t, err
}

// AssignTask gives a pending task to an available employee who has the skill.
func (s *Store) AssignTask(taskID, staffID string, distance float64, at time.Time) error {
	return s.mutate(func() error {
		ti, si := s.taskIndex(taskID), s.staffIndex(staffID)
		if ti < 0 {
			return fmt.Errorf("task %s not found", taskID)
		}
		if si < 0 {
			return fmt.Errorf("employee %s not found", staffID)
		}
		if s.tasks[ti].Status != models.StatusPending {
			return fmt.Errorf("task %s is not pending", taskID)
		}
		if s.staff[si].Status != models.StatusAvailable {
			return fmt.Errorf("employee %s is not available", staffID)
		}
		if !s.staff[si].HasSkill(s.tasks[ti].Type) {
			return fmt.Errorf("employee %s lacks the required skill", staffID)
		}
		s.tasks[ti].Status = models.StatusAssigned
		s.tasks[ti].AssignedStaffID = staffID
s.tasks[ti].AssignedAt = at
		s.tasks[ti].Distance = distance
		s.staff[si].Status = models.StatusBusy
		s.staff[si].CurrentTaskID = taskID
		s.staff[si].TasksHandled++
		return nil
	})
}

// CompleteTask finishes an active task and frees the employee.
func (s *Store) CompleteTask(taskID string, at time.Time) error {
	return s.mutate(func() error {
		ti := s.taskIndex(taskID)
		if ti < 0 {
			return fmt.Errorf("task %s not found", taskID)
		}
		t := &s.tasks[ti]
		if t.Status != models.StatusAssigned && t.Status != models.StatusInProgress {
			return fmt.Errorf("task %s is not active", taskID)
		}
		t.Status = models.StatusCompleted
		t.CompletedAt = at
		if si := s.staffIndex(t.AssignedStaffID); si >= 0 {
			s.staff[si].Status = models.StatusAvailable
			s.staff[si].CurrentTaskID = ""
		}
		return nil
	})
}