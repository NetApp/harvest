//+build install_rpm

package main

import (
	"github.com/Netapp/harvest-automation/test/installer"
	"log"
	"os"
	"testing"
)

func TestRHELInstall(t *testing.T) {
	var path = os.Getenv("BUILD_PATH")
	if len(path) == 0 {
		panic("BUILD_PATH variable is not set.")
	}
	installObject, error := installer.GetInstaller(installer.RHEL, path)
	if error != nil {
		log.Println("Unable to initialize installer object")
		panic(error)
	}
	if installObject.Install() {
		log.Println("Installation is successful..")
	} else {
		log.Println("Setup completed")
		panic("installation is failed.")
	}
	harvestObj := new(installer.Harvest)
	if harvestObj.AllRunning() {
		log.Println("All pollers are running")
	} else {
		t.Errorf("One or more pollers are not running.")
	}

}
