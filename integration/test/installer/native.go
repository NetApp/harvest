package installer

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/setup"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
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
	log.Println("Downloaded: " + n.path)
	Uninstall()
	log.Println("Installing " + tarFileName)
	unTarOutput := utils.Run("tar", "-xf", tarFileName, "--one-top-level=harvest", "--strip-components", "1", "-C", "/opt")
	log.Println(unTarOutput)
	utils.RemoveSafely(HarvestHome + "/" + harvestFile)
	utils.UseCertFile(HarvestHome)
	utils.Run("cp", setup.GetPerfFileWithQosCounters(setup.ZapiPerfDefaultFile, "defaultZapi.yaml"), HarvestHome+"/"+setup.ZapiPerfDefaultFile)
	utils.Run("cp", setup.GetPerfFileWithQosCounters(setup.RestPerfDefaultFile, "defaultRest.yaml"), HarvestHome+"/"+setup.RestPerfDefaultFile)
	err = utils.CopyFile(harvestFile, HarvestHome+"/"+harvestFile)
	if err != nil {
		panic(err)
	}
	harvestObj.Start()
	status := harvestObj.AllRunning()
	return status
}

func (n *Native) Upgrade() bool {
	utils.PanicIfNotNil(fmt.Errorf("not supported"))
	return false
}
