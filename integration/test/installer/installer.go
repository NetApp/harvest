package installer

import (
	"bufio"
	"github.com/Netapp/harvest-automation/test/utils"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

const (
	ZapiPerfDefaultFile = "conf/zapiperf/default.yaml"
	RestPerfDefaultFile = "conf/restperf/default.yaml"
)

type Installer interface {
	Install() bool
	Upgrade() bool
}

func GetPerfFileWithQosCounters(source string, target string) string {
	// Create a file for writing
	modifiedFilePath := utils.GetHarvestRootDir() + "/" + target
	utils.RemoveSafely(modifiedFilePath)
	writeFile, _ := os.Create(modifiedFilePath)
	writeBuffer := bufio.NewWriter(writeFile)
	file, err := os.Open(utils.GetHarvestRootDir() + "/" + source)
	if err != nil {
		log.Error().Err(err)
	}
	defer func(file *os.File) { _ = file.Close() }(file)

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		lineString := scanner.Text()
		if strings.Contains(lineString, "#  Workload") {
			lineString = strings.ReplaceAll(lineString, "#  Workload", "  Workload")
		}
		_, _ = writeBuffer.WriteString(lineString + "\n")
	}
	if err := scanner.Err(); err != nil {
		log.Error().Err(err)
	}
	_ = writeBuffer.Flush()
	return modifiedFilePath
}
