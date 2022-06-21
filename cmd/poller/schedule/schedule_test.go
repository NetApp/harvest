package schedule

import (
	"testing"
	"time"
)

func setupSchedule() *Schedule {
	s := New()
	err1 := s.NewTaskString("counter", "1200s", nil, false, "")
	err2 := s.NewTaskString("data", "180s", nil, false, "")
	err3 := s.NewTaskString("instance", "600s", nil, false, "")
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
		if "counter" == task.Name {
			if task.interval != 1.2e+12 {
				t.Errorf("expected = %v, got %d", 1.2e+12, task.interval)
			}
		} else if "data" == task.Name {
			if task.interval != 1.8e+11 {
				t.Errorf("expected = %b, got %b", 1.8e+11, task.interval)
			}
		} else if "instance" == task.Name {
			if task.interval != 6e+11 {
				t.Errorf("expected = %b, got %b", 6e+11, task.interval)
			}
		}
	}
}
