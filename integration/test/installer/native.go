package installer

import (
	"errors"
	"github.com/Netapp/harvest-automation/test/utils"
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
	harvestObj := new(Harvest)
	tarFileName := "harvest.tar.gz"
	utils.RemoveSafely(tarFileName)
	err := utils.DownloadFile(tarFileName, n.path)
	if err != nil {
		panic(err)
	}
	slog.Info("Downloaded: " + n.path)
	Uninstall()
	slog.Info("Installing " + tarFileName)
	unTarOutput, err := utils.Run("tar", "-xf", tarFileName, "--one-top-level=harvest", "--strip-components", "1", "-C", "/opt")
	if err != nil {
		slog.Error("", slogx.Err(err))
		panic(err)
	}
	slog.Info("Untar output: " + unTarOutput)
	utils.RemoveSafely(HarvestHome + "/" + harvestFile)
	utils.UseCertFile(HarvestHome)
	_, err1 := utils.Run("cp", GetPerfFileWithQosCounters(ZapiPerfDefaultFile, "defaultZapi.yaml"), HarvestHome+"/"+ZapiPerfDefaultFile)
	_, err2 := utils.Run("cp", GetPerfFileWithQosCounters(RestPerfDefaultFile, "defaultRest.yaml"), HarvestHome+"/"+RestPerfDefaultFile)
	_, err3 := utils.Run("cp", harvestFile, HarvestHome+"/"+harvestFile)
	err = errors.Join(err1, err2, err3)
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
	utils.PanicIfNotNil(errors.New("not supported"))
	return false
}

func (n *Native) Stop() bool {
	if utils.FileExists(HarvestHome) {
		harvestObj := new(Harvest)
		if utils.FileExists(HarvestHome + "/bin/harvest") {
			if harvestObj.AllRunning() {
				harvestObj.Stop()
			}
		}
	}
	return true
}
