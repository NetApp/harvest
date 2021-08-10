//+build install_docker

package main

import (
	"github.com/Netapp/harvest-automation/test/installer"
	"log"
	"os"
	"testing"
)

func TestDockerInstall(t *testing.T) {
	var path = os.Getenv("BUILD_PATH")
	if len(path) == 0 {
		panic("BUILD_PATH variable is not set.")
	}
	installObject, err := installer.GetInstaller(installer.DOCKER, path)
	if err != nil {
		log.Println("Unable to initialize installer object")
		panic(err)
	}
	if installObject.Install() {
		log.Println("Installation is successful..")
	} else {
		log.Println("Setup completed")
		panic("installation is failed.")
	}

}
