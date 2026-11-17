package simulation

import (
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"smart-hotel-dispatch/internal/data"
	"smart-hotel-dispatch/internal/dispatch"
	"smart-hotel-dispatch/internal/location"
	"smart-hotel-dispatch/internal/models"
)

const (
	tickInterval = 100 * time.Millisecond
	maxEvents    = 200
)

type phase int

const (
	phaseMoving phase = iota
	phaseWorking
)

// worker is the simulation state of one employee who has a task.
type worker struct {
	taskID    string
	phase     phase
	mover     *Mover
	remaining float64 // seconds of work left
}

// Event is one line of the simulation log.
type Event struct {
	Time time.Time // simulated time
	Text string
}

// Engine runs the simulation. Engine methods are safe to call from any goroutine.
// Store listeners run while the engine is locked, so a listener must never call
// engine methods directly (in the GUI use fyne.Do).
type Engine struct {
	mu       sync.Mutex
	store    *data.Store
	strategy dispatch.Strategy
	simNow   atomic.Int64 // simulated time, unix nanoseconds
	workers  map[string]*worker
	events   []Event
	running  bool
	speed    float64
	stop     chan struct{}
}

func NewEngine(store *data.Store, strategy dispatch.Strategy) *Engine {
	e := &Engine{
		store:    store,
		strategy: strategy,
		workers:  map[string]*worker{},
		speed:    1,
	}
	e.simNow.Store(time.Now().UnixNano())
	store.SetClock(e.SimNow)
	return e
}

// SimNow returns the simulated time. It never blocks.
func (e *Engine) SimNow() time.Time { return time.Unix(0, e.simNow.Load()) }

func (e *Engine) Running() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.running
}

func (e *Engine) Speed() float64 {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.speed
}

func (e *Engine) SetSpeed(multiplier float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if multiplier > 0 {
		e.speed = multiplier
	}
}

func (e *Engine) SetStrategy(s dispatch.Strategy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.strategy = s
}

// Events returns a copy of the log (oldest first).
func (e *Engine) Events() []Event {
	e.mu.Lock()
	defer e.mu.Unlock()
	out := make([]Event, len(e.events))
	copy(out, e.events)
	return out
}

// ---- control ----

func (e *Engine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.running {
		return
	}
	e.running = true
	e.stop = make(chan struct{})
	go e.loop(e.stop)
}

func (e *Engine) Resume() { e.Start() }

func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.stopLocked()
}

func (e *Engine) stopLocked() {
	if e.running {
		e.running = false
		close(e.stop)
	}
}

// Reset stops the simulation and replaces all data with ds.
func (e *Engine) Reset(ds *data.Dataset) {
	e.mu.Lock()
	e.stopLocked()
	e.workers = map[string]*worker{}
	e.events = nil
	e.simNow.Store(time.Now().UnixNano())
	e.mu.Unlock()
	e.store.Reset(ds)
}

// Step advances the simulation by dt simulated seconds.
// Tests and the Day 8 comparison call this directly.
func (e *Engine) Step(dt float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.step(dt)
}

func (e *Engine) loop(stop <-chan struct{}) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	last := time.Now()
	for {
		select {
		case <-stop:
			return
		case t := <-ticker.C:
			real := t.Sub(last).Seconds()
			last = t
			if real > maxRealStep {
				real = maxRealStep
			}
			e.mu.Lock()
			select {
			case <-stop: // paused while we waited for the lock
				e.mu.Unlock()
				return
			default:
			}
			e.step(real * BaseTimeScale * e.speed)
			e.mu.Unlock()
		}
	}
}

// ---- one simulation step (caller holds e.mu) ----

func (e *Engine) step(dt float64) {
	if dt <= 0 {
		return
	}
	now := time.Unix(0, e.simNow.Add(int64(dt*1e9)))
	e.dispatchPending(now)
	e.progress(dt, now)
}

func (e *Engine) logf(at time.Time, format string, args ...any) {
	e.events = append(e.events, Event{Time: at, Text: fmt.Sprintf(format, args...)})
	if len(e.events) > maxEvents {
		e.events = e.events[len(e.events)-maxEvents:]
	}
}

func hasPending(tasks []models.Task) bool {
	for _, t := range tasks {
		if t.Status == models.StatusPending {
			return true
		}
	}
	return false
}

func indexStaff(staff []models.Staff, id string) int {
	for i := range staff {
		if staff[i].ID == id {
			return i
		}
	}
	return -1
}

func indexTask(tasks []models.Task, id string) int {
	for i := range tasks {
		if tasks[i].ID == id {
			return i
		}
	}
	return -1
}

// dispatchPending assigns pending tasks to available employees.
func (e *Engine) dispatchPending(now time.Time) {
	tasks := e.store.Tasks()
	if !hasPending(tasks) {
		return
	}
	for _, d := range dispatch.Dispatch(e.strategy, tasks, e.store.Staff()) {
		if d.Chosen == nil {
			continue
		}
		c := d.Chosen
		ti := indexTask(tasks, d.TaskID)
		if ti < 0 {
			continue
		}
		if err := e.store.AssignTask(d.TaskID, c.Staff.ID, c.Distance, now); err != nil {
			continue
		}
		route := location.Route(location.StaffPoint(c.Staff), location.RoomPoint(tasks[ti].RoomNumber))
		e.workers[c.Staff.ID] = &worker{taskID: d.TaskID, phase: phaseMoving, mover: NewMover(route)}
		e.logf(now, "%s assigned to %s [%s]: %s", d.TaskID, c.Staff.Name, d.Strategy, d.Reason)
	}
}

// progress moves employees and advances the task lifecycle:
// Assigned -> (walking) -> In Progress -> (working) -> Completed.
func (e *Engine) progress(dt float64, now time.Time) {
	if len(e.workers) == 0 {
		return
	}
	ids := make([]string, 0, len(e.workers))
	for id := range e.workers {
		ids = append(ids, id)
	}
	sort.Strings(ids) // fixed order keeps runs deterministic

	e.store.Update(func(staff []models.Staff, tasks []models.Task) {
		for _, id := range ids {
			w := e.workers[id]
			si, ti := indexStaff(staff, id), indexTask(tasks, w.taskID)
			if si < 0 || ti < 0 {
				delete(e.workers, id)
				continue
			}
			s, t := &staff[si], &tasks[ti]

			switch w.phase {
			case phaseMoving:
				w.mover.Advance(s.Speed * dt)
				s.Floor, s.X, s.Y = w.mover.Pos.Floor, w.mover.Pos.X, w.mover.Pos.Y
				if w.mover.Arrived() {
					w.phase = phaseWorking
					w.remaining = float64(t.EstimatedDuration) * 60
					t.Status = models.StatusInProgress
					e.logf(now, "%s arrived at room %d, started %s", s.Name, t.RoomNumber, t.ID)
				}
			case phaseWorking:
				w.remaining -= dt
				if w.remaining <= 0 {
					t.Status = models.StatusCompleted
					t.CompletedAt = now
					s.Status = models.StatusAvailable
					s.CurrentTaskID = ""
					e.logf(now, "%s completed %s in room %d", s.Name, t.ID, t.RoomNumber)
					delete(e.workers, id)
				}
			}
		}
	})
}

// RemainingRoute returns the rest of a walking employee's route,
// starting at the current position. It returns nil when the employee
// is not walking.
func (e *Engine) RemainingRoute(staffID string) []location.Point {
	e.mu.Lock()
	defer e.mu.Unlock()
	w, ok := e.workers[staffID]
	if !ok || w.phase != phaseMoving || w.mover.Arrived() {
		return nil
	}
	out := []location.Point{w.mover.Pos}
	out = append(out, w.mover.Route[w.mover.Next:]...)
	return out
}