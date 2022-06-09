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
	log.Println("Prometheus image : " + p.image)
	imageName := "prometheus"
	docker.StopContainers(imageName)
	docker.RemoveImage(imageName)
	docker.PullImage(p.image)
	path, _ := os.Getwd()
	ipAddress := utils.GetOutboundIP()
	cmd := exec.Command("docker", "run", "-d", "-p", utils.PrometheusPort+":"+utils.PrometheusPort,
		"--add-host=localhost:"+ipAddress,
		"-v", path+"/../../docker/prometheus/:/etc/prometheus/",
		"prom/prometheus")
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	utils.PanicIfNotNil(err)
	waitCount := 0
	for waitCount < 5 {
		waitCount++
		time.Sleep(20 * time.Second)
		if utils.IsURLReachable("http://localhost:" + utils.PrometheusPort) {
			return true
		}
	}
	log.Println("Reached maximum timeout. Prometheus is failed to start after 1 min")
	return false
}

func (p *Prometheus) Upgrade() bool {
	utils.PanicIfNotNil(fmt.Errorf("not supported"))
	return false
}
