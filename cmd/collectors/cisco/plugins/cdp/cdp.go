package cdp

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/cisco/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"slices"
	"strings"
	"time"
)

const (
	labels = "labels"
)

type CDP struct {
	*plugin.AbstractPlugin
	matrix         *matrix.Matrix
	client         *rest.Client
	templateObject string // object name from the template
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &CDP{AbstractPlugin: p}
}

func (c *CDP) Init(remote conf.Remote) error {
	var (
		client *rest.Client
		err    error
	)

	if err = c.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)

	if client, err = rest.New(conf.ZapiPoller(c.ParentParams), timeout, c.Auth); err != nil {
		return fmt.Errorf("error creating new client: %w", err)
	}

	if err := client.Init(2, remote); err != nil {
		return err
	}

	c.client = client
	c.templateObject = c.ParentParams.GetChildContentS("object")

	c.matrix = matrix.New(c.Parent+".CDP", c.templateObject, c.templateObject)

	return nil
}

func (c *CDP) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[c.Object]
	c.client.Metadata.Reset()

	cdpMat, err := c.initMatrix(c.templateObject)
	if err != nil {
		return nil, nil, fmt.Errorf("error while initializing matrix: %w", err)
	}

	// Set all global labels if they don't already exist
	cdpMat.SetGlobalLabels(data.GetGlobalLabels())

	data.Reset()

	command := c.ParentParams.GetChildContentS("query")
	output, err := c.client.CLIShowArray(command)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	c.parseCDP(output, cdpMat)

	c.client.Metadata.NumCalls = 1
	c.client.Metadata.BytesRx = uint64(len(output.Raw))
	c.client.Metadata.PluginInstances = uint64(len(cdpMat.GetInstances()))

	return []*matrix.Matrix{cdpMat}, c.client.Metadata, nil
}

func (c *CDP) initMatrix(name string) (*matrix.Matrix, error) {

	mat := matrix.New(c.Parent+name, name, name)

	if err := matrix.CreateMetric(labels, mat); err != nil {
		return nil, fmt.Errorf("error while creating metric %s: %w", labels, err)
	}

	return mat, nil
}

func (c *CDP) parseCDP(output gjson.Result, mat *matrix.Matrix) {

	rowQuery := "output.body.TABLE_cdp_neighbor_detail_info.ROW_cdp_neighbor_detail_info"

	var models []Model

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		c.SLogger.Warn("Unable to parse CDP because rows are missing", slog.String("query", rowQuery))
		return
	}

	rows.ForEach(func(_, value gjson.Result) bool {
		cdpModel := NewCDPModel(value)
		models = append(models, cdpModel)
		return true
	})

	for _, model := range models {
		instanceKey := model.RemoteName + "-" + model.RemotePort + "-" + model.LocalPort
		instance, err := mat.NewInstance(instanceKey)
		if err != nil {
			c.SLogger.Warn("Failed to create cdp instance", slog.String("key", instanceKey))
			continue
		}

		instance.SetLabel("capabilities", strings.Join(model.Capabilities, ","))
		instance.SetLabel("local_interface_mac", model.LocalInterfaceMAC)
		instance.SetLabel("local_platform", c.client.Remote().Serial)
		instance.SetLabel("local_port", model.LocalPort)
		instance.SetLabel("remote_interface_mac", model.RemoteInterfaceMAC)
		instance.SetLabel("remote_name", model.RemoteName)
		instance.SetLabel("remote_platform", model.RemotePlatform)
		instance.SetLabel("remote_port", model.RemotePort)
		instance.SetLabel("remote_version", model.RemoteVersion)

		mat.GetMetric(labels).SetValueFloat64(instance, 1.0)
	}
}

type Model struct {
	Capabilities       []string
	RemoteName         string
	LocalInterfaceMAC  string
	RemotePlatform     string
	LocalPort          string
	RemoteInterfaceMAC string
	TTL                int64
	RemoteVersion      string
	RemotePort         string
}

func NewCDPModel(output gjson.Result) Model {

	var m Model

	m.RemoteName = output.Get("device_id").ClonedString()
	m.RemotePlatform = output.Get("platform_id").ClonedString()
	m.RemotePort = output.Get("port_id").ClonedString()
	m.LocalPort = output.Get("intf_id").ClonedString()
	m.TTL = output.Get("ttl").Int()
	m.RemoteVersion = output.Get("version").ClonedString()
	m.LocalInterfaceMAC = output.Get("local_intf_mac").ClonedString()
	m.RemoteInterfaceMAC = output.Get("remote_intf_mac").ClonedString()

	caps := output.Get("capability")
	if caps.IsArray() {
		caps.ForEach(func(_, value gjson.Result) bool {
			m.Capabilities = append(m.Capabilities, value.String())
			return true
		})
	} else if caps.Exists() {
		m.Capabilities = []string{caps.ClonedString()}
	}

	slices.Sort(m.Capabilities)

	return m
}
