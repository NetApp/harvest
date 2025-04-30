package version

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/cisco/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"strconv"
	"time"
)

const (
	labels = "labels"
)

type Version struct {
	*plugin.AbstractPlugin
	matrix         *matrix.Matrix
	client         *rest.Client
	templateObject string // object name from the template
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &Version{AbstractPlugin: p}
}

func (v *Version) Init(_ conf.Remote) error {
	var (
		client *rest.Client
		err    error
	)

	if err = v.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)

	if client, err = rest.New(conf.ZapiPoller(v.ParentParams), timeout, v.Auth); err != nil {
		return fmt.Errorf("error creating new client: %w", err)
	}

	v.client = client
	v.templateObject = v.ParentParams.GetChildContentS("object")

	v.matrix = matrix.New(v.Parent+".Version", v.templateObject, v.templateObject)

	return nil
}

func (v *Version) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
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
	output, err := v.client.CallAPI(command)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	v.parseVersion(output, versionMat)

	v.client.Metadata.NumCalls = 1
	v.client.Metadata.BytesRx = uint64(len(output.Raw))
	v.client.Metadata.PluginInstances = uint64(len(versionMat.GetInstances()))

	return []*matrix.Matrix{versionMat}, v.client.Metadata, nil
}

func (v *Version) initMatrix(name string) (*matrix.Matrix, error) {

	mat := matrix.New(v.Parent+name, name, name)

	if err := matrix.CreateMetric(labels, mat); err != nil {
		return nil, fmt.Errorf("error while creating metric %s: %w", labels, err)
	}

	return mat, nil
}

func (v *Version) parseVersion(output gjson.Result, versionMat *matrix.Matrix) {
	biosVersion := output.Get("bios_ver_str").ClonedString()
	chassis := output.Get("chassis_id").ClonedString()
	hostname := output.Get("host_name").ClonedString()
	osVersion := output.Get("nxos_ver_str").ClonedString()

	uptmDays := output.Get("kern_uptm_days").Float()
	uptmHrs := output.Get("kern_uptm_hrs").Float()
	uptmMins := output.Get("kern_uptm_mins").Float()
	uptmSeconds := output.Get("kern_uptm_secs").Float()
	totalSeconds := (60 * 60 * 24 * uptmDays) + (60 * 60 * uptmHrs) + (60 * uptmMins) + uptmSeconds
	upTime := strconv.FormatFloat(totalSeconds, 'f', -1, 64)

	instanceKey := chassis
	instance, err := versionMat.NewInstance(instanceKey)
	if err != nil {
		v.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
		return
	}

	instance.SetLabel("biosVersion", biosVersion)
	instance.SetLabel("chassis", chassis)
	instance.SetLabel("hostname", hostname)
	instance.SetLabel("osVersion", osVersion)
	instance.SetLabel("upTime", upTime)

	v.setMetricValue(labels, instance, 1.0, versionMat)
}

func (v *Version) setMetricValue(metric string, instance *matrix.Instance, value float64, mat *matrix.Matrix) {
	if err := mat.GetMetric(metric).SetValueFloat64(instance, value); err != nil {
		v.SLogger.Error(
			"Unable to set value on metric",
			slogx.Err(err),
			slog.String("metric", metric),
		)
	}
}
