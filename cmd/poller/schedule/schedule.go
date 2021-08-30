/*
	Copyright NetApp Inc, 2021 All rights reserved

	Package Schedule provides a mechanism to run tasks at fixed time interals.
	It is intended to be used by collectors, but can be used by any other
	package as well. Tasks can be dynamically pointed to the poll functions
	of the collector. (This why poll functions of collectors are public and
	have the same signature).

	At least one task should be added to Schedule before it can be used.
	Tasks are yielded in the same order as added (FIFO). The intervals of tasks
	can be safely changed any time.

   	Create Schedule:
    	- Initialize empty Schedule with New(),
    	- Add tasks with NewTask() or NewTaskString(),
        the task is marked as due immediately!

	Use Schedule (usually in a closed loop):
    	- iterate over all tasks with GetTasks()
        	- check if it's time to run the task with IsDue(task)
        	- run the task with task.Run() or run "manually" with task.Start()
	   - suspend the goroutine until another task is due Sleep()/Wait()

	The Schedule can enter standByMode when a critical task has failed. In this
	scenario all tasks are stalled until the critical task has succeeded. This is
	sometimes useful when a target system is unreachable and we have to wait
	until it's up again.

	Schedule is meant to be used by at most one goroutine and is not
	concurrent-safe.
*/
package schedule

import (
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"time"
)

// Task represents a scheduled task
type task struct {
	Name       string                         // name of the task
	interval   time.Duration                  // the schedule interval
	timer      time.Time                      // last time task was executed
	foo        func() (*matrix.Matrix, error) // pointer to the function that executes the task
	identifier string                         // optional additional information about schedule i.e. collector name
}

// Start marks the task as started by updating timer
// Use this method if you are executing the task yourself and you need to register
// when task started. If the task has a pointer to the executing function, use
// Run() instead.
func (t *task) Start() {
	t.timer = time.Now()
}

// Run marks the task as started and executes it
func (t *task) Run() (*matrix.Matrix, error) {
	t.Start()
	return t.foo()
}

// GetDuration tells duration of executing the task
// it assumes that the task just completed
func (t *task) GetDuration() time.Duration {
	return time.Since(t.timer)
}

// GetInterval tells the scheduled interval of the task
func (t *task) GetInterval() time.Duration {
	return t.interval
}

// NextDue tells time until the task is due
func (t *task) NextDue() time.Duration {
	return t.interval - time.Since(t.timer)
}

// IsDue tells whether it's time to run the task
func (t *task) IsDue() bool {
	return t.NextDue() <= 0
}

// Schedule contains a collection of tasks and the current state of the schedule
type Schedule struct {
	tasks          []*task                  // list of tasks that Schedule needs to run
	standByMode    bool                     // if true, Schedule waitsfor a stalled task
	standByTask    *task                    // stalled task in standByMode
	standByTimer   time.Time                // timer for stalled task in standByMode
	cachedInterval map[string]time.Duration // normal interval of the stalled tasks
}

// New creates and initializes an empty Schedule.
func New() *Schedule {
	s := Schedule{}
	s.tasks = make([]*task, 0)
	s.standByMode = false
	s.cachedInterval = make(map[string]time.Duration)
	return &s
}

// IsStandBy tells if schedule is in IsStandBy.
// If false, Schedule is in "normal" mode
func (s *Schedule) IsStandBy() bool {
	return s.standByMode
}

// SetStandByMode initializes StandbyMode: Schedule will suspend all tasks until
// the critical task t has succeeded. The temporary interval i will be used for
// the task until Schedule recovers to normal mode.
func (s *Schedule) SetStandByMode(t *task, i time.Duration) {
	for _, x := range s.tasks {
		if x.Name == t.Name {
			s.standByTask = t
			t.interval = i
			t.timer = time.Now()
			s.standByMode = true
			return
		}
	}
	panic("invalid task: " + t.Name)
}

// Recover undoes StandbyMode and restores normal state of the Schedule
func (s *Schedule) Recover() {

	if s.standByMode {
		for _, t := range s.tasks {
			if interval, ok := s.cachedInterval[t.Name]; ok {
				t.interval = interval
			}
			// reset timer of the critical task, assume that it just completed
			if t.Name == s.standByTask.Name {
				t.timer = time.Now()
				// all the other tasks that were suspended need to run asap
			} else {
				t.timer = time.Now().Add(-t.interval)
			}
		}
		//s.cachedInterval = nil
		s.standByTask = nil
		s.standByMode = false
		return
	}
	panic("recover in non-standByMode")
}

// NewTask creates new task named n with interval i. If f is not nil, f will be called
// to execute task when task.Run() is called. Task name n should be unique. Interval i
// should be positive.
// The order in which tasks are added is maintained: GetTasks() will
// return tasks in FIFO order.
func (s *Schedule) NewTask(n string, i time.Duration, f func() (*matrix.Matrix, error), runNow bool, identifier string) error {
	if s.GetTask(n) == nil {
		if i > 0 {
			t := &task{Name: n, interval: i, foo: f, identifier: identifier}
			s.cachedInterval[n] = t.interval // remember normal interval of task
			if runNow {
				t.timer = time.Now().Add(-i) // set to run immediately
			} else {
				t.timer = time.Now().Add(0) // run after interval has elapsed
			}
			s.tasks = append(s.tasks, t)
			return nil
		}
		return errors.New(errors.INVALID_PARAM, "interval :"+i.String())
	}
	return errors.New(errors.INVALID_PARAM, "duplicate task :"+n)
}

// NewTaskString creates a new task, the interval is parsed from string i
func (s *Schedule) NewTaskString(n, i string, f func() (*matrix.Matrix, error), runNow bool, identifier string) error {
	if d, err := time.ParseDuration(i); err == nil {
		return s.NewTask(n, d, f, runNow, identifier)
	} else {
		return err
	}
}

// GetTasks returns scheduled tasks
func (s *Schedule) GetTasks() []*task {
	if !s.standByMode {
		return s.tasks
	}
	return []*task{s.standByTask}
}

// GetTask returns the task named n or nil if it doesn't exist
func (s *Schedule) GetTask(n string) *task {
	for _, t := range s.tasks {
		if t.Name == n {
			return t
		}
	}
	return nil
}

// Sleep sleeps until at least one task is due
func (s *Schedule) Sleep() {
	time.Sleep(s.NextDue())
}

// Wait returns a blocking channel until a task is due
// Similar to Sleep(), but the goroutine can wait for other jobs as well
func (s *Schedule) Wait() <-chan time.Time {
	return time.After(s.NextDue())
}

// NextDue tells duration until at least one task is due
// If no tasks are scheduled, NextDue returns an arbitrary long duration
// (This is useful for collectors that run background jobs and need to
// wait indefinitely).
func (s *Schedule) NextDue() time.Duration {

	if s.standByMode {
		return s.standByTask.NextDue()
	}
	d := 1000000 * time.Hour

	for _, t := range s.tasks {
		if due := t.NextDue(); due < d {
			d = due
		}
	}

	return d
}
