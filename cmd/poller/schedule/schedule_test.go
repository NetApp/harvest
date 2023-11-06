package schedule

import (
	"github.com/netapp/harvest/v2/pkg/tree"
	"reflect"
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

func TestLoadTasks(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    []*TaskSpec
		wantErr bool
	}{
		{
			name: "with jitter",
			yaml: `
schedule:
  - counter: 1h
  - instance: 5m
  - data: 3m

jitter:
  - instance: 3m
  - data: 1m
`,
			want: []*TaskSpec{
				{Name: "counter", Interval: "1h", Jitter: ""},
				{Name: "instance", Interval: "5m", Jitter: "3m"},
				{Name: "data", Interval: "3m", Jitter: "1m"},
			},
		},

		{
			name: "no jitter",
			yaml: `
schedule:
 - counter:  20m
 - instance: 10m
 - data:      3m`,
			want: []*TaskSpec{
				{Name: "counter", Interval: "20m", Jitter: ""},
				{Name: "instance", Interval: "10m", Jitter: ""},
				{Name: "data", Interval: "3m", Jitter: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := tree.LoadYaml([]byte(tt.yaml))
			if err != nil {
				t.Errorf("LoadTasks() error = %v", err)
				return
			}
			got, err := LoadTasks(node)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTasks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadTasks() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
