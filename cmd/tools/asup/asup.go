package asup

import (
	"context"
	"encoding/json"
	"fmt"
	"goharvest2/cmd/harvest/version"
	"goharvest2/cmd/poller/collector"
	client "goharvest2/pkg/api/ontapi/zapi"
	"goharvest2/pkg/conf"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/matrix"
	"goharvest2/pkg/tree/node"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

func DoAsupMessage(config string, collectors []collector.Collector, status *matrix.Matrix, pollerName string, harvestUUID string) error {

	var (
		msg *asupMessage
		err error
	)

	connection := getdata(config, pollerName)

	if msg, err = buildAsupMessage(collectors, status, connection, harvestUUID); err != nil {
		return errors.New(errors.ERR_CONFIG, "failed to build ASUP message")
	}

	if err = sendAsupMessage(msg, pollerName); err != nil {
		return errors.New(errors.ERR_CONFIG, "failed to send ASUP message")
	}

	return nil
}

// This function would be used to invoke harvest-asup(private repo)
func sendAsupMessage(msg *asupMessage, pollerName string) error {
	var (
		payloadPath string
		err         error
	)
	asupTimeOutLimit := 10 * time.Second
	workingDir := "asup"
	//asupExecPath := "./../../Harvest/harvest-private/harvest-asup/bin/asup"
	asupExecPath := "./bin/asup"

	if payloadPath, err = getPayloadPath(workingDir, pollerName); err != nil {
		return err
	}
	fmt.Printf("%s", payloadPath)

	// name of the file: {poller_name}_payload.json
	file, err := os.OpenFile(payloadPath, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		errors.New(errors.ERR_CONFIG, "asup json creation failed")
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	if err = encoder.Encode(msg); err != nil {
		errors.New(errors.ERR_CONFIG, "writing to payload failed")
	}

	// Invoke private-asup binary
	cont, cancel := context.WithTimeout(context.Background(), asupTimeOutLimit)
	defer cancel()
	out, err := exec.CommandContext(cont, asupExecPath, "--payload="+payloadPath, "--working-dir="+workingDir).Output()

	// make sure to timeout after x minutes, kill that process
	if cont.Err() == context.DeadlineExceeded {
		fmt.Print("asup call timed out ")
		return errors.New(errors.ERR_CONFIG, "asup call timed out")
	}

	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("sent ASUP message")
	fmt.Printf("%s", string(out))
	return nil
}

func buildAsupMessage(collectors []collector.Collector, status *matrix.Matrix, connection *client.Client, harvestUUID string) (*asupMessage, error) {

	var (
		msg  *asupMessage
		arch string
		cpus uint8
	)

	// @DEBUG all log messages are info or higher, only for development/debugging
	fmt.Print("building ASUP message")

	msg = new(asupMessage)

	// add harvest release info
	msg.Harvest = new(harvestInfo)
	msg.Harvest.UUID = harvestUUID
	msg.Harvest.Version = version.VERSION
	msg.Harvest.Release = version.Release
	msg.Harvest.Commit = version.Commit
	msg.Harvest.BuildDate = version.BuildDate
	msg.Harvest.NumClusters = getNumClusters(collectors)

	// add info about platform (where Harvest is running)
	msg.Platform = new(platformInfo)
	arch, cpus = getCPUInfo()
	msg.Platform.Arch = arch
	msg.Platform.CPUs = cpus
	msg.Platform.MemoryKb = getRamSize()
	msg.Platform.OS = getOSName()

	// info about ONTAP host and instances
	msg.Target, msg.Nodes, msg.Volumes, msg.Svms = getInstanceInfo(collectors, status, connection)

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
		arch, countString, line string
		fields                  []string
		count                   uint64
		output                  []byte
		err                     error
	)

	if output, err = exec.Command("lscpu").Output(); err == nil {
		for _, line = range strings.Split(string(output), "\n") {
			if fields = strings.Fields(line); len(fields) >= 2 {
				if fields[0] == "Architecture:" {
					arch = fields[1]
				} else if fields[0] == "CPU(s):" {
					countString = fields[1]
				}
			}
		}
	}

	if countString != "" {
		if count, err = strconv.ParseUint(countString, 10, 8); err != nil {
		}
	}

	return arch, uint8(count)
}

func getRamSize() uint64 {

	var (
		output           []byte
		err              error
		line, sizeString string
		fields           []string
		size             uint64
	)

	if output, err = exec.Command("free", "--kilo").Output(); err == nil {
		for _, line = range strings.Split(string(output), "\n") {
			if fields = strings.Fields(line); len(fields) >= 4 && fields[0] == "Mem:" {
				sizeString = fields[1]
				break
			}
		}
	}

	if sizeString != "" {
		if size, err = strconv.ParseUint(sizeString, 10, 64); err != nil {
			size = 0
		}
	}

	return size
}

func getOSName() string {

	var (
		output     []byte
		err        error
		name, line string
		fields     []string
	)

	if output, err = ioutil.ReadFile("/etc/os-release"); err == nil {
		for _, line = range strings.Split(string(output), "\n") {
			if fields = strings.SplitN(line, "=", 2); len(fields) == 2 {
				if fields[0] == "NAME" {
					name = fields[1]
				} else if fields[1] == "PRETTY_NAME" {
					name = fields[1]
					break
				}
			}
		}
	}
	return strings.Trim(name, `"`)
}

func getNumClusters(collectors []collector.Collector) uint8 {
	var count uint8

	for _, collector := range collectors {
		if collector.GetName() == "Zapi" || collector.GetName() == "ZapiPerf" {
			count++
			break
		}
	}

	return count
}

// Gives asup payload dir path based on input.
// Ex. asupDir = asup, then o/p would be asup/payload
func getPayloadDir(asupDir string) string {
	return fmt.Sprintf("%s/%s", asupDir, "payload")
}

// Gives asup payload json file path based on input.
// Ex. asupDir = asup, pollerName = poller-1
// o/p would be asup/payload/poller-1_payload.json
func getPayloadPath(asupDir string, pollerName string) (string, error) {
	payloadDir := getPayloadDir(asupDir)

	// Create the asup payload directory if needed
	if _, err := os.Stat(payloadDir); os.IsNotExist(err) {
		if err = os.MkdirAll(payloadDir, 0777); err != nil {
			errors.New(errors.ERR_CONFIG, "Could not create asup payload directory.")
			return "", err
		}
	}

	return fmt.Sprintf("%s/%s_%s", payloadDir, pollerName, "payload.json"), nil
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
		fmt.Printf("H")
		errors.New(errors.ERR_CONFIG, "failure while invoking zapi call")
		return nil, err
	}

	if attrs := response.GetChildS("attributes-list"); attrs != nil {
		nodes = attrs.GetChildren()
	}

	for _, node := range nodes {
		uuid := node.GetChildContentS("node-uuid")
		uuids = append(uuids, uuid)
	}

	fmt.Printf("node uuid %s", uuids)

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
		errors.New(errors.ERR_CONFIG, "failure while invoking zapi call")
		return err
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
	fmt.Printf("cluster uuid %s", target.ClusterUuid)

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
