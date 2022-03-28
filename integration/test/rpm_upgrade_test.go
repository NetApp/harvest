package main

import (
	"github.com/Netapp/harvest-automation/test/installer"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
	"os"
	"testing"
)

func TestRHELUpgrade(t *testing.T) {
	utils.SetupLogging()
	var path = os.Getenv("BUILD_PATH")
	if len(path) == 0 {
		panic("BUILD_PATH variable is not set.")
	}
	installObject, err := installer.GetInstaller(installer.RHEL, path)
	utils.PanicIfNotNil(err)
	if installObject.Upgrade() {
		log.Println("Upgrade is successful..")
	} else {
		log.Println("Setup completed")
		panic("Upgrade is failed.")
	}
}
