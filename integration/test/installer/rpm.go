package installer

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/setup"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
	"strings"
)

type RPM struct {
	path string
}

func (r *RPM) Init(path string) {
	r.path = path
}

func (r *RPM) Install() bool {
	harvestFile := "harvest.yml"
	rpmFileName := "harvest.rpm"
	utils.RemoveSafely(rpmFileName)
	err := utils.DownloadFile(rpmFileName, r.path)
	if err != nil {
		panic(err)
	}
	log.Println("Downloaded: " + r.path)
	Uninstall()
	harvestObj := new(Harvest)
	log.Println("Installing " + rpmFileName)
	installOutput, err := utils.Run("yum", "install", "-y", rpmFileName)
	if err != nil {
		log.Printf("error %s", err)
		panic(err)
	}
	log.Println(installOutput)
	log.Println("Stopping harvest")
	harvestObj.Stop()
	copyErr := utils.CopyFile(harvestFile, HarvestHome+"/harvest.yml")
	if copyErr != nil {
		return false
	} //use file directly from the repo
	harvestObj.Start()
	status := harvestObj.AllRunning()
	asupExecPath := HarvestHome + "/autosupport/asup"
	isValidAsup := harvestObj.IsValidAsup(asupExecPath)
	return status && isValidAsup
	return status
}

func (r *RPM) Upgrade() bool {
	rpmFileName := "harvest.rpm"
	utils.RemoveSafely(rpmFileName)
	harvestObj := new(Harvest)
	if !harvestObj.AllRunning() {
		utils.PanicIfNotNil(fmt.Errorf("pollers are not in a running state before upgrade"))
	}
	versionCmd := []string{"-qa", "harvest"}
	out, err := utils.Run("rpm", versionCmd...)
	if err != nil {
		log.Printf("error %s", err)
		panic(err)
	}
	previousVersion := strings.TrimSpace(out)
	err = utils.DownloadFile(rpmFileName, r.path)
	utils.PanicIfNotNil(err)
	log.Println("Downloaded: " + r.path)
	log.Println("Updating " + rpmFileName)
	installOutput, _ := utils.Run("yum", "upgrade", "-y", rpmFileName)
	log.Println(installOutput)
	out, _ = utils.Run("rpm", versionCmd...)
	installedVersion := strings.TrimSpace(out)
	if previousVersion == installedVersion {
		utils.PanicIfNotNil(fmt.Errorf("upgrade is failed"))
	}
	utils.Run("cp", setup.GetPerfFileWithQosCounters(setup.ZapiPerfDefaultFile, "defaultZapi.yaml"), HarvestHome+"/"+setup.ZapiPerfDefaultFile)
	utils.Run("cp", setup.GetPerfFileWithQosCounters(setup.RestPerfDefaultFile, "defaultRest.yaml"), HarvestHome+"/"+setup.RestPerfDefaultFile)
	harvestObj.Stop()
	harvestObj.Start()
	status := harvestObj.AllRunning()
	asupExecPath := HarvestHome + "/autosupport/asup"
	isValidAsup := harvestObj.IsValidAsup(asupExecPath)
	return status && isValidAsup
}
