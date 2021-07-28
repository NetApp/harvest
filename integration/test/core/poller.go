package core

import "strings"

type Poller struct {
	DataCenter string
	Poller     string
	Pid        string
	PromPort   string
	Status     string
	metricUrl  string
}

func (p *Poller) New(dataCenter string,
	poller string, pid string, promPort string,
	status string) {
	p.Pid = strings.TrimSpace(pid)
	p.PromPort = strings.TrimSpace(promPort)
	p.Status = status
	p.DataCenter = strings.TrimSpace(dataCenter)
}

func (p *Poller) MetricUrl() string {
	return "http://localhost:" + strings.TrimSpace(p.PromPort) + "/metrics"
}
