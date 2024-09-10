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

type Prometheus struct {
	image string
}

func (p *Prometheus) Init(image string) {
	p.image = image
}

func (p *Prometheus) Install() bool {
	p.image = "prom/prometheus:v2.33.0"
	slog.Info("Prometheus image : " + p.image)
	imageName := "prometheus"
	err := docker.StopContainers(imageName)
	utils.PanicIfNotNil(err)
	path, _ := os.Getwd()
	ipAddress := utils.GetOutboundIP()
	cmd := exec.Command("docker", "run", "-d", "-p", utils.PrometheusPort+":"+utils.PrometheusPort, //nolint:gosec
		"--add-host=localhost:"+ipAddress,
		"-v", path+"/../../container/prometheus/:/etc/prometheus/",
		p.image)
	cmd.Stdout = os.Stdout
	err = cmd.Start()
	utils.PanicIfNotNil(err)
	waitCount := 0
	maxWaitCount := 5
	for waitCount < maxWaitCount {
		waitCount++
		time.Sleep(60 * time.Second)
		if utils.IsURLReachable("http://localhost:" + utils.PrometheusPort) {
			return true
		}
	}
	slog.Info("Reached maximum timeout. Prometheus is failed to start", slog.Int("maxWaitCount", maxWaitCount))
	return false
}

func (p *Prometheus) Upgrade() bool {
	utils.PanicIfNotNil(errors.New("not supported"))
	return false
}

func (p *Prometheus) Stop() bool {
	utils.PanicIfNotNil(errors.New("not supported"))
	return false
}
