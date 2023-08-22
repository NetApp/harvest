package collector

import (
	"cmp"
	"context"
	"crypto/sha1" //nolint:gosec // used for sha1sum not for security
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/cmd/harvest/version"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"os"
	"os/exec"
	"path"
	"regexp"
	"slices"
	"time"
)

type Payload struct {
	Target     *TargetInfo
	Harvest    *harvestInfo
	Platform   *platformInfo
	Nodes      *InstanceInfo `json:"Nodes,omitempty"`
	Volumes    *InstanceInfo `json:"Volumes,omitempty"`
	Tenants    *InstanceInfo `json:"Tenants,omitempty"`
	Collectors *[]AsupCollector
	path       string
}

type TargetInfo struct {
	Version     string
	Model       string
	Serial      string
	Ping        float64
	ClusterUUID string
}

type ID struct {
	SerialNumber string `json:"serial-number"`
	SystemID     string `json:"system-id"`
}

type InstanceInfo struct {
	Count      int64
	DataPoints int64
	PollTime   int64
	APITime    int64
	ParseTime  int64
	PluginTime int64
	Ids        []ID `json:"Ids,omitempty"`
}

type Process struct {
	Pid      int32
	User     string
	Ppid     int32
	Ctime    int64
	RssBytes uint64
	Threads  int32
	Cmdline  string
}

type platformInfo struct {
	OS     string
	Arch   string
	Memory struct {
		TotalKb     uint64
		AvailableKb uint64
		UsedKb      uint64
	}
	CPUs         uint8
	NumProcesses uint64
	Processes    []Process
}

type harvestInfo struct {
	HostHash     string
	UUID         string
	Version      string
	Release      string
	Commit       string
	BuildDate    string
	NumClusters  uint8
	NumPollers   uint64
	NumExporters uint64
	NumPortRange uint64
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

const (
	workingDir = "asup"
	HundredMB  = 104_857_600
)

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
	err := exec.CommandContext(cont, asupExecPath, "--payload", msg.path, "--working-dir", workingDir).Run() //nolint:gosec
	if err != nil {
		var exitError *exec.ExitError
		if ok := errors.Is(err, exitError); ok {
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
		msg          *Payload
		arch         string
		cpus         uint8
		numPortRange uint64
	)

	// add info about the platform (where Harvest is running)
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

	// count the number of Prometheus exporters with portRange
	for _, e := range conf.Config.Exporters {
		if e.PortRange != nil {
			numPortRange++
		}
	}
	hostname, _ := os.Hostname()
	// add harvest release info
	msg.Harvest = &harvestInfo{
		// harvest uuid creation from sha1 of cluster uuid
		UUID:         sha1Sum(msg.Target.ClusterUUID),
		Version:      version.VERSION,
		Release:      version.Release,
		Commit:       version.Commit,
		BuildDate:    version.BuildDate,
		HostHash:     sha1Sum(hostname),
		NumClusters:  1,
		NumPollers:   uint64(len(conf.Config.Pollers)),
		NumExporters: uint64(len(conf.Config.Exporters)),
		NumPortRange: numPortRange,
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
	var perm os.FileMode = 0600
	file, err := os.OpenFile(payloadPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
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

var psAllowListRe = regexp.MustCompile(`harvest|poller|grafana|prometheus|influxdb|netapp`)

func attachMemory(msg *Payload) {
	virtualMemory, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	msg.Platform.Memory.TotalKb = virtualMemory.Total / 1024
	msg.Platform.Memory.AvailableKb = virtualMemory.Available / 1024
	msg.Platform.Memory.UsedKb = virtualMemory.Used / 1024

	// Include basic ps information for processes that pass the allow list
	// This is similar to ps -ww -eo user,pid,ppid,stime,nlwp,rss,cmd
	processes, err := process.Processes()
	if err != nil {
		return
	}

	msg.Platform.NumProcesses = uint64(len(processes))
	msg.Platform.Processes = make([]Process, 0)

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}
		pp := Process{
			Pid: p.Pid,
		}
		memInfo, err := p.MemoryInfo()
		if err == nil {
			pp.RssBytes = memInfo.RSS
		}

		// Ignore processes that:
		//   - don't pass the allowList, unless the process is using more that 100MB of memory
		//   - have no cmdline (these are not interesting)
		if !psAllowListRe.MatchString(name) {
			if pp.RssBytes < HundredMB {
				continue
			}
		}

		cmdline, err := p.Cmdline()
		if err != nil {
			continue
		}
		pp.Cmdline = cmdline
		if len(cmdline) == 0 {
			continue
		}

		username, err := p.Username()
		if err == nil {
			pp.User = username
		}
		ppid, err := p.Ppid()
		if err == nil {
			pp.Ppid = ppid
		}
		ctime, err := p.CreateTime()
		if err == nil {
			pp.Ctime = ctime
		}
		threads, err := p.NumThreads()
		if err == nil {
			pp.Threads = threads
		}
		msg.Platform.Processes = append(msg.Platform.Processes, pp)
	}

	// keep at most 100, ordered by size descending
	if len(msg.Platform.Processes) > 100 {
		slices.SortStableFunc(msg.Platform.Processes, func(a, b Process) int {
			return cmp.Compare(b.RssBytes, a.RssBytes)
		})
		msg.Platform.Processes = msg.Platform.Processes[:100]
	}

	// sort processes by pid
	slices.SortStableFunc(msg.Platform.Processes, func(a, b Process) int {
		return cmp.Compare(a.Pid, b.Pid)
	})
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

	// name of the file: {poller_name}_payload.json
	var perm os.FileMode = 0750
	err := checkAndDeleteIfPermissionsMismatch(workingDir, perm)
	if err != nil {
		logging.Get().Warn().Err(err).Send()
	}
	err = checkAndDeleteIfPermissionsMismatch(payloadDir, perm)
	if err != nil {
		logging.Get().Warn().Err(err).Send()
	}
	// Create the asup payload directory if needed
	if _, err := os.Stat(payloadDir); os.IsNotExist(err) {
		if err = os.MkdirAll(payloadDir, perm); err != nil {
			return "", fmt.Errorf("could not create asup payload directory %s: %w", payloadDir, err)
		}
	}

	return path.Join(payloadDir, fmt.Sprintf("%s_%s", pollerName, "payload.json")), nil
}

func sha1Sum(s string) string {
	hash := sha1.New() //nolint:gosec // using sha1 for a hash, not a security risk
	hash.Write([]byte(s))
	return hex.EncodeToString(hash.Sum(nil))
}

// checkAndDeleteIfPermissionsMismatch checks if the permissions of the file or directory at the given path
// match the required permissions. If they don't match, it deletes the file or directory.
func checkAndDeleteIfPermissionsMismatch(path string, requiredFileMode os.FileMode) error {
	// Get the file or directory information
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("error checking permissions: %w", err)
	}
	// Check if the current permissions match the required permissions
	currentPermissions := fileInfo.Mode().Perm()
	if currentPermissions != requiredFileMode {
		err = os.RemoveAll(path)
		if err != nil {
			return fmt.Errorf("error deleting file or directory: %w", err)
		}
	}
	return nil
}
