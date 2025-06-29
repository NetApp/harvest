package installer

import (
	"errors"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/errs"
	"log/slog"
	"os"
	"os/exec"
	"time"
)

type Grafana struct {
	image string
}

func (g *Grafana) Init(image string) {
	g.image = image
}

func (g *Grafana) Install() bool {
	g.image = "grafana/grafana:8.1.8"
	slog.Info("Grafana image : " + g.image)
	imageName := "grafana"
	err := docker.StopContainers(imageName)
	if err != nil {
		slog.Warn("Error while stopping Grafana container", slog.Any("err", err))
	}
	cmd := exec.Command("docker", "run", "-d", "--name", "grafana", "-e", "GF_LOG_LEVEL=debug", "-p", cmds.GrafanaPort+":"+cmds.GrafanaPort, g.image) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	errs.PanicIfNotNil(err)
	waitCount := 0
	maxWaitCount := 3
	for waitCount < maxWaitCount {
		waitCount++
		time.Sleep(1 * time.Minute)
		if cmds.IsURLReachable("http://localhost:" + cmds.GrafanaPort) {
			return true
		}
		slog.Info("Grafana is not yet reachable.", slog.Int("waitCount", waitCount), slog.Int("maxWaitCount", maxWaitCount))
	}
	slog.Info("Reached maximum wait count. Grafana failed to start")
	return false
}

func (g *Grafana) Upgrade() bool {
	errs.PanicIfNotNil(errors.New("not supported"))
	return false
}

func (g *Grafana) Stop() bool {
	errs.PanicIfNotNil(errors.New("not supported"))
	return false
}
