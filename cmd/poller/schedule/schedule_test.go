package schedule

import (
	"github.com/netapp/harvest/v2/assert"
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
		switch task.Name {
		case "counter":
			assert.Equal(t, task.interval, 1.2e+12)
		case "data":
			assert.Equal(t, task.interval, 1.8e+11)
		case "instance":
			assert.Equal(t, task.interval, 6e+11)
		}
	}
}

func TestNewTaskString(t *testing.T) {
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
			assert.Nil(t, err)

			// create another task with runNow set to tc.runNow and jitter 0
			err = s.NewTaskString(tc.name+"2", tc.interval, 0, f, tc.runNow, "testID2")
			assert.Nil(t, err)

			assert.Equal(t, len(s.tasks), 2)

			// Check that the task with jitter should run after the task with jitter set to 0
			assert.False(t, s.tasks[0].timer.Before(s.tasks[1].timer))
		})
	}
}
