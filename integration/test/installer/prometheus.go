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
	errs.PanicIfNotNil(err)
	path, _ := os.Getwd()
	ipAddress := cmds.GetOutboundIP()
	cmd := exec.Command("docker", "run", "-d", "-p", cmds.PrometheusPort+":"+cmds.PrometheusPort, //nolint:gosec
		"--add-host=localhost:"+ipAddress,
		"-v", path+"/../../container/prometheus/:/etc/prometheus/",
		p.image)
	cmd.Stdout = os.Stdout
	err = cmd.Start()
	errs.PanicIfNotNil(err)
	waitCount := 0
	maxWaitCount := 5
	for waitCount < maxWaitCount {
		waitCount++
		time.Sleep(60 * time.Second)
		if cmds.IsURLReachable("http://localhost:" + cmds.PrometheusPort) {
			return true
		}
	}
	slog.Info("Reached maximum timeout. Prometheus is failed to start", slog.Int("maxWaitCount", maxWaitCount))
	return false
}

func (p *Prometheus) Upgrade() bool {
	errs.PanicIfNotNil(errors.New("not supported"))
	return false
}

func (p *Prometheus) Stop() bool {
	errs.PanicIfNotNil(errors.New("not supported"))
	return false
}
