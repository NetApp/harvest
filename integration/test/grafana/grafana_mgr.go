package grafana

import (
	"errors"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/docker"
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
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
	slog.Info("Verify Grafana and Prometheus are configured")
	var re = regexp.MustCompile(`404|not-found|error`)
	if !cmds.IsURLReachable(cmds.GetGrafanaHTTPURL()) {
		panic(errors.New("grafana is not reachable"))
	}
	if !cmds.IsURLReachable(cmds.GetPrometheusURL()) {
		panic(errors.New("prometheus is not reachable"))
	}
	slog.Info("Import dashboard from grafana/dashboards")
	containerIDs, err := docker.Containers("poller")
	if err != nil {
		panic(err)
	}
	grafanaURL := cmds.GetGrafanaURL()
	if docker.IsDockerBasedPoller() {
		grafanaURL = "grafana:3000"
	}
	importCmds := []string{"grafana", "import", "--addr", grafanaURL}
	if docker.IsDockerBasedPoller() {
		params := []string{"exec", containerIDs[0].ID, "bin/harvest"} //nolint:prealloc
		params = append(params, importCmds...)
		importOutput, err = cmds.Run("docker", params...)
	} else {
		slog.Info("It is non docker based harvest")
		importOutput, err = cmds.Exec(installer.HarvestHome, "bin/harvest", nil, importCmds...)
	}
	if err != nil {
		slog.Error("error", slogx.Err(err))
		panic(err)
	}
	status = !(re.MatchString(importOutput))
	slog.Info("Grafana import status", slog.Bool("status", status))
	return status, importOutput
}
