package installer

import (
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/utils"
	"goharvest2/pkg/conf"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Docker struct {
	path string
}

func (d *Docker) Init(path string) {
	d.path = path
}

func (d *Docker) Install() bool {
	log.Println("Docker build : " + d.path)
	pollerProcessName := "bin/poller"
	pollerNames, _ := conf.GetPollerNames(HarvestConfigFile)
	docker.StopContainers(pollerProcessName)
	docker.RemoveImage("harvest")
	var dockerImageName string
	if !strings.Contains(d.path, ".tar") {
		dockerImageName = d.path
		docker.PullImage(dockerImageName)
	} else {
		tarFileName := "harvest_latest.tar"
		utils.RemoveSafely(tarFileName)
		err := utils.DownloadFile(tarFileName, d.path)
		if err != nil {
			log.Println("Unable to download " + d.path)
			panic(err)
		}
		imageInfo := utils.Run("docker", "load", "-i", tarFileName)
		imageInfoArray := strings.Split(imageInfo, ":")
		if len(imageInfoArray) != 3 {
			panic("docker loaded image has invalid output format")
		}
		dockerImageName = strings.TrimSpace(imageInfoArray[1])
	}
	log.Println("Docker image name " + dockerImageName)
	path, _ := os.Getwd()
	for _, pollerName := range pollerNames {
		port, _ := conf.GetPrometheusExporterPorts(pollerName)
		portString := strconv.Itoa(port)
		ipAddress := utils.GetOutboundIP()
		cmd := exec.Command("docker", "run", "--rm", "-p", portString+":"+portString,
			"--add-host=localhost:"+ipAddress,
			"--volume", path+"/harvest.yml:/opt/harvest/harvest.yml",
			dockerImageName, "--poller", pollerName)
		cmd.Stdout = os.Stdout
		err := cmd.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Poller [" + pollerName + "] has been started successfully.")
	}
	waitCount := 0
	for waitCount < 5 {
		waitCount++
		time.Sleep(20 * time.Second)
		if docker.HasAllStarted(pollerProcessName, len(pollerNames)) {
			return true
		}
	}
	log.Println("Reached maximum timeout. One or more poller are failed to start after 1 min")
	return false
}
