package schedule

import (
	"github.com/netapp/harvest/v2/pkg/matrix"
	"testing"
	"time"
)

func setupSchedule() *Schedule {
	s := New()
	err1 := s.NewTaskString("counter", "1200s", 0, nil, false, "")
	err2 := s.NewTaskString("data", "180s", 0, nil, false, "")
	err3 := s.NewTaskString("instance", "600s", 0, nil, false, "")
	if err1 != nil || err2 != nil || err3 != nil {
		panic("error creating tasks")
	}
	return s
}

func setStandByTasks(s *Schedule) {
	retryDelay := 1024
	for _, task := range s.GetTasks() {
		s.SetStandByMode(task, time.Duration(retryDelay)*time.Second)
	}
}

func TestSchedule_Recover(t *testing.T) {
	s := setupSchedule()
	setStandByTasks(s)
	s.Recover()
	for _, task := range s.GetTasks() {
		if task.Name == "counter" {
			if task.interval != 1.2e+12 {
				t.Errorf("expected = %v, got %d", 1.2e+12, task.interval)
			}
		} else if task.Name == "data" {
			if task.interval != 1.8e+11 {
				t.Errorf("expected = %b, got %b", 1.8e+11, task.interval)
			}
		} else if task.Name == "instance" {
			if task.interval != 6e+11 {
				t.Errorf("expected = %b, got %b", 6e+11, task.interval)
			}
		}
	}
}

func TestNewTaskString_RunNow_True(t *testing.T) {
	// Define a dummy function for the task
	f := func() (map[string]*matrix.Matrix, error) {
		return nil, nil
	}

	testCases := []struct {
		name     string
		interval string
		jitter   time.Duration
		runNow   bool
	}{
		{"test1", "10s", 5 * time.Second, true},
		{"test2", "10s", 5 * time.Second, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := &Schedule{
				tasks:          []*Task{},
				cachedInterval: map[string]time.Duration{},
			}

			// Create a task with runNow set to tc.runNow
			err := s.NewTaskString(tc.name+"1", tc.interval, tc.jitter, f, tc.runNow, "testID1")
			if err != nil {
				t.Errorf("NewTaskString returned an error: %v", err)
			}

			// create another task with runNow set to tc.runNow and jitter 0
			err = s.NewTaskString(tc.name+"2", tc.interval, 0, f, tc.runNow, "testID2")
			if err != nil {
				t.Errorf("NewTaskString returned an error: %v", err)
			}

			if len(s.tasks) != 2 {
				t.Errorf("Expected 2 tasks, got %d", len(s.tasks))
			}

			// Check that the task with jitter should run after the task with jitter set to 0
			if s.tasks[0].timer.Before(s.tasks[1].timer) {
				t.Errorf("Expected first task to run after second task, got first task timer '%s' and second task timer '%s'", s.tasks[0].timer, s.tasks[1].timer)
			}
		})
	}
}
