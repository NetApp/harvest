package setup

import (
	"bufio"
	"github.com/Netapp/harvest-automation/test/utils"
	log "github.com/cihub/seelog"
	"os"
	"runtime"
	"strings"
)

const ZapiPerfDefaultFile = "conf/zapiperf/default.yaml"

const IsMac = runtime.GOOS == "darwin"

func GetZapiPerfFileWithQosCounters() string {
	// Create a file for writing
	modifiedFilePath := utils.GetHarvestRootDir() + "/default.yaml"
	utils.RemoveSafely(modifiedFilePath)
	writeFile, _ := os.Create(modifiedFilePath)
	writeBuffer := bufio.NewWriter(writeFile)
	file, err := os.Open(utils.GetHarvestRootDir() + "/" + ZapiPerfDefaultFile)
	if err != nil {
		log.Error(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		lineString := scanner.Text()
		if strings.Contains(lineString, "#  Workload") {
			lineString = strings.ReplaceAll(lineString, "#  Workload", "  Workload")
		}
		writeBuffer.WriteString(lineString + "\n")
	}
	if err := scanner.Err(); err != nil {
		log.Error(err)
	}
	writeBuffer.Flush()
	return modifiedFilePath
}
