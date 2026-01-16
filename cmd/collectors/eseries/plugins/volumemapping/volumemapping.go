package volumemapping

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/netapp/harvest/v2/cmd/collectors/eseries/cluster"
	"github.com/netapp/harvest/v2/cmd/collectors/eseries/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/auth"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

type VolumeMapping struct {
	*plugin.AbstractPlugin
	client          *rest.Client
	lunMapping      *matrix.Matrix
	poolMapping     *matrix.Matrix
	workloadMapping *matrix.Matrix
	hostMapping     *matrix.Matrix
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

	// Matrix 1: LUN Mappings (1-to-many: volume → hosts/clusters)
	v.lunMapping = matrix.New(v.Parent+".LunMapping", "eseries_volume_lun", "eseries_volume_lun")
	exportOptions1 := node.NewS("export_options")
	instanceKeys1 := exportOptions1.NewChildS("instance_keys", "")
	instanceKeys1.NewChildS("", "volume")
	instanceKeys1.NewChildS("", "lun")
	instanceKeys1.NewChildS("", "host")
	instanceKeys1.NewChildS("", "type")
	v.lunMapping.SetExportOptions(exportOptions1)

	// Create metric for LUN mapping
	if _, err := v.lunMapping.NewMetricUint8("mapping"); err != nil {
		return err
	}

	// Matrix 2: Pool Mappings (1-to-1: volume → pool)
	v.poolMapping = matrix.New(v.Parent+".PoolMapping", "eseries_volume_pool", "eseries_volume_pool")
	exportOptions2 := node.NewS("export_options")
	instanceKeys2 := exportOptions2.NewChildS("instance_keys", "")
	instanceKeys2.NewChildS("", "volume")
	instanceKeys2.NewChildS("", "pool")
	v.poolMapping.SetExportOptions(exportOptions2)

	if _, err := v.poolMapping.NewMetricUint8("mapping"); err != nil {
		return err
	}

	// Matrix 3: Workload Mappings (1-to-1: volume → workload)
	v.workloadMapping = matrix.New(v.Parent+".WorkloadMapping", "eseries_volume_workload", "eseries_volume_workload")
	exportOptions3 := node.NewS("export_options")
	instanceKeys3 := exportOptions3.NewChildS("instance_keys", "")
	instanceKeys3.NewChildS("", "volume")
	instanceKeys3.NewChildS("", "workload")
	v.workloadMapping.SetExportOptions(exportOptions3)

	if _, err := v.workloadMapping.NewMetricUint8("mapping"); err != nil {
		return err
	}

	// Matrix 4: Host Mappings (1-to-1: volume → host)
	v.hostMapping = matrix.New(v.Parent+".HostMapping", "eseries_volume_host", "eseries_volume_host")
	exportOptions4 := node.NewS("export_options")
	instanceKeys4 := exportOptions4.NewChildS("instance_keys", "")
	instanceKeys4.NewChildS("", "volume")
	instanceKeys4.NewChildS("", "host")
	instanceKeys4.NewChildS("", "type")
	v.hostMapping.SetExportOptions(exportOptions4)

	if _, err := v.hostMapping.NewMetricUint8("mapping"); err != nil {
		return err
	}

	return nil
}

func (v *VolumeMapping) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[v.Object]

	// Purge and reset all matrices
	v.lunMapping.PurgeInstances()
	v.lunMapping.Reset()
	v.poolMapping.PurgeInstances()
	v.poolMapping.Reset()
	v.workloadMapping.PurgeInstances()
	v.workloadMapping.Reset()
	v.hostMapping.PurgeInstances()
	v.hostMapping.Reset()

	// Set global labels for all matrices
	globalLabels := data.GetGlobalLabels()
	v.lunMapping.SetGlobalLabels(globalLabels)
	v.poolMapping.SetGlobalLabels(globalLabels)
	v.workloadMapping.SetGlobalLabels(globalLabels)
	v.hostMapping.SetGlobalLabels(globalLabels)

	// Get systemID from ParentParams
	systemID := v.ParentParams.GetChildContentS("system_id")
	if systemID == "" {
		v.SLogger.Warn("systemID not found in ParentParams, skipping volume mapping")
		return []*matrix.Matrix{v.lunMapping, v.poolMapping, v.workloadMapping, v.hostMapping}, nil, nil
	}

	// Build lookup maps
	poolNames, err := v.buildPoolLookup(systemID)
	if err != nil {
		v.SLogger.Warn("Failed to build pool lookup", slogx.Err(err))
	}

	hostNames, err := v.buildHostLookup(systemID)
	if err != nil {
		v.SLogger.Warn("Failed to build host lookup", slogx.Err(err))
	}

	clusterNames, err := cluster.BuildClusterLookup(v.client, systemID, v.SLogger)
	if err != nil {
		v.SLogger.Warn("Failed to build cluster lookup", slogx.Err(err))
	}

	workloadNames, err := v.buildWorkloadLookup(systemID)
	if err != nil {
		v.SLogger.Warn("Failed to build workload lookup", slogx.Err(err))
	}

	// Process each volume instance
	for _, volumeInstance := range data.GetInstances() {
		volumeName := volumeInstance.GetLabel("volume")
		if volumeName == "" {
			continue
		}

		// Process LUN mappings (1-to-many)
		v.processLunMappings(volumeInstance, volumeName, hostNames, clusterNames)

		// Process pool mapping (1-to-1)
		v.processPoolMapping(volumeInstance, volumeName, poolNames)

		// Process workload mapping (1-to-1)
		v.processWorkloadMapping(volumeInstance, volumeName, workloadNames)

		// Process host mapping (1-to-1)
		v.processHostMapping(volumeInstance, volumeName, hostNames, clusterNames)
	}

	return []*matrix.Matrix{v.lunMapping, v.poolMapping, v.workloadMapping, v.hostMapping}, nil, nil
}

func (v *VolumeMapping) buildPoolLookup(systemID string) (map[string]string, error) {
	poolNames := make(map[string]string)

	apiPath := v.client.GetAPIPath() + "/storage-systems/" + systemID + "/storage-pools"
	pools, err := v.client.Fetch(apiPath, nil)
	if err != nil {
		return poolNames, fmt.Errorf("failed to fetch pools: %w", err)
	}

	for _, pool := range pools {
		poolRef := pool.Get("id").String()
		if poolRef == "" {
			poolRef = pool.Get("volumeGroupRef").String()
		}
		poolName := pool.Get("name").String()
		if poolName == "" {
			poolName = pool.Get("label").String()
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

	apiPath := v.client.GetAPIPath() + "/storage-systems/" + systemID + "/hosts"
	hosts, err := v.client.Fetch(apiPath, nil)
	if err != nil {
		return hostNames, fmt.Errorf("failed to fetch hosts: %w", err)
	}

	for _, host := range hosts {
		hostRef := host.Get("hostRef").String()
		if hostRef == "" {
			hostRef = host.Get("id").String()
		}
		hostName := host.Get("name").String()
		if hostName == "" {
			hostName = host.Get("label").String()
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

	apiPath := v.client.GetAPIPath() + "/storage-systems/" + systemID + "/workloads"
	workloads, err := v.client.Fetch(apiPath, nil)
	if err != nil {
		return workloadNames, fmt.Errorf("failed to fetch workloads: %w", err)
	}

	for _, workload := range workloads {
		workloadID := workload.Get("id").String()
		workloadName := workload.Get("name").String()
		if workloadID != "" && workloadName != "" {
			workloadNames[workloadID] = workloadName
		}
	}

	v.SLogger.Debug("built workload lookup", slog.Int("count", len(workloadNames)))
	return workloadNames, nil
}

func (v *VolumeMapping) processLunMappings(volumeInstance *matrix.Instance, volumeName string, hostNames, clusterNames map[string]string) {
	listOfMappingsJSON := volumeInstance.GetLabel("list_of_mappings")
	if listOfMappingsJSON == "" || listOfMappingsJSON == "[]" {
		return
	}

	mappings := gjson.Parse(listOfMappingsJSON)
	if !mappings.IsArray() {
		return
	}

	for _, mapping := range mappings.Array() {
		lun := mapping.Get("lun").String()
		mapRef := mapping.Get("mapRef").String()
		mapType := mapping.Get("type").String()

		if lun == "" || mapRef == "" {
			continue
		}

		// Create composite instance key: volume#lun#mapRef#lunNumber
		instanceKey := volumeName + "#lun#" + mapRef + "#" + lun

		instance, err := v.lunMapping.NewInstance(instanceKey)
		if err != nil {
			v.SLogger.Warn("Failed to create LUN mapping instance",
				slog.String("key", instanceKey),
				slogx.Err(err))
			continue
		}

		// Set labels
		instance.SetLabel("volume", volumeName)
		instance.SetLabel("lun", lun)
		instance.SetLabel("type", mapType)
		// Resolve host or cluster name
		switch mapType {
		case "cluster":
			if clusterName, ok := clusterNames[mapRef]; ok {
				instance.SetLabel("host", clusterName)
			} else {
				instance.SetLabel("host", mapRef)
			}
		case "host":
			if hostName, ok := hostNames[mapRef]; ok {
				instance.SetLabel("host", hostName)
			} else {
				instance.SetLabel("host", mapRef)
			}
		default:
			// Unknown type - log and skip this instance
			v.SLogger.Warn("Unknown mapping type, skipping host mapping",
				slog.String("volume", volumeName),
				slog.String("type", mapType),
				slog.String("mapRef", mapRef))
			return
		}

		v.lunMapping.GetMetric("mapping").SetValueUint8(instance, 1)
	}
}

func (v *VolumeMapping) processPoolMapping(volumeInstance *matrix.Instance, volumeName string, poolNames map[string]string) {
	poolRef := volumeInstance.GetLabel("volume_group_ref")
	if poolRef == "" {
		return
	}

	instanceKey := volumeName + "#pool#" + poolRef

	instance, err := v.poolMapping.NewInstance(instanceKey)
	if err != nil {
		v.SLogger.Warn("Failed to create pool mapping instance",
			slog.String("key", instanceKey),
			slogx.Err(err))
		return
	}

	instance.SetLabel("volume", volumeName)

	if poolName, ok := poolNames[poolRef]; ok {
		instance.SetLabel("pool", poolName)
	} else {
		instance.SetLabel("pool", poolRef)
	}

	v.poolMapping.GetMetric("mapping").SetValueUint8(instance, 1)
}

func (v *VolumeMapping) processWorkloadMapping(volumeInstance *matrix.Instance, volumeName string, workloadNames map[string]string) {
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
		if item.Get("key").String() == "workloadId" {
			workloadID = item.Get("value").String()
			break
		}
	}

	if workloadID == "" {
		return
	}

	// Create instance key: volume#workload#workloadID
	instanceKey := volumeName + "#workload#" + workloadID

	instance, err := v.workloadMapping.NewInstance(instanceKey)
	if err != nil {
		v.SLogger.Warn("Failed to create workload mapping instance",
			slog.String("key", instanceKey),
			slogx.Err(err))
		return
	}

	instance.SetLabel("volume", volumeName)

	if workloadName, ok := workloadNames[workloadID]; ok {
		instance.SetLabel("workload", workloadName)
	} else {
		instance.SetLabel("workload", workloadID)
	}

	v.workloadMapping.GetMetric("mapping").SetValueUint8(instance, 1)
}

func (v *VolumeMapping) processHostMapping(volumeInstance *matrix.Instance, volumeName string, hostNames, clusterNames map[string]string) {
	listOfMappingsJSON := volumeInstance.GetLabel("list_of_mappings")
	if listOfMappingsJSON == "" || listOfMappingsJSON == "[]" {
		return
	}

	mappings := gjson.Parse(listOfMappingsJSON)
	if !mappings.IsArray() {
		return
	}

	// Get the first mapping to establish the 1-to-1 relationship
	// This represents the primary host/cluster for this volume
	if len(mappings.Array()) == 0 {
		return
	}

	firstMapping := mappings.Array()[0]
	mapRef := firstMapping.Get("mapRef").String()
	mapType := firstMapping.Get("type").String()

	if mapRef == "" {
		return
	}

	// Create instance key: volume#host#mapRef
	instanceKey := volumeName + "#host#" + mapRef

	instance, err := v.hostMapping.NewInstance(instanceKey)
	if err != nil {
		v.SLogger.Warn("Failed to create host mapping instance",
			slog.String("key", instanceKey),
			slogx.Err(err))
		return
	}

	instance.SetLabel("volume", volumeName)
	instance.SetLabel("type", mapType)

	switch mapType {
	case "cluster":
		if clusterName, ok := clusterNames[mapRef]; ok {
			instance.SetLabel("host", clusterName)
		} else {
			instance.SetLabel("host", mapRef)
		}
	case "host":
		if hostName, ok := hostNames[mapRef]; ok {
			instance.SetLabel("host", hostName)
		} else {
			instance.SetLabel("host", mapRef)
		}
	default:
		// Unknown type - log and skip this instance
		v.SLogger.Warn("Unknown mapping type, skipping host mapping",
			slog.String("volume", volumeName),
			slog.String("type", mapType),
			slog.String("mapRef", mapRef))
		return
	}

	v.hostMapping.GetMetric("mapping").SetValueUint8(instance, 1)
}
