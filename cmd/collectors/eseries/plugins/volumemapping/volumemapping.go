package volumemapping

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/eseries/hostcluster"
	"log/slog"
	"strings"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

type VolumeMapping struct {
	*plugin.AbstractPlugin
	client           *rest.Client
	schedule         int
	poolNames        map[string]string
	hostNames        map[string]string
	hostClusterNames map[string]string
	workloadNames    map[string]string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &VolumeMapping{AbstractPlugin: p}
}

func (v *VolumeMapping) Init(remote conf.Remote) error {
	if err := v.InitAbc(); err != nil {
		return err
	}

	// Initialize REST client
	timeout, _ := time.ParseDuration(rest.DefaultTimeout)
	poller, err := conf.PollerNamed(v.Options.Poller)
	if err != nil {
		return err
	}

	credentials := auth.NewCredentials(poller, v.SLogger)
	if v.client, err = rest.New(poller, timeout, credentials, ""); err != nil {
		return err
	}

	if err := v.client.Init(1, remote); err != nil {
		return err
	}

	v.poolNames = make(map[string]string)
	v.hostNames = make(map[string]string)
	v.hostClusterNames = make(map[string]string)
	v.workloadNames = make(map[string]string)

	v.schedule = v.SetPluginInterval()
	return nil
}

func (v *VolumeMapping) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[v.Object]

	// Get arrayID from ParentParams
	arrayID := v.ParentParams.GetChildContentS("array_id")
	if arrayID == "" {
		v.SLogger.Warn("arrayID not found in ParentParams, skipping volume mapping")
		return nil, nil, nil
	}

	// Rebuild caches at plugin interval to reduce API calls
	if v.schedule >= v.PluginInvocationRate {
		v.schedule = 0
		v.rebuildCaches(arrayID)
	}
	v.schedule++

	v.applyLabelsToVolumes(data)

	return nil, nil, nil
}

func (v *VolumeMapping) rebuildCaches(arrayID string) {
	// Clear existing caches
	v.poolNames = make(map[string]string)
	v.hostNames = make(map[string]string)
	v.hostClusterNames = make(map[string]string)
	v.workloadNames = make(map[string]string)

	// Build lookup maps and cache them
	poolNames, err := v.buildPoolLookup(arrayID)
	if err != nil {
		v.SLogger.Warn("Failed to build pool lookup", slogx.Err(err))
	} else {
		v.poolNames = poolNames
	}

	hostNames, err := v.buildHostLookup(arrayID)
	if err != nil {
		v.SLogger.Warn("Failed to build host lookup", slogx.Err(err))
	} else {
		v.hostNames = hostNames
	}

	hostClusterNames, err := hostcluster.BuildHostClusterLookup(v.client, arrayID, v.SLogger)
	if err != nil {
		v.SLogger.Warn("Failed to build host group lookup", slogx.Err(err))
	} else {
		v.hostClusterNames = hostClusterNames
	}

	workloadNames, err := v.buildWorkloadLookup(arrayID)
	if err != nil {
		v.SLogger.Warn("Failed to build workload lookup", slogx.Err(err))
	} else {
		v.workloadNames = workloadNames
	}
}

func (v *VolumeMapping) applyLabelsToVolumes(data *matrix.Matrix) {
	// Process each volume instance
	for _, volumeInstance := range data.GetInstances() {
		volumeName := volumeInstance.GetLabel("volume")
		if volumeName == "" {
			continue
		}

		// Add pool label
		v.addPoolLabel(volumeInstance, v.poolNames)

		// Add workload label
		v.addWorkloadLabel(volumeInstance, v.workloadNames)

		// Add LUN, host and type labels (comma-separated)
		v.addLunAndHostLabels(volumeInstance, v.hostNames, v.hostClusterNames)
	}
}

func (v *VolumeMapping) buildPoolLookup(systemID string) (map[string]string, error) {
	poolNames := make(map[string]string)

	apiPath := v.client.APIPath + "/storage-systems/" + systemID + "/storage-pools"
	pools, err := v.client.Fetch(apiPath, nil)
	if err != nil {
		return poolNames, fmt.Errorf("failed to fetch pools: %w", err)
	}

	for _, pool := range pools {
		poolRef := pool.Get("id").ClonedString()
		if poolRef == "" {
			poolRef = pool.Get("volumeGroupRef").ClonedString()
		}
		poolName := pool.Get("name").ClonedString()
		if poolName == "" {
			poolName = pool.Get("label").ClonedString()
		}
		if poolRef != "" && poolName != "" {
			poolNames[poolRef] = poolName
		}
	}

	v.SLogger.Debug("built pool lookup", slog.Int("count", len(poolNames)))
	return poolNames, nil
}

func (v *VolumeMapping) buildHostLookup(systemID string) (map[string]string, error) {
	hostNames := make(map[string]string)

	apiPath := v.client.APIPath + "/storage-systems/" + systemID + "/hosts"
	hosts, err := v.client.Fetch(apiPath, nil)
	if err != nil {
		return hostNames, fmt.Errorf("failed to fetch hosts: %w", err)
	}

	for _, host := range hosts {
		hostRef := host.Get("hostRef").ClonedString()
		if hostRef == "" {
			hostRef = host.Get("id").ClonedString()
		}
		hostName := host.Get("name").ClonedString()
		if hostName == "" {
			hostName = host.Get("label").ClonedString()
		}
		if hostRef != "" && hostName != "" {
			hostNames[hostRef] = hostName
		}
	}

	v.SLogger.Debug("built host lookup", slog.Int("count", len(hostNames)))
	return hostNames, nil
}

func (v *VolumeMapping) buildWorkloadLookup(systemID string) (map[string]string, error) {
	workloadNames := make(map[string]string)

	apiPath := v.client.APIPath + "/storage-systems/" + systemID + "/workloads"
	workloads, err := v.client.Fetch(apiPath, nil)
	if err != nil {
		return workloadNames, fmt.Errorf("failed to fetch workloads: %w", err)
	}

	for _, workload := range workloads {
		workloadID := workload.Get("id").ClonedString()
		workloadName := workload.Get("name").ClonedString()
		if workloadID != "" && workloadName != "" {
			workloadNames[workloadID] = workloadName
		}
	}

	v.SLogger.Debug("built workload lookup", slog.Int("count", len(workloadNames)))
	return workloadNames, nil
}

func (v *VolumeMapping) addLunAndHostLabels(volumeInstance *matrix.Instance, hostNames, hostClusterNames map[string]string) {
	listOfMappingsJSON := volumeInstance.GetLabel("list_of_mappings")
	if listOfMappingsJSON == "" || listOfMappingsJSON == "[]" {
		return
	}

	mappings := gjson.Parse(listOfMappingsJSON)
	if !mappings.IsArray() {
		return
	}

	var luns, hosts, types []string

	for _, mapping := range mappings.Array() {
		lun := mapping.Get("lun").ClonedString()
		mapRef := mapping.Get("mapRef").ClonedString()
		mapType := mapping.Get("type").ClonedString()

		if mapRef == "" {
			continue
		}

		if lun != "" {
			luns = append(luns, lun)
		}
		types = append(types, mapType)

		// Resolve host or host cluster name
		var hostName string
		switch mapType {
		case "cluster":
			if name, ok := hostClusterNames[mapRef]; ok {
				hostName = name
			} else {
				hostName = mapRef
			}
		case "host":
			if name, ok := hostNames[mapRef]; ok {
				hostName = name
			} else {
				hostName = mapRef
			}
		default:
			v.SLogger.Warn("Unknown mapping type",
				slog.String("volume", volumeInstance.GetLabel("volume")),
				slog.String("type", mapType),
				slog.String("mapRef", mapRef))
			hostName = mapRef
		}
		hosts = append(hosts, hostName)
	}

	if len(luns) > 0 {
		volumeInstance.SetLabel("luns", strings.Join(luns, ","))
	}

	if len(hosts) > 0 {
		volumeInstance.SetLabel("hosts", strings.Join(hosts, ","))
		volumeInstance.SetLabel("mapping_types", strings.Join(types, ","))
	}
}

func (v *VolumeMapping) addPoolLabel(volumeInstance *matrix.Instance, poolNames map[string]string) {
	poolRef := volumeInstance.GetLabel("volume_group_ref")
	if poolRef == "" {
		return
	}

	if poolName, ok := poolNames[poolRef]; ok {
		volumeInstance.SetLabel("pool", poolName)
	} else {
		volumeInstance.SetLabel("pool", poolRef)
	}
}

func (v *VolumeMapping) addWorkloadLabel(volumeInstance *matrix.Instance, workloadNames map[string]string) {
	metadataJSON := volumeInstance.GetLabel("metadata")
	if metadataJSON == "" || metadataJSON == "[]" {
		return
	}

	metadata := gjson.Parse(metadataJSON)
	if !metadata.IsArray() {
		return
	}

	var workloadID string
	for _, item := range metadata.Array() {
		if item.Get("key").ClonedString() == "workloadId" {
			workloadID = item.Get("value").ClonedString()
			break
		}
	}

	if workloadID == "" {
		return
	}

	if workloadName, ok := workloadNames[workloadID]; ok {
		volumeInstance.SetLabel("workload", workloadName)
	} else {
		volumeInstance.SetLabel("workload", workloadID)
	}
}
