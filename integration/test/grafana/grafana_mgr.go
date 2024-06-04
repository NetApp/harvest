package grafana

import (
	"errors"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
	"regexp"
)

type Mgr struct {
}

func (g *Mgr) Import() (bool, string) {
	var (
		importOutput string
		status       bool
		err          error
	)
	log.Println("Verify Grafana and Prometheus are configured")
	var re = regexp.MustCompile(`404|not-found|error`)
	if !utils.IsURLReachable(utils.GetGrafanaHTTPURL()) {
		panic(errors.New("grafana is not reachable"))
	}
	if !utils.IsURLReachable(utils.GetPrometheusURL()) {
		panic(errors.New("prometheus is not reachable"))
	}
	log.Println("Import dashboard from grafana/dashboards")
	containerIDs, err := docker.Containers("poller")
	if err != nil {
		panic(err)
	}
	grafanaURL := utils.GetGrafanaURL()
	if docker.IsDockerBasedPoller() {
		grafanaURL = "grafana:3000"
	}
	importCmds := []string{"grafana", "import", "--overwrite", "--addr", grafanaURL}
	if docker.IsDockerBasedPoller() {
		params := []string{"exec", containerIDs[0].ID, "bin/harvest"}
		params = append(params, importCmds...)
		importOutput, err = utils.Run("docker", params...)
	} else {
		log.Println("It is non docker based harvest")
		importOutput, err = utils.Exec(installer.HarvestHome, "bin/harvest", nil, importCmds...)
	}
	if err != nil {
		log.Printf("error %s", err)
		panic(err)
	}
	if re.MatchString(importOutput) {
		status = false
	} else {
		status = true
	}
	log.Printf("Grafana import status : %t", status)
	return status, importOutput
}
