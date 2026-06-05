package version

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/arista/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"time"
)

const (
	labels = "labels"
	uptime = "uptime"
)

var metrics = []string{
	labels,
	uptime,
}

type Version struct {
	*plugin.AbstractPlugin
	matrix         *matrix.Matrix
	client         *rest.Client
	templateObject string // object name from the template
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Version{AbstractPlugin: p}
}

func (v *Version) Init(remote conf.Remote) error {
	var (
		client *rest.Client
		err    error
	)

	if err = v.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}

	if client, err = rest.New(conf.ZapiPoller(v.ParentParams), v.Auth); err != nil {
		return fmt.Errorf("error creating new client: %w", err)
	}

	if err := client.Init(2, remote); err != nil {
		return err
	}

	v.client = client
	v.templateObject = v.ParentParams.GetChildContentS("object")

	v.matrix = matrix.New(v.Parent+".Version", v.templateObject, v.templateObject)

	return nil
}

func (v *Version) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[v.Object]
	v.client.Metadata.Reset()

	versionMat, err := v.initMatrix(v.templateObject)
	if err != nil {
		return nil, nil, fmt.Errorf("error while initializing matrix: %w", err)
	}

	// Set all global labels if they don't already exist
	versionMat.SetGlobalLabels(data.GetGlobalLabels())

	data.Reset()

	command := v.ParentParams.GetChildContentS("query")
	output, err := v.client.RunCmds(rest.SplitCommands(command)...)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	v.parseVersion(output, versionMat)

	v.client.Metadata.NumCalls.Store(1)
	v.client.Metadata.BytesRx.Store(uint64(len(output.Raw)))
	v.client.Metadata.PluginInstances.Store(uint64(len(versionMat.GetInstances())))

	return []*matrix.Matrix{versionMat}, v.client.Metadata, nil
}

func (v *Version) initMatrix(name string) (*matrix.Matrix, error) {

	mat := matrix.New(v.Parent+name, name, name)

	for _, k := range metrics {
		if err := matrix.CreateMetric(k, mat); err != nil {
			return nil, fmt.Errorf("error while creating metric %s: %w", k, err)
		}
	}

	return mat, nil
}

func (v *Version) parseVersion(output gjson.Result, versionMat *matrix.Matrix) {

	versionOutput := output.Get("0")
	hostnameOutput := output.Get("1")

	model := versionOutput.Get("modelName").ClonedString()
	osVersion := versionOutput.Get("version").ClonedString()
	internalVersion := versionOutput.Get("internalVersion").ClonedString()
	serial := versionOutput.Get("serialNumber").ClonedString()
	systemMac := versionOutput.Get("systemMacAddress").ClonedString()
	hardwareRevision := versionOutput.Get("hardwareRevision").ClonedString()
	architecture := versionOutput.Get("architecture").ClonedString()

	hostname := hostnameOutput.Get("hostname").ClonedString()
	if hostname == "" {
		hostname = hostnameOutput.Get("fqdn").ClonedString()
	}

	// EOS reports the boot time as a unix timestamp. Uptime is the elapsed time
	// since boot.
	bootupTimestamp := versionOutput.Get("bootupTimestamp").Float()
	var uptimeSeconds float64
	if bootupTimestamp > 0 {
		uptimeSeconds = float64(time.Now().Unix()) - bootupTimestamp
		if uptimeSeconds < 0 {
			uptimeSeconds = 0
		}
	}

	instanceKey := serial
	if instanceKey == "" {
		instanceKey = systemMac
	}
	instance, err := versionMat.NewInstance(instanceKey)
	if err != nil {
		v.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
		return
	}

	instance.SetLabel("model", model)
	instance.SetLabel("serial", serial)
	instance.SetLabel("hostname", hostname)
	instance.SetLabel("osVersion", osVersion)
	instance.SetLabel("internalVersion", internalVersion)
	instance.SetLabel("systemMac", systemMac)
	instance.SetLabel("hardwareRevision", hardwareRevision)
	instance.SetLabel("architecture", architecture)

	versionMat.GetMetric(labels).SetValueFloat64(instance, 1.0)
	versionMat.GetMetric(uptime).SetValueFloat64(instance, uptimeSeconds)
}
