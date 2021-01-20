package schedule

import "time"

type Schedule struct {
	Interval time.Duration
	Started time.Time
}

func New(interval int) *Schedule {
	duration := time.Duration(interval) * time.Second
	sch := Schedule{ Interval: duration }
	return &sch
}

func (s *Schedule) Start() {
	s.Started = time.Now()
}

func (s *Schedule) Pause() time.Duration {
	elapsed := time.Since(s.Started)
	delta := elapsed - s.Interval
	time.Sleep(-delta)
	return delta
}
