package vscanpool

import (
	"cmp"
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/cmd/tools/rest"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"slices"
	"strings"
	"time"
)

var replaceStr = strings.NewReplacer("[", "", "]", "", "\"", "", "\n", "", " ", "")

type VscanPool struct {
	*plugin.AbstractPlugin
	client      *rest.Client
	testFile    string
	vscanServer *matrix.Matrix
}

type ServerData struct {
	ip         string
	state      string
	updateTime float64
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &VscanPool{AbstractPlugin: p}
}

func (v *VscanPool) Init(remote conf.Remote) error {

	var err error

	if err := v.InitAbc(); err != nil {
		return err
	}

	v.vscanServer = matrix.New(v.Parent+".VscanServer", "vscan_server", "vscan_server")
	exportOptions := node.NewS("export_options")
	instanceKeys := exportOptions.NewChildS("instance_keys", "")
	instanceKeys.NewChildS("", "vscan_server")
	instanceKeys.NewChildS("", "svm")
	v.vscanServer.SetExportOptions(exportOptions)
	_, err = v.vscanServer.NewMetricFloat64("disconnected", "disconnected")
	if err != nil {
		v.SLogger.Error("add metric", slogx.Err(err))
		return err
	}

	if v.Options.IsTest {
		return nil
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	if v.client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout, v.Auth); err != nil {
		v.SLogger.Error("connecting", slogx.Err(err))
		return err
	}

	if err := v.client.Init(5, remote); err != nil {
		return err
	}

	return nil
}

func (v *VscanPool) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	data := dataMap[v.Object]
	v.client.Metadata.Reset()
	v.vscanServer.Reset()

	// Purge and reset data
	v.vscanServer.PurgeInstances()
	v.vscanServer.Reset()

	// Set all global labels if they do not already exist
	v.vscanServer.SetGlobalLabels(data.GetGlobalLabels())

	svmPoolMap := v.fetchPool(data)
	vserverServerStateMap, err := v.getVScanServerInfo()
	if err != nil {
		v.SLogger.Error("Failed to collect vscan server info data", slogx.Err(err))
	} else {
		// update vscan disconnected instance labels
		v.updateVscanLabels(svmPoolMap, vserverServerStateMap)
	}

	return []*matrix.Matrix{v.vscanServer}, v.client.Metadata, nil
}

func (v *VscanPool) updateVscanLabels(svmPoolMap map[string][]string, vserverServerStateMap map[string]map[string]string) {
	for svm, pools := range svmPoolMap {
		notConectedServers := make([]string, 0)
		serverStateMap := vserverServerStateMap[svm]

		for _, pool := range pools {
			if state := serverStateMap[pool]; state == "disconnected" {
				notConectedServers = append(notConectedServers, pool)
			}
		}
		slices.Sort(notConectedServers)

		if len(notConectedServers) > 0 {
			instanceKey := svm
			vscanDisconnectedInstance, err := v.vscanServer.NewInstance(instanceKey)
			if err != nil {
				v.SLogger.Error("Failed to create new instance", slogx.Err(err), slog.String("instanceKey", instanceKey))
				continue
			}
			vscanDisconnectedInstance.SetLabel("vscan_server", strings.Join(notConectedServers, ","))
			if err = v.vscanServer.GetMetric("disconnected").SetValueFloat64(vscanDisconnectedInstance, 1); err != nil {
				v.SLogger.Error("Failed to set value", slogx.Err(err), slog.String("instanceKey", instanceKey))
			}
		}
	}
}

func (v *VscanPool) getVScanServerInfo() (map[string]map[string]string, error) {
	var (
		result []gjson.Result
		err    error
	)

	serverMap := make(map[string][]ServerData)
	vserverServerStateMap := make(map[string]map[string]string)
	fields := []string{"node.name", "svm.name", "ip", "update_time", "state", "interface.name"}
	query := "api/protocols/vscan/server-status"
	href := rest.NewHrefBuilder().
		APIPath(query).
		Fields(fields).
		MaxRecords(collectors.DefaultBatchSize).
		Build()

	if result, err = collectors.InvokeRestCallWithTestFile(v.client, href, v.testFile); err != nil {
		return nil, err
	}

	for _, vscanServer := range result {
		svmName := vscanServer.Get("svm.name").ClonedString()
		serverMap[svmName] = append(serverMap[svmName], ServerData{ip: vscanServer.Get("ip").ClonedString(), state: vscanServer.Get("state").ClonedString(), updateTime: collectors.HandleTimestamp(vscanServer.Get("update_time").ClonedString())})
	}

	for svm, serverData := range serverMap {
		serverStateMap := make(map[string]string)
		// Sort the slice by serverData in descending order
		slices.SortFunc(serverData, func(a, b ServerData) int {
			return cmp.Or(
				strings.Compare(a.ip, b.ip),
				cmp.Compare(b.updateTime, a.updateTime),
			)
		})

		for _, serverDetail := range serverData {
			if _, exist := serverStateMap[serverDetail.ip]; !exist {
				serverStateMap[serverDetail.ip] = serverDetail.state
			}
		}
		vserverServerStateMap[svm] = serverStateMap
	}
	return vserverServerStateMap, nil
}

func (v *VscanPool) fetchPool(data *matrix.Matrix) map[string][]string {
	svmPoolMap := make(map[string][]string)
	var vsName string
	var servers []string
	for _, instance := range data.GetInstances() {
		instance.SetExportable(false)
		if scannerPools := instance.GetLabel("scanner_pools"); scannerPools != "" {
			pools := gjson.Result{Type: gjson.JSON, Raw: scannerPools}
			for _, pool := range pools.Array() {
				vsName = pool.Get("vsName").ClonedString()
				servers = strings.Split(replaceStr.Replace(pool.Get("servers").ClonedString()), ",")
			}
			svmPoolMap[vsName] = append(svmPoolMap[vsName], servers...)
		}
	}
	return svmPoolMap
}
