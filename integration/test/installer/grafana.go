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

func (d *Grafana) Init(image string) {
	d.image = image
}

func (d *Grafana) Install() bool {
	log.Println("Grafana image : " + d.image)
	imageName := "grafana"
	docker.StopContainers(imageName)
	docker.RemoveImage(imageName)
	docker.PullImage(d.image)
	cmd := exec.Command("docker", "run", "-d", "-p", utils.GrafanaPort+":"+utils.GrafanaPort, "grafana/grafana:8.1.7")
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	utils.PanicIfNotNil(err)
	waitCount := 0
	for waitCount < 5 {
		waitCount++
		time.Sleep(20 * time.Second)
		if utils.IsUrlReachable("http://localhost:" + utils.GrafanaPort) {
			return true
		}
	}
	log.Println("Reached maximum timeout. Grafana is failed to start after 1 min")
	return false
}

func (g *Grafana) Upgrade() bool {
	utils.PanicIfNotNil(fmt.Errorf("not supported"))
	return false
}
