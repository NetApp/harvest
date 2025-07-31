package lldp

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

type LLDP struct {
	*plugin.AbstractPlugin
	matrix         *matrix.Matrix
	client         *rest.Client
	templateObject string // object name from the template
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

	timeout, _ := time.ParseDuration(rest.DefaultTimeout)

	if client, err = rest.New(conf.ZapiPoller(l.ParentParams), timeout, l.Auth); err != nil {
		return fmt.Errorf("error creating new client: %w", err)
	}

	if err := client.Init(2, remote); err != nil {
		return err
	}

	l.client = client
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
	output, err := l.client.CLIShowArray(command)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	l.parseLLDP(output, lldpMat)

	l.client.Metadata.NumCalls = 1
	l.client.Metadata.BytesRx = uint64(len(output.Raw))
	l.client.Metadata.PluginInstances = uint64(len(lldpMat.GetInstances()))

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

	rowQuery := "output.body.TABLE_nbor_detail.ROW_nbor_detail"

	var models []Model

	rows := output.Get(rowQuery)

	if !rows.Exists() {
		l.SLogger.Warn("Unable to parse LLDP because rows are missing", slog.String("query", rowQuery))
		return
	}

	rows.ForEach(func(_, value gjson.Result) bool {
		lldpModel := NewLLDPModel(value)
		models = append(models, lldpModel)
		return true
	})

	for _, model := range models {
		instanceKey := model.ChassisID
		instance, err := mat.NewInstance(instanceKey)
		if err != nil {
			l.SLogger.Warn("Failed to create lldp instance", slog.String("key", instanceKey))
			continue
		}

		instance.SetLabel("remote_name", model.RemoteName)
		instance.SetLabel("remote_platform", model.RemotePlatform)
		instance.SetLabel("chassis", model.ChassisID)
		instance.SetLabel("local_port", model.LocalPort)
		instance.SetLabel("remote_port", model.RemotePort)
		instance.SetLabel("capabilities", strings.Join(model.Capabilities, ","))
		instance.SetLabel("local_platform", l.client.Remote().Serial)

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

func NewLLDPModel(output gjson.Result) Model {

	var m Model

	m.RemoteName = output.Get("sys_name").ClonedString()
	m.RemotePlatform = output.Get("sys_desc").ClonedString()
	m.ChassisID = output.Get("chassis_id").ClonedString()
	m.LocalPort = output.Get("l_port_id").ClonedString()
	m.TTL = output.Get("ttl").Int()
	m.RemotePort = output.Get("port_id").ClonedString()
	m.RemoteDesc = output.Get("port_desc").ClonedString()
	m.Capabilities = lldpCapabilities(output.Get("enabled_capability").String())

	if m.RemotePlatform == "null" {
		m.RemotePlatform = ""
	}
	if m.RemoteName == "null" {
		m.RemoteName = ""
	}
	if m.RemoteDesc == "null" {
		m.RemoteDesc = ""
	}

	return m
}

func lldpCapabilities(capStr string) []string {
	// show lldp neighbors detail
	// "system_capability" : "B, R",
	// Capability codes:
	//  (R) Router, (B) Bridge, (T) Telephone, (C) DOCSIS Cable Device
	//  (W) WLAN Access Point, (P) Repeater, (S) Station, (O) Other

	var (
		capabilities []string
		code         string
	)

	splits := strings.Split(capStr, ",")
	for _, split := range splits {
		letter := strings.TrimSpace(split)
		// Ignore empty strings
		if letter == "" {
			continue
		}

		switch letter {
		case "R":
			code = "Router"
		case "B":
			code = "Bridge"
		case "T":
			code = "Telephone"
		case "C":
			code = "DOCSIS Cable Device"
		case "W":
			code = "WLAN Access Point"
		case "P":
			code = "Repeater"
		case "S":
			code = "Station"
		case "O":
			code = "Other"
		default:
			code = fmt.Sprintf("Unknown (%s)", letter)
		}

		if code != "" {
			capabilities = append(capabilities, code)
		}
	}

	slices.Sort(capabilities)

	return capabilities
}
