package schedule

import (
    "fmt"
    "time"
    "errors"
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

type Schedule struct {
    tasks []string
    foos map[string]interface{}
    intervals map[string]time.Duration
    timers map[string]time.Time
    standby_mode bool
    standby_task string
    standby_interval time.Duration
    standby_timer time.Time
}

// Create and initialize and empty Schedule. It is not safe
// to run the Schedule until tasks have been added
func New() *Schedule {
	s := Schedule{}
    s.tasks = make([]string, 0)
    s.foos = make(map[string]interface{})
    s.intervals = make(map[string]time.Duration)
    s.timers = make(map[string]time.Time)
    s.standby_mode = false
	return &s
}

func (s *Schedule) IsStandbyMode() bool {
    return s.standby_mode
}

// StandbyMode will make the schedule stop all tasks except one
// and will set a temporary extended interval for one task
func (s *Schedule) SetStandbyMode(task string, interval time.Duration) {
    s.standby_mode = true
    s.standby_task = task
    s.standby_interval = interval
    s.standby_timer = time.Now()
}

// Undo StandbyMode
func (s *Schedule) UnsetStandbyMode() {
    s.standby_mode = false
    for _, task := range s.tasks {
        if task == s.standby_task {
            s.timers[task] = time.Now()
        } else {
            s.timers[task] = time.Now().Add(-s.getInterval(task))
        }
    }
}

// Add new task to Schedule and set to run immediately.
//  @task       name of the task,
//  @interval   interval or frequency of the task,
//  @foo        pointer to method or function that performs
//              the task, can be nil if not useful.
// The order in which tasks are added is maintained: GetTasks() will
// return tasks in FIFO order.
func (s *Schedule) AddTask(task string, interval time.Duration, foo interface{}) error {
    if interval > 0 {
        s.tasks = append(s.tasks, task)
        s.foos[task] = foo
        s.intervals[task] = interval
        s.timers[task] = time.Now().Add(-interval) // set to run immediately
        return nil
    }
    return errors.New("invalid interval :" + interval.String())
}

// Same as AddTask, but interval is parsed from string
func (s *Schedule) AddTaskString(task, interval string, foo interface{}) error {
    d, err := time.ParseDuration(interval)
    if err != nil {
        return err
    }
    return s.AddTask(task, d, foo)
}

// Format details about task into string. Meant for debugging.
func (s *Schedule) String(task string) string {
    details := `
        Task: %s
          interface    => %v
          interval     => %s
          last started => %s
          next due in  => %s
    `
    str := fmt.Sprintf(
        details,
        task,
        s.getFoo(task),
        s.getInterval(task).String(),
        time.Since( s.getTimer(task) ).String(),
        s.getNextDue(task),
    )
    return str
}

// Change the interval of a task, safe to do this even
// in the middle of running a task.
func (s *Schedule) ChangeTask(task string, interval time.Duration) error {
    if foo, ok := s.foos[task]; !ok {
        return errors.New("invalid task name: " + task)
    } else {
        return s.AddTask(task, interval, foo)
    }
}

// Same as ChangeTask, but interval is parsed from string
func (s *Schedule) ChangeTaskString(task, interval string) error {
    if foo, ok := s.foos[task]; !ok {
        return errors.New("invalid task name: " + task)
    } else {
        return s.AddTaskString(task, interval, foo)
    }
}

// Return list of task names
func (s *Schedule) GetTasks() []string {
    return s.tasks
}


func (s *Schedule) GetInterval(task string) time.Duration {
    if interval, ok := s.intervals[task]; ok {
        return interval
    }
    panic("invalid task: " + task)
}

// Tell if it's time to run a task
// @task     name of the task
// @foo      interface that performs the task
// 
// If task was added with nil interface, then calling foo
// will crash the program.
func (s *Schedule) IsDue(task string) (bool) {
    if !s.standby_mode {
        return s.getNextDue(task) <= 0
    } 
    
    if task == s.standby_task {
        return (s.standby_interval - time.Since(s.standby_timer)) <= 0
    }

    return false
}

// Start timer for the task
func (s *Schedule) Start(name string) {
    if !s.standby_mode {
        s.timers[name] = time.Now()
    } else {
        s.standby_timer = time.Now()
    }
}

// Tell duration of running task
func (s *Schedule) Stop(name string) time.Duration {
    if !s.standby_mode {
        started, _ := s.timers[name]
        return time.Since(started)
    }
    return time.Since(s.standby_timer)
}

// Sleep until a task is due
func (s *Schedule) Sleep() {
    _, next_due := s.earliestNextDue()
    time.Sleep(next_due)
}

func (s *Schedule) SleepDuration() time.Duration {
    _, next_due := s.earliestNextDue()
    return next_due
}

// Get blocking channel until a task is due
// Similar to Sleep(), but we can perform other tasks
// while waiting.
func (s *Schedule) Wait() <-chan time.Time {
    _, next_due := s.earliestNextDue()
    return time.After(next_due)
}

/* private methods */

// tell the interval of the task
func (s *Schedule) getInterval(task string) time.Duration {
    if interval, ok := s.intervals[task]; ok {
        return interval
    }
    // requesting unknown task = bug/typo in program
    panic("invalid task: " + task)
}

// tell when was task last started
func (s *Schedule) getTimer(task string) time.Time {
    if timer, ok := s.timers[task]; ok {
        return timer
    }
    panic("invalid task: " + task)
}

func (s *Schedule) getFoo(task string) interface{} {
    if foo, ok := s.foos[task]; ok {
        return foo
    }
    panic("invalid task: " + task)
}

// tell how much time is left until task is due
func (s *Schedule) getNextDue(task string) time.Duration {
    return s.getInterval(task) - time.Since( s.getTimer(task) )
}

// Tells name of the task and duration until
// a new task is due.
func (s *Schedule) earliestNextDue() (string, time.Duration) {

    if s.standby_mode {
        return s.standby_task, s.standby_interval - time.Since(s.standby_timer)
    }

    next_task := s.tasks[0]
    next_due := s.getNextDue(next_task)

    for _, task := range s.tasks[1:] {
        if due := s.getNextDue(task); due < next_due {
            next_due = due
            next_task = task
        }
    }

    return next_task, next_due
}