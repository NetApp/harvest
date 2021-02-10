package schedule

import (
    "fmt"
    "time"
    "goharvest2/poller/errors"
)

/*
    Schedule helps to run tasks on fixed time interals. At least
    one task should be added to Schedule before it can be used.
    Tasks are yielded in the same order as they were added (FIFO).
    The intervals of tasks can be safely changed any time.

    As a convinience, tasks can be mapped to interfaces 
    (functions or methods), allowing to run all tasks in a single
    loop.

    General workflow of intializing a Schedule:
        - Initialize empty Schedule with New(),
        - Add tasks with AddTask() or AddTaskString(),
          the task is marked as due immediately.

    Workflow of using the Schedule (usually in a loop):
        - GetTasks() gives list of tasks 
            - IsDue(task) tells if it's time to run the task
            - Start(task) starts a timer for the task
                (task is run here...)
            - Stop(task) tells duration of completing the task
        - Sleep()/Wait() stop the goroutine until another task is due

    Schedule is not concurrent-safe and is meant to be used by at 
    most one goroutine.

*/

type Task struct {
    Name string
    interval time.Duration
    timer time.Time
    foo func() (*matrix.Matrix, error)
}

func (t *Task) Start() {
    t.timer = time.Now()
}

func (t *Task) Run() (*matrix.Matrix, error) {
    t.Start()
    return t.foo()
}

func (t *Task) Runtime() time.Duration {
    return time.Since(t.timer)
}

func (t *Task) GetInterval() time.Duration {
    return t.interval
}

func (t *Task) NextDue() time.Duration {
    return t.interval - time.Since(t.timer)   
}

type Schedule struct {
    tasks []*Task
    standby_mode bool
    standby_task Task
    cached_interval time.Duration
    standby_timer time.Time
}

// Create and initialize and empty Schedule. It is not safe
// to run the Schedule until tasks have been added
func New() *Schedule {
	s := Schedule{}
    s.tasks = make([]*Task, 0)
    s.standby_mode = false
	return &s
}

func (s *Schedule) IsStandBy() bool {
    return s.standby_mode
}

// StandbyMode will make the schedule stop all tasks except one
// and will set a temporary extended interval for one task
func (s *Schedule) SetStandByMode(t string, i time.Duration) {
    if task, ok := s.tasks[t]; ok {
        s.standby_mode = true
        s.standby_task = task
        s.cached_interval = task.interval // remember normal interval of task
        task.interval = i
        task.timer = time.Now()
    } else {
        panic("invalid task: " + task_name)
    }
}

// Undo StandbyMode. We assume that stalled task was success, so schedule
// the other tasks to run on next poll
func (s *Schedule) Recover() {

    if task, ok := s.tasks[s.standby_task]; ok {
        task.interval = s.cached_interval
        task.timer = time.Now()
    }

    for _, task := range s.tasks {
        if task != s.standby_task {
            task.timer = time.Now().Add(-task.interval)
        }
    }
    s.cached_interval = nil
    s.standby_task = nil
    s.standby_mode = false
}

// Add new task to Schedule and set to run immediately.
//  @task       name of the task,
//  @interval   interval or frequency of the task,
//  @foo        pointer to method or function that performs
//              the task, can be nil if not useful.
// The order in which tasks are added is maintained: GetTasks() will
// return tasks in FIFO order.
func (s *Schedule) AddTask(t string, i time.Duration, f func() (*matrix.Matrix, error)) error {
    if i > 0 {
        task := &Task{Name: t, interval: i, foo: f}
        task.timer = time.Now().Add(-i) // set to run immediately
        s.tasks = append(s.tasks, task)
        return nil
    }
    return errors.New("invalid interval :" + i.String())
}

// Same as AddTask, but interval is parsed from string
func (s *Schedule) AddTaskString(t, i string, f func() (*matrix.Matrix, error)) error {
    if d, err := time.ParseDuration(i); err == nil {
        return s.AddTask(t, d, foo)
    } else {
        return err
    }
}

// Change the interval of a task, safe to do this even
// in the middle of running a task.
func (s *Schedule) SetInterval(t string, i time.Duration) error {
    for _, task := range s.tasks {
        if task.Name == t {
            if i > 0 {
                task.interval = i
                return nil
            } else {
                return errors.New(errors.ERR_SCHEDULE, "invalid interval :" + i.String())
            }
        }
    }
    return errors.New(errors.ERR_SCHEDULE, "invalid task: " + t)
}

// Same as SetInterval, but interval is parsed from string
func (s *Schedule) SetIntervalString(t, i string) error {
    if d, err := time.ParseDuration(i); err == nil {
        return s.SetInterval(t, d)
    } else {
        return err
    }
}

// Return tasks in Schedule
func (s *Schedule) GetTasks() []*Task {
    if !s.standby_mode {
        return s.tasks
    }
    return []*Task{s.standby_task}
}

// Tell if it's time to run a task
// @t     name of the task
func (s *Schedule) IsDue(t string) (bool) {
    if task, ok := s.tasks[t]; ok {
        // normal schedule
        if !s.standby_mode {
            return task.nextDue() <= 0
        }
        // standby mode: task can only be due if it's stalled
        if task == s.standby_task {
            return task.nextDue() <= 0
        }
        return false
    }
    panic("invalid task: " + t)
}


// Sleep until a task is due
func (s *Schedule) Sleep() {
    time.Sleep(s.NextDue())
}

// Get blocking channel until a task is due
// Similar to Sleep(), but we can perform other tasks
// while waiting.
func (s *Schedule) Wait() <-chan time.Time {
    return time.After(s.NextDue())
}

// Tells duration until a next earliest task is due.
func (s *Schedule) NextDue() time.Duration {

    if s.standby_mode {
        return s.standby_task.NextDue()
    }

    next_due := s.tasks[0].NextDue()

    for _, task := range s.tasks[1:] {
        if due := task.NextDue(); due < next_due {
            next_due = due
        }
    }

    return next_due
}
