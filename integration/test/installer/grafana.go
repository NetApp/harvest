package installer

import (
	"errors"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/utils"
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
	_ = docker.StopContainers(imageName)
	cmd := exec.Command("docker", "run", "-d", "-e", "GF_LOG_LEVEL=debug", "-p", utils.GrafanaPort+":"+utils.GrafanaPort, g.image) //nolint:gosec
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	utils.PanicIfNotNil(err)
	waitCount := 0
	maxWaitCount := 15
	for waitCount < maxWaitCount {
		waitCount++
		time.Sleep(1 * time.Minute)
		if utils.IsURLReachable("http://localhost:" + utils.GrafanaPort) {
			return true
		}
	}
	slog.Info("Reached maximum timeout. Grafana is failed to start", slog.Int("maxWaitCount", maxWaitCount))
	return false
}

func (g *Grafana) Upgrade() bool {
	utils.PanicIfNotNil(errors.New("not supported"))
	return false
}

func (g *Grafana) Stop() bool {
	utils.PanicIfNotNil(errors.New("not supported"))
	return false
}
