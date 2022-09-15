package grafana

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
	"regexp"
)

type GrafanaMgr struct {
}

func (g *GrafanaMgr) Import(jsonDir string) (bool, string) {
	var (
		importOutput string
		status       bool
	)
	log.Println("Verify Grafana and Prometheus are configured")
	var re = regexp.MustCompile(`404|not-found|error`)
	if !utils.IsURLReachable(utils.GetGrafanaHTTPURL()) {
		panic(fmt.Errorf("grafana is not reachable"))
	}
	if !utils.IsURLReachable(utils.GetPrometheusURL()) {
		panic(fmt.Errorf("prometheus is not reachable"))
	}
	log.Println("Import dashboard from grafana/dashboards")
	containerIDs := docker.GetContainerID("poller")
	directoryOption := ""
	if len(jsonDir) > 0 {
		directoryOption = "--directory"
	}
	if !docker.IsDockerBasedPoller() {
		//assuming non docker based harvest grafana
		log.Println("It is non docker based harvest")
		importOutput = utils.Exec(installer.HarvestHome, "bin/grafana", nil, "import", "--addr", utils.GetGrafanaURL(), directoryOption, jsonDir)
	} else {
		params := []string{"exec", containerIDs[0], "bin/grafana", "import", "--addr", "grafana:3000", directoryOption, jsonDir}
		importOutput = utils.Run("docker", params...)
	}
	if re.MatchString(importOutput) {
		status = false
	} else {
		status = true
	}
	log.Println(fmt.Sprintf("Grafana import status : %t", status))
	return status, importOutput
}
