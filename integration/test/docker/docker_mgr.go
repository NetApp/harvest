package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
	"os/exec"
	"strings"
)

func IsDockerBasedPoller() bool {
	containers, err := Containers("poller")
	if err != nil {
		panic(err)
	}
	return len(containers) > 0
}

func StoreContainerLog(containerID string, logFile string) error {
	cli := fmt.Sprintf(`docker logs %s &> "%s"`, containerID, logFile)
	command := exec.Command("bash", "-c", cli)
	err := command.Run()
	if err != nil {
		return err
	}
	return nil
}

func ReStartContainers(commandSubString string) error {
	log.Printf("ReStartContainers start %s\n", commandSubString)
	containers, err := Containers(commandSubString)
	if err != nil {
		return err
	}

	var errs []error
	for _, container := range containers {
		log.Println("Restarting container", container.ID[:10], "...")
		command := exec.Command("docker", "container", "restart", container.ID) //nolint:gosec
		err = command.Run()
		if err != nil {
			errs = append(errs, err)
		}
	}
	log.Printf("ReStartContainers complete %s\n", commandSubString)
	return errors.Join(errs...)
}

func StopContainers(commandSubString string) error {
	log.Printf("StopContainers start  %s  \n", commandSubString)
	containers, err := Containers(commandSubString)
	if err != nil {
		return err
	}

	var errs []error
	for _, container := range containers {
		log.Println("Stopping container", container.ID[:10], "...")
		command := exec.Command("docker", "container", "stop", container.ID) //nolint:gosec
		err = command.Run()
		if err != nil {
			errs = append(errs, err)
		}
	}
	log.Printf("StopContainers complete  %s  \n", commandSubString)
	return errors.Join(errs...)
}

func CopyFile(containerID string, src string, dest string) {
	_, _ = utils.Run("docker", "cp", src, containerID+":"+dest)
}

func Containers(cmdPattern string) ([]Container, error) {
	cmd := exec.Command("curl", "--silent", "--unix-socket", "/var/run/docker.sock", "http://localhost/containers/json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("docker ps -a err: %w", err)
	}
	var all []Container
	err = json.Unmarshal(output, &all)
	if err != nil {
		return nil, err
	}

	matches := make([]Container, 0)
	for _, container := range all {
		if strings.Contains(container.Command, cmdPattern) {
			matches = append(matches, container)
		}
	}
	return matches, nil
}

type Container struct {
	ID      string   `json:"Id"`
	Names   []string `json:"Names"`
	Image   string   `json:"Image"`
	ImageID string   `json:"ImageID"`
	Command string   `json:"Command"`
	Created int      `json:"Created"`
	Ports   []struct {
		IP          string `json:"IP"`
		PrivatePort int    `json:"PrivatePort"`
		PublicPort  int    `json:"PublicPort"`
		Type        string `json:"Type"`
	} `json:"Ports"`
	State  string `json:"State"`
	Status string `json:"Status"`
	Mounts []struct {
		Type        string `json:"Type"`
		Name        string `json:"Name,omitempty"`
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		Driver      string `json:"Driver,omitempty"`
		Mode        string `json:"Mode"`
		RW          bool   `json:"RW"`
		Propagation string `json:"Propagation"`
	} `json:"Mounts"`
}

func (c Container) Name() string {
	return strings.Join(c.Names, " ")
}
