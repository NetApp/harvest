package installer

import (
	"fmt"
	"github.com/Netapp/harvest-automation/test/core"
	"github.com/Netapp/harvest-automation/test/utils"
	"log"
	"strings"
	"time"
)

const HarvestHome = "/opt/harvest"
const HarvestBin = "./bin/harvest"

type Harvest struct {
}

func (h *Harvest) Start() {
	status, err := utils.Exec(HarvestHome, HarvestBin, nil, "start")
	if err != nil {
		log.Printf("error %s", err)
		panic(err)
	}
	fmt.Println(status)
	time.Sleep(30 * time.Second)
	h.AllRunning()
}

func (h *Harvest) StartByHarvestUser() {
	status, err := utils.Exec(HarvestHome, "sudo", nil, "-u", "harvest", HarvestBin, "start")
	if err != nil {
		log.Printf("error %s", err)
		panic(err)
	}
	fmt.Println(status)
	time.Sleep(30 * time.Second)
	h.AllRunning()

}
func (h *Harvest) Stop() {
	status, err := utils.Exec(HarvestHome, HarvestBin, nil, "stop")
	if err != nil {
		log.Printf("error %s", err)
		panic(err)
	}
	fmt.Println(status)
}

func (h *Harvest) AllRunning() bool {
	pollerArray := h.GetPollerInfo()
	for _, poller := range pollerArray {
		if poller.Status != "running" {
			return false
		}
	}
	return true
}

func (h *Harvest) AllStopped() bool {
	pollerArray := h.GetPollerInfo()
	for _, poller := range pollerArray {
		if poller.Status != "not running" {
			return false
		}
	}
	return true
}

func (h *Harvest) GetPollerInfo() []core.Poller {
	log.Println("Getting all pollers details")
	harvestStatus, err := utils.Exec(HarvestHome, HarvestBin, nil, "status")
	if err != nil {
		log.Printf("error %s", err)
		panic(err)
	}
	fmt.Println(harvestStatus)
	rows := strings.Split(harvestStatus, "\n")
	var pollerArray []core.Poller
	for i := range rows {
		columns := strings.Split(rows[i], `|`)
		count := len(columns)
		if count != 5 || i == 0 { //ignore header and junk entries
			continue
		}
		dataCenter := columns[0]
		poller := columns[1]
		pid := columns[2]
		promPort := columns[3]
		pollerStatus := columns[4]
		pollerObject := core.Poller{
			DataCenter: dataCenter,
			Poller:     poller,
			Pid:        pid,
			PromPort:   promPort,
			Status:     strings.TrimSpace(pollerStatus),
		}
		pollerArray = append(pollerArray, pollerObject)
	}
	return pollerArray
}

func (h *Harvest) IsValidAsup(asupExecPath string) bool {
	out, err := utils.Exec("", asupExecPath, nil, "--version")
	if err != nil {
		fmt.Printf("error %s\n", err)
		return false
	}
	if !strings.Contains(out, "endpoint:stable") {
		fmt.Printf("asup endpoint is not stable %s\n", out)
		return false
	}
	fmt.Printf("asup validation successful %s\n", out)
	return true
}
