package main

import (
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/Netapp/harvest-automation/test/errs"
	"github.com/Netapp/harvest-automation/test/installer"
	"log"
	"os"
	"testing"
)

func TestRHELUpgrade(t *testing.T) {
	cmds.SkipIfMissing(t, cmds.UpgradeRPM)
	var path = os.Getenv("BUILD_PATH")
	if path == "" {
		panic("BUILD_PATH variable is not set.")
	}
	installObject, err := installer.GetInstaller(installer.RHEL, path)
	errs.PanicIfNotNil(err)
	if installObject.Upgrade() {
		log.Println("Upgrade is successful..")
	} else {
		log.Println("Setup completed")
		panic("Upgrade is failed.")
	}
}
