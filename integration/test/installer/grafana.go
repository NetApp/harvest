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

type Grafana struct {
	image string
}

func (g *Grafana) Init(image string) {
	g.image = image
}

func (g *Grafana) Install() bool {
	g.image = "grafana/grafana:8.1.8"
	log.Println("Grafana image : " + g.image)
	imageName := "grafana"
	docker.StopContainers(imageName)
	//docker.PullImage(g.image)
	cmd := exec.Command("docker", "run", "-d", "-p", utils.GrafanaPort+":"+utils.GrafanaPort, g.image)
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
	log.Printf("Reached maximum timeout. Grafana is failed to start after %d min\n", maxWaitCount)
	return false
}

func (g *Grafana) Upgrade() bool {
	utils.PanicIfNotNil(fmt.Errorf("not supported"))
	return false
}
