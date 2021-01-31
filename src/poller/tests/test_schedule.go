//package schedule_test
package main

import (
	"testing"
	"fmt"
	"time"
	"poller/schedule"
)

func TestAddTask(t *testing.T) {
	s := schedule.New()
	err := s.AddTask("jobA", time.Duration(10), nil)
	if err != nil {
		t.Errorf("FAIL - AddTask: %s\n", err)
	} else {
		t.Logf("OK - AddTask")
		fmt.Println(s.String("jobA"))
	}
}

func TestAddTaskString(t *testing.T) {
	s := schedule.New()
	err := s.AddTaskString("jobB", "2.2s", nil)
	if err != nil {
		t.Errorf("AddTaskString failed: %s\n", err)
	} else {
		t.Logf("OK - AddTaskString")
		fmt.Println(s.String("jobB"))
	}
}

func TestChangeTaskString(t *testing.T) {
	s := schedule.New()
	err := s.AddTaskString("jobX", "20s", nil)
	if err != nil {
		t.Errorf("FAIL - AddTaskString: %s\n", err)
	} else {
		fmt.Println("Before change:")
		fmt.Println(s.String("jobX"))
	}

	err = s.ChangeTaskString("NoSuchTask", "5s")
	if err == nil {
		t.Errorf("FAIL - ChangeTaskString: invalid name but no error")
	}

	err = s.ChangeTaskString("jobX", "5s")
	if err != nil {
		t.Errorf("Fail - ChangeTaskString: %v", err)
	} else {
		fmt.Println("After change:")
		fmt.Println(s.String("jobX"))
	}
}


func main() {
	t := &testing.T{}
	TestAddTask(t)
	TestAddTaskString(t)
	TestChangeTaskString(t)
}