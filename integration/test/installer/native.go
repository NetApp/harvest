package installer

import (
	"errors"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/errs"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
)

type Native struct {
	path string
}

func (n *Native) Init(path string) {
	n.path = path
}

func (n *Native) Install() bool {
	harvestFile := "harvest.yml"
	limitedStatPerfFile := "/home/harvestfiles/statperf/limited.yaml"
	harvestObj := new(Harvest)
	tarFileName := "harvest.tar.gz"
	cmds.RemoveSafely(tarFileName)
	err := cmds.DownloadFile(tarFileName, n.path)
	if err != nil {
		panic(err)
	}
	slog.Info("Downloaded: " + n.path)
	Uninstall()
	slog.Info("Installing " + tarFileName)
	unTarOutput, err := cmds.Run("tar", "-xf", tarFileName, "--one-top-level=harvest", "--strip-components", "1", "-C", "/opt")
	if err != nil {
		slog.Error("", slogx.Err(err))
		panic(err)
	}
	slog.Info("Untar output: " + unTarOutput)
	cmds.RemoveSafely(HarvestHome + "/" + harvestFile)
	cmds.UseCertFile(HarvestHome)
	_, err1 := cmds.Run("cp", GetPerfFileWithQosCounters(ZapiPerfDefaultFile, "defaultZapi.yaml"), HarvestHome+"/"+ZapiPerfDefaultFile)
	_, err2 := cmds.Run("cp", GetPerfFileWithQosCounters(RestPerfDefaultFile, "defaultRest.yaml"), HarvestHome+"/"+RestPerfDefaultFile)
	_, err3 := cmds.Run("cp", harvestFile, HarvestHome+"/"+harvestFile)
	_, err4 := cmds.Run("cp", limitedStatPerfFile, HarvestHome+"/"+"conf/statperf/")
	err = errors.Join(err1, err2, err3, err4)
	if err != nil {
		panic(err)
	}
	harvestObj.Start()
	status := harvestObj.AllRunning()
	asupExecPath := HarvestHome + "/autosupport/asup"
	isValidAsup := harvestObj.IsValidAsup(asupExecPath)
	return status && isValidAsup
}

func (n *Native) Upgrade() bool {
	errs.PanicIfNotNil(errors.New("not supported"))
	return false
}

func (n *Native) Stop() bool {
	if cmds.FileExists(HarvestHome) {
		harvestObj := new(Harvest)
		if cmds.FileExists(HarvestHome + "/bin/harvest") {
			if harvestObj.AllRunning() {
				harvestObj.Stop()
			}
		}
	}
	return true
}
