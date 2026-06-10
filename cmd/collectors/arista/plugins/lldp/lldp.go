package lldp

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/arista/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"sort"
	"strings"
)

const (
	labels = "labels"
)

type LLDP struct {
	*plugin.AbstractPlugin
	matrix         *matrix.Matrix
	client         *rest.Client
	templateObject string // object name from the template
	RemoteSerial   string
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &LLDP{AbstractPlugin: p}
}

func (l *LLDP) Init(remote conf.Remote) error {
	var (
		client *rest.Client
		err    error
	)

	if err = l.InitAbc(); err != nil {
		return fmt.Errorf("failed to initialize AbstractPlugin: %w", err)
	}

	if client, err = rest.New(conf.ZapiPoller(l.ParentParams), l.Auth); err != nil {
		return fmt.Errorf("error creating new client: %w", err)
	}

	if err := client.Init(2, remote); err != nil {
		return err
	}

	l.client = client
	l.RemoteSerial = client.Remote().Serial
	l.templateObject = l.ParentParams.GetChildContentS("object")

	l.matrix = matrix.New(l.Parent+".LLDP", l.templateObject, l.templateObject)

	return nil
}

func (l *LLDP) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *collector.Metadata, error) {
	data := dataMap[l.Object]
	l.client.Metadata.Reset()

	lldpMat, err := l.initMatrix(l.templateObject)
	if err != nil {
		return nil, nil, fmt.Errorf("error while initializing matrix: %w", err)
	}

	// Set all global labels if they don't already exist
	lldpMat.SetGlobalLabels(data.GetGlobalLabels())

	data.Reset()

	command := l.ParentParams.GetChildContentS("query")
	output, err := l.client.RunCmds(rest.SplitCommands(command)...)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	l.parseLLDP(output, lldpMat)

	l.client.Metadata.NumCalls.Store(1)
	l.client.Metadata.BytesRx.Store(uint64(len(output.Raw)))
	l.client.Metadata.PluginInstances.Store(uint64(len(lldpMat.GetInstances())))

	return []*matrix.Matrix{lldpMat}, l.client.Metadata, nil
}

func (l *LLDP) initMatrix(name string) (*matrix.Matrix, error) {

	mat := matrix.New(l.Parent+name, name, name)

	if err := matrix.CreateMetric(labels, mat); err != nil {
		return nil, fmt.Errorf("error while creating metric %s: %w", labels, err)
	}

	return mat, nil
}

func (l *LLDP) parseLLDP(output gjson.Result, mat *matrix.Matrix) {

	rowQuery := "0.lldpNeighbors"

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		l.SLogger.Warn("Unable to parse LLDP because lldpNeighbors are missing", slog.String("query", rowQuery))
		return
	}

	var models []Model

	rows.ForEach(func(localPort, value gjson.Result) bool {
		port := localPort.ClonedString()
		value.Get("lldpNeighborInfo").ForEach(func(_, nbor gjson.Result) bool {
			models = append(models, NewLLDPModel(port, nbor))
			return true
		})
		return true
	})

	for _, model := range models {
		instanceKey := model.ChassisID + "-" + model.LocalPort
		instance, err := mat.NewInstance(instanceKey)
		if err != nil {
			l.SLogger.Warn("Failed to create lldp instance", slog.String("key", instanceKey), slogx.Err(err))
			continue
		}

		instance.SetLabel("remote_name", model.RemoteName)
		instance.SetLabel("remote_platform", model.RemotePlatform)
		instance.SetLabel("chassis", model.ChassisID)
		instance.SetLabel("local_port", model.LocalPort)
		instance.SetLabel("remote_port", model.RemotePort)
		instance.SetLabel("capabilities", strings.Join(model.Capabilities, ","))
		instance.SetLabel("local_platform", l.RemoteSerial)

		mat.GetMetric(labels).SetValueFloat64(instance, 1.0)
	}
}

type Model struct {
	Capabilities   []string
	ChassisID      string
	RemotePlatform string
	RemoteName     string
	LocalPort      string
	RemotePort     string
	RemoteDesc     string
	TTL            int64
}

func NewLLDPModel(localPort string, nbor gjson.Result) Model {

	var m Model

	m.LocalPort = localPort
	m.RemoteName = nbor.Get("systemName").ClonedString()
	m.RemotePlatform = nbor.Get("systemDescription").ClonedString()
	m.RemoteDesc = nbor.Get("systemDescription").ClonedString()
	m.ChassisID = nbor.Get("chassisId").ClonedString()
	m.TTL = nbor.Get("ttl").Int()

	ni := nbor.Get("neighborInterfaceInfo")
	remotePort := ni.Get("interfaceId").ClonedString()
	if remotePort == "" {
		remotePort = ni.Get("interfaceDescription").ClonedString()
	}
	m.RemotePort = rest.TrimQuotes(remotePort)

	// systemCapabilities is a map of capability name -> bool (e.g. {"bridge": true, "router": true})
	caps := nbor.Get("systemCapabilities")
	caps.ForEach(func(capName, enabled gjson.Result) bool {
		if enabled.Bool() {
			m.Capabilities = append(m.Capabilities, capName.ClonedString())
		}
		return true
	})
	sort.Strings(m.Capabilities)

	return m
}
