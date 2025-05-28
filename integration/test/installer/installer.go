package installer

import (
	"bufio"
	"github.com/Netapp/harvest-automation/test/cmds"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"log/slog"
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
	Stop() bool
}

func GetPerfFileWithQosCounters(source string, target string) string {
	// Create a file for writing
	modifiedFilePath := cmds.GetHarvestRootDir() + "/" + target
	cmds.RemoveSafely(modifiedFilePath)
	writeFile, _ := os.Create(modifiedFilePath)
	writeBuffer := bufio.NewWriter(writeFile)
	file, err := os.Open(cmds.GetHarvestRootDir() + "/" + source)
	if err != nil {
		slog.Error("", slogx.Err(err))
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
		slog.Error("", slogx.Err(err))
	}
	_ = writeBuffer.Flush()
	return modifiedFilePath
}
