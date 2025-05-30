package ps

import (
	ps2 "github.com/netapp/harvest/v2/third_party/tklauser/ps"
	"regexp"
	"strings"
)

type PollerStatus struct {
	Name          string
	Status        Status
	Pid           int
	ProfilingPort string
	PromPort      string
}

type Status string

const (
	StatusStopped        Status = "stopped"
	StatusStoppingFailed Status = "stopping failed"
	StatusRunning        Status = "running"
	StatusNotRunning     Status = "not running"
	StatusKilled         Status = "killed"
	StatusAlreadyExited  Status = "already exited"
	StatusDisabled       Status = "disabled"
)

var profRegex = regexp.MustCompile(`--profiling (\d+)`)
var promRegex = regexp.MustCompile(`--promPort (\d+)`)

func GetPollerStatuses() ([]PollerStatus, error) {

	result := make([]PollerStatus, 0)

	processes, err := ps2.Processes()
	if err != nil {
		return nil, err
	}

	for _, p := range processes {
		if !strings.HasSuffix(p.Command(), "poller") {
			continue
		}
		args := p.ExecutableArgs()

		name := ""
		for i, arg := range args {
			if arg == "--poller" && i+1 < len(args) {
				name = args[i+1]
				break
			}
		}

		if name == "" {
			continue
		}

		s := PollerStatus{
			Name:   name,
			Pid:    p.PID(),
			Status: "running",
		}

		line := strings.Join(args, " ")

		promMatches := promRegex.FindStringSubmatch(line)
		if len(promMatches) > 0 {
			s.PromPort = promMatches[1]
		}
		profMatches := profRegex.FindStringSubmatch(line)
		if len(profMatches) > 0 {
			s.ProfilingPort = profMatches[1]
		}
		result = append(result, s)
	}

	return result, nil
}
