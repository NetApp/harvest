package collector

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/harvest/version"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"os"
	"os/exec"
	"path"
	"time"
)

type Payload struct {
	Target     *TargetInfo
	Harvest    *harvestInfo
	Platform   *platformInfo
	Nodes      *InstanceInfo
	Volumes    *InstanceInfo
	Collectors *[]AsupCollector
	path       string
}

type TargetInfo struct {
	Version     string
	Model       string
	Serial      string
	Ping        float64
	ClusterUuid string
}

type Id struct {
	SerialNumber string `json:"serial-number"`
	SystemId     string `json:"system-id"`
}

type InstanceInfo struct {
	Count      int64
	DataPoints int64
	PollTime   int64
	ApiTime    int64
	ParseTime  int64
	PluginTime int64
	Ids        []Id `json:"Ids,omitempty"`
}

type platformInfo struct {
	OS     string
	Arch   string
	Memory struct {
		TotalKb     uint64
		AvailableKb uint64
		UsedKb      uint64
	}
	CPUs uint8
}

type harvestInfo struct {
	HostHash    string
	UUID        string
	Version     string
	Release     string
	Commit      string
	BuildDate   string
	NumClusters uint8
}

type Counters struct {
	Count int
	List  []string
}

type Schedule struct {
	Name     string
	Schedule string
}

type AsupCollector struct {
	Name          string
	Query         string
	BatchSize     string `json:"BatchSize,omitempty"`
	ClientTimeout string
	Schedules     []Schedule
	Exporters     []string
	Counters      Counters
}

const workingDir = "asup"

func (p *Payload) AddCollectorAsup(a AsupCollector) {
	if p.Collectors == nil {
		p.Collectors = &[]AsupCollector{}
	}
	*p.Collectors = append(*p.Collectors, a)
}

func SendAutosupport(collectors []Collector, status *matrix.Matrix, pollerName string) error {

	var (
		msg *Payload
		err error
	)

	if msg, err = BuildAndWriteAutoSupport(collectors, status, pollerName); err != nil {
		return fmt.Errorf("failed to build ASUP message poller:%s %w", pollerName, err)
	}

	if err = sendAsupMessage(msg); err != nil {
		return fmt.Errorf("failed to send ASUP message poller:%s %w", pollerName, err)
	}

	return nil
}

// This function forks the autosupport binary
func sendAsupMessage(msg *Payload) error {
	err := sendAsupVia(msg, "./autosupport/asup")
	if errors.Is(err, os.ErrNotExist) {
		err = sendAsupVia(msg, "../harvest-private/harvest-asup/bin/asup")
	}
	if err != nil {
		return err
	}
	return nil
}

func sendAsupVia(msg *Payload, asupExecPath string) error {
	asupTimeOutLimit := 10 * time.Second

	// Invoke autosupport binary
	cont, cancel := context.WithTimeout(context.Background(), asupTimeOutLimit)
	defer cancel()
	logging.Get().Info().
		Str("payloadPath", msg.path).
		Msg("Fork autosupport binary.")

	exitStatus := 0
	err := exec.CommandContext(cont, asupExecPath, "--payload", msg.path, "--working-dir", workingDir).Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitStatus = exitError.ExitCode()
		}
	}

	// make sure to timeout after x minutes, kill that process
	if errors.Is(cont.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("autosupport call to %s timed out:%w", asupExecPath, err)
	}

	if err != nil {
		return err
	}

	logging.Get().Info().
		Str("payloadPath", msg.path).
		Int("exitStatus", exitStatus).
		Msg("Autosupport binary forked successfully.")

	return nil
}

func BuildAndWriteAutoSupport(collectors []Collector, status *matrix.Matrix, pollerName string) (*Payload, error) {

	var (
		msg  *Payload
		arch string
		cpus uint8
	)

	// add info about platform (where Harvest is running)
	arch, cpus = getCPUInfo()
	msg = &Payload{
		Platform: &platformInfo{
			Arch: arch,
			CPUs: cpus,
			OS:   getOSName(),
		},
		Target: &TargetInfo{
			Ping: status.LazyValueFloat64("ping", "host"),
		},
	}
	attachMemory(msg)

	// give each collector the opportunity to attach autosupport information
	for _, c := range collectors {
		c.CollectAutoSupport(msg)
	}

	hostname, _ := os.Hostname()
	// add harvest release info
	msg.Harvest = &harvestInfo{
		// harvest uuid creation from sha1 of cluster uuid
		UUID:        sha1Sum(msg.Target.ClusterUuid),
		Version:     version.VERSION,
		Release:     version.Release,
		Commit:      version.Commit,
		BuildDate:   version.BuildDate,
		HostHash:    sha1Sum(hostname),
		NumClusters: 1,
	}
	payloadPath, err := writeAutoSupport(msg, pollerName)
	if err != nil {
		return nil, err
	}
	msg.path = payloadPath
	return msg, nil
}

func writeAutoSupport(msg *Payload, pollerName string) (string, error) {
	var (
		payloadPath string
		err         error
	)

	if payloadPath, err = getPayloadPath(workingDir, pollerName); err != nil {
		return "", err
	}

	// name of the file: {poller_name}_payload.json
	file, err := os.OpenFile(payloadPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("autosupport failed to open payloadPath:%s %w", payloadPath, err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err = encoder.Encode(msg); err != nil {
		return "", fmt.Errorf("autosupport failed to encode payloadPath:%s %w", payloadPath, err)
	}
	return payloadPath, nil
}

func attachMemory(msg *Payload) {
	virtualMemory, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	msg.Platform.Memory.TotalKb = virtualMemory.Total / 1024
	msg.Platform.Memory.AvailableKb = virtualMemory.Available / 1024
	msg.Platform.Memory.UsedKb = virtualMemory.Used / 1024
}

func getCPUInfo() (string, uint8) {

	var (
		arch     string
		cpuCount int
		cpuInfo  []cpu.InfoStat
		hostInfo *host.InfoStat
		err      error
	)

	if cpuInfo, err = cpu.Info(); err == nil {
		if cpuInfo != nil {
			cpuCount = len(cpuInfo)
		}
	}

	if hostInfo, err = host.Info(); err == nil {
		if hostInfo != nil {
			arch = (*hostInfo).Platform
		}
	}

	return arch, uint8(cpuCount)
}

func getOSName() string {
	info, err := host.Info()
	if err != nil {
		return ""
	}
	return info.OS
}

// Gives asup payload json file path based on input.
// Ex. asupDir = asup, pollerName = poller-1
// o/p would be asup/payload/poller-1_payload.json
func getPayloadPath(asupDir string, pollerName string) (string, error) {
	payloadDir := path.Join(asupDir, "payload")

	// Create the asup payload directory if needed
	if _, err := os.Stat(payloadDir); os.IsNotExist(err) {
		if err = os.MkdirAll(payloadDir, 0777); err != nil {
			return "", fmt.Errorf("could not create asup payload directory %s: %w", payloadDir, err)
		}
	}

	return path.Join(payloadDir, fmt.Sprintf("%s_%s", pollerName, "payload.json")), nil
}

func sha1Sum(s string) string {
	hash := sha1.New()
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}
