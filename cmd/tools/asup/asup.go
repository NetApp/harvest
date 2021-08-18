package asup

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/v3/mem"
	"goharvest2/cmd/harvest/version"
	"goharvest2/cmd/poller/collector"
	client "goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/logging"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"os"
	"os/exec"
	"path"
	"time"
)

type asupMessage struct {
	Target   *targetInfo
	Nodes    *NodeInstanceInfo
	Volumes  *instanceInfo
	Svms     *instanceInfo
	Platform *platformInfo
	Harvest  *harvestInfo
}

type targetInfo struct {
	Version     string
	Model       string
	Serial      string
	Ping        float64
	ClusterUuid string
}

type NodeInstanceInfo struct {
	Count      int64
	DataPoints int64
	PollTime   int64
	ApiTime    int64
	ParseTime  int64
	PluginTime int64
	NodeUuid   []string
}

type instanceInfo struct {
	Count      int64
	DataPoints int64
	PollTime   int64
	ApiTime    int64
	ParseTime  int64
	PluginTime int64
}

type platformInfo struct {
	OS       string
	Arch     string
	MemoryKb uint64
	CPUs     uint8
}

type harvestInfo struct {
	UUID        string
	Version     string
	Release     string
	Commit      string
	BuildDate   string
	NumClusters uint8
}

func DoAsupMessage(config string, collectors []collector.Collector, status *matrix.Matrix, pollerName string) error {

	var (
		msg *asupMessage
		err error
	)

	connection := getdata(config, pollerName)

	if msg, err = buildAsupMessage(collectors, status, connection); err != nil {
		return fmt.Errorf("failed to build ASUP message poller:%s %w", pollerName, err)
	}

	if err = sendAsupMessage(msg, pollerName); err != nil {
		return fmt.Errorf("failed to send ASUP message poller:%s %w", pollerName, err)
	}

	return nil
}

// This function forks the autosupport binary
func sendAsupMessage(msg *asupMessage, pollerName string) error {
	err := sendAsupVia(msg, pollerName, "./bin/asup")
	if errors.Is(err, os.ErrNotExist) {
		err = sendAsupVia(msg, pollerName, "../harvest-private/harvest-asup/bin/asup")
	}
	if err != nil {
		return err
	}
	return nil
}

func sendAsupVia(msg *asupMessage, pollerName string, asupExecPath string) error {
	var (
		payloadPath string
		err         error
	)
	asupTimeOutLimit := 10 * time.Second
	workingDir := "asup"

	if payloadPath, err = getPayloadPath(workingDir, pollerName); err != nil {
		return err
	}

	// name of the file: {poller_name}_payload.json
	file, err := os.OpenFile(payloadPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("autosupport failed to open payloadPath:%s %w", payloadPath, err)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err = encoder.Encode(msg); err != nil {
		return fmt.Errorf("autosupport failed to encode payloadPath:%s %w", payloadPath, err)
	}

	// Invoke autosupport binary
	cont, cancel := context.WithTimeout(context.Background(), asupTimeOutLimit)
	defer cancel()
	logging.Get().Info().
		Str("payloadPath", payloadPath).
		Msg("Fork autosupport binary.")

	exitStatus := 0
	err = exec.CommandContext(cont, asupExecPath, "--payload", payloadPath, "--working-dir", workingDir).Run()
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
		logging.Get().Error().
			Err(err).
			Str("payloadPath", payloadPath).
			Msg("Unable to execute autosupport binary")
		return err
	}

	logging.Get().Info().
		Str("payloadPath", payloadPath).
		Int("exitStatus", exitStatus).
		Msg("Autosupport binary forked successfully.")

	return nil
}

func buildAsupMessage(collectors []collector.Collector, status *matrix.Matrix, connection *client.Client) (*asupMessage, error) {

	var (
		msg  *asupMessage
		arch string
		cpus uint8
	)

	msg = new(asupMessage)

	// add info about platform (where Harvest is running)
	msg.Platform = new(platformInfo)
	arch, cpus = getCPUInfo()
	msg.Platform.Arch = arch
	msg.Platform.CPUs = cpus
	msg.Platform.MemoryKb = getRamSize()
	msg.Platform.OS = getOSName()

	// info about ONTAP host and instances
	msg.Target, msg.Nodes, msg.Volumes, msg.Svms = getInstanceInfo(collectors, status, connection)

	// add harvest release info
	msg.Harvest = new(harvestInfo)
	// harvest uuid creation from sha1 of cluster uuid
	msg.Harvest.UUID = getSha1Uuid(msg.Target.ClusterUuid)
	msg.Harvest.Version = version.VERSION
	msg.Harvest.Release = version.Release
	msg.Harvest.Commit = version.Commit
	msg.Harvest.BuildDate = version.BuildDate
	msg.Harvest.NumClusters = getNumClusters(collectors)
	return msg, nil
}

func getInstanceInfo(collectors []collector.Collector, status *matrix.Matrix, connection *client.Client) (*targetInfo, *NodeInstanceInfo, *instanceInfo, *instanceInfo) {
	target := new(targetInfo)
	nodes := new(NodeInstanceInfo)
	vols := new(instanceInfo)
	svms := new(instanceInfo)

	// get ping value from poller metadata
	target.Ping, _ = status.LazyGetValueFloat64("ping", "host")

	// scan collectors
	for _, c := range collectors {
		if c.GetName() == "Zapi" {
			if c.GetObject() == "Node" {
				md := c.GetMetadata()

				nodes.getNodeInfo(md, connection)

				target.Version = c.GetHostVersion()
				target.Model = c.GetHostModel()
				target.Serial = c.GetHostUUID()
				target.getClusterInfo(connection)

			} else if c.GetObject() == "Volume" {
				md := c.GetMetadata()
				vols.Count, _ = md.LazyGetValueInt64("count", "instance")
				vols.DataPoints, _ = md.LazyGetValueInt64("count", "data")
				vols.PollTime, _ = md.LazyGetValueInt64("poll_time", "data")
				vols.ApiTime, _ = md.LazyGetValueInt64("api_time", "data")
				vols.ParseTime, _ = md.LazyGetValueInt64("parse_time", "data")
				vols.PluginTime, _ = md.LazyGetValueInt64("plugin_time", "data")

				//} else if c.GetObject() == "Svm" {
				//	md := c.GetMetadata()
				//	svms.Count, _ = md.LazyGetValueInt64("count", "instance")
				//	svms.DataPoints, _ = md.LazyGetValueInt64("count", "data")
				//	svms.PollTime, _ = md.LazyGetValueInt64("poll_time", "data")
				//	svms.ApiTime, _ = md.LazyGetValueInt64("api_time", "data")
				//	svms.ParseTime, _ = md.LazyGetValueInt64("parse_time", "data")
				//	svms.PluginTime, _ = md.LazyGetValueInt64("plugin_time", "data")
			}
		} else if c.GetName() == "ZapiPerf" {
			// TODO: SVM data population
			if c.GetObject() == "Svm" {
				md := c.GetMetadata()
				svms.Count, _ = md.LazyGetValueInt64("count", "instance")
				svms.DataPoints, _ = md.LazyGetValueInt64("count", "data")
				svms.PollTime, _ = md.LazyGetValueInt64("poll_time", "data")
				svms.ApiTime, _ = md.LazyGetValueInt64("api_time", "data")
				svms.ParseTime, _ = md.LazyGetValueInt64("parse_time", "data")
				svms.PluginTime, _ = md.LazyGetValueInt64("plugin_time", "data")
			}
		}
	}

	return target, nodes, vols, svms
}

func getCPUInfo() (string, uint8) {

	var (
		arch     string
		cpuCount int
		cpuinfo  []cpu.InfoStat
		hostinfo *host.InfoStat
		err      error
	)

	if cpuinfo, err = cpu.Info(); err == nil {
		if cpuinfo != nil {
			cpuCount = len(cpuinfo)
		}
	}

	if hostinfo, err = host.Info(); err == nil {
		if hostinfo != nil {
			arch = (*hostinfo).Platform
		}
	}

	return arch, uint8(cpuCount)
}

func getRamSize() uint64 {

	var (
		//output           []byte
		err error
		//line, sizeString string
		//fields           []string
		memory *mem.VirtualMemoryStat
		size   uint64
	)

	if memory, err = mem.VirtualMemory(); err == nil {
		fmt.Printf("%s", *memory)
		size = memory.Free / 1024
	}

	return size
}

func getOSName() string {
	info, err := host.Info()
	if err != nil {
		return ""
	}
	return info.OS
}

func getNumClusters(collectors []collector.Collector) uint8 {
	var count uint8

	for _, c := range collectors {
		if c.GetName() == "Zapi" || c.GetName() == "ZapiPerf" {
			count++
			break
		}
	}

	return count
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

func getClusterNodeInfo(client *client.Client) ([]string, error) {
	var (
		response *node.Node
		nodes    []*node.Node
		err      error
		uuids    []string
	)

	request := "cluster-node-get-iter"

	if response, err = client.InvokeRequestString(request); err != nil {
		return nil, fmt.Errorf("failure invoking zapi: %s %w", request, err)
	}

	if attrs := response.GetChildS("attributes-list"); attrs != nil {
		nodes = attrs.GetChildren()
	}

	for _, n := range nodes {
		uuid := n.GetChildContentS("node-uuid")
		uuids = append(uuids, uuid)
	}

	return uuids, nil
}

func (target *targetInfo) getClusterInfo(connection *client.Client) error {
	var (
		response *node.Node
		err      error
	)

	isClustered := target.Model == "cdot"
	// fetch system name and serial number
	request := "cluster-identity-get"
	if !isClustered {
		request = "system-get-info"
	}

	if response, err = connection.InvokeRequestString(request); err != nil {
		return fmt.Errorf("failure invoking zapi: %s %w", request, err)
	}

	if isClustered {
		if attrs := response.GetChildS("attributes"); attrs != nil {
			if info := attrs.GetChildS("cluster-identity-info"); info != nil {
				target.ClusterUuid = info.GetChildContentS("cluster-uuid")
			}
		}
	} else {
		if info := response.GetChildS("system-info"); info != nil {
			// There is no uuid for non cluster mode, using system-id.
			target.ClusterUuid = info.GetChildContentS("system-id")
		}
	}
	return nil
}

func (node *NodeInstanceInfo) getNodeInfo(md *matrix.Matrix, connection *client.Client) error {
	var uuids []string
	var err error
	if uuids, err = getClusterNodeInfo(connection); err != nil {
		fmt.Printf("issue while fetching node uuids")
	}
	node.Count, _ = md.LazyGetValueInt64("count", "instance")
	node.DataPoints, _ = md.LazyGetValueInt64("count", "data")
	node.PollTime, _ = md.LazyGetValueInt64("poll_time", "data")
	node.ApiTime, _ = md.LazyGetValueInt64("api_time", "data")
	node.ParseTime, _ = md.LazyGetValueInt64("parse_time", "data")
	node.PluginTime, _ = md.LazyGetValueInt64("plugin_time", "data")
	node.NodeUuid = uuids
	return nil
}

func getdata(config, pollername string) *client.Client {
	var (
		params     *node.Node
		err        error
		connection *client.Client
	)

	if params, err = conf.GetPoller(config, pollername); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if connection, err = client.New(params); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err = connection.Init(2); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return connection
}

func getSha1Uuid(clusterUuid string) string {
	shaHash := sha1.New()
	if _, err := shaHash.Write([]byte(clusterUuid)); err != nil {
		fmt.Printf("issue while invoking sha1 of uuid")
		return ""
	}
	return hex.EncodeToString(shaHash.Sum(nil))
}
