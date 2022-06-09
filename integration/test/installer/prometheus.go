package installer

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
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
	p.image = "prom/prometheus:v2.24.0"
	log.Println("Prometheus image : " + p.image)
	imageName := "prometheus"
	docker.StopContainers(imageName)
	docker.PullImage(p.image)
	path, _ := os.Getwd()
	ipAddress := utils.GetOutboundIP()
	cmd := exec.Command("docker", "run", "-d", "-p", utils.PrometheusPort+":"+utils.PrometheusPort,
		"--add-host=localhost:"+ipAddress, "-v",
		path+"/prometheus.yml:/etc/prometheus/prometheus.yml", p.image)
	cmd.Stdout = os.Stdout
	err := cmd.Start()
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
	log.Printf("Reached maximum timeout. Prometheus is failed to start after %d min\n", maxWaitCount)
	return false
}

func (p *Prometheus) Upgrade() bool {
	utils.PanicIfNotNil(fmt.Errorf("not supported"))
	return false
}
