package version

import (
	"fmt"
	"github.com/netapp/harvest/v2/cmd/collectors/cisco/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"log/slog"
	"regexp"
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
	output, err := v.client.CLIShow(command)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch data: %w", err)
	}

	v.parseVersionAndBanner(output, versionMat)

	v.client.Metadata.NumCalls = 1
	v.client.Metadata.BytesRx = uint64(len(output.Raw))
	v.client.Metadata.PluginInstances = uint64(len(versionMat.GetInstances()))

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

func (v *Version) parseVersionAndBanner(output gjson.Result, versionMat *matrix.Matrix) {

	versionOutput := output.Get("output.0.body")
	bannerOutput := output.Get("output.1.body")

	biosVersion := versionOutput.Get("bios_ver_str").ClonedString()
	chassis := versionOutput.Get("chassis_id").ClonedString()
	hostname := versionOutput.Get("host_name").ClonedString()
	osVersion := versionOutput.Get("nxos_ver_str").ClonedString()

	uptmDays := versionOutput.Get("kern_uptm_days").Float()
	uptmHrs := versionOutput.Get("kern_uptm_hrs").Float()
	uptmMins := versionOutput.Get("kern_uptm_mins").Float()
	uptmSeconds := versionOutput.Get("kern_uptm_secs").Float()
	uptimeSeconds := (60 * 60 * 24 * uptmDays) + (60 * 60 * uptmHrs) + (60 * uptmMins) + uptmSeconds

	instanceKey := chassis
	instance, err := versionMat.NewInstance(instanceKey)
	if err != nil {
		v.SLogger.Warn("Failed to create instance", slog.String("key", instanceKey))
		return
	}

	bannerMsg := bannerOutput.Get("banner_msg.b_msg").ClonedString()
	anRCF := parseRCF(bannerMsg)
	if anRCF.Filename == "" {
		v.SLogger.Warn("Failed to parse RCF filename", slog.String("banner", bannerOutput.Raw))
	}

	if anRCF.Version == "" {
		v.SLogger.Warn("Failed to parse RCF version", slog.String("banner", bannerOutput.Raw))
	}

	instance.SetLabel("biosVersion", biosVersion)
	instance.SetLabel("chassis", chassis)
	instance.SetLabel("hostname", hostname)
	instance.SetLabel("osVersion", osVersion)
	instance.SetLabel("rcf_filename", anRCF.Filename)
	instance.SetLabel("rcf_version", anRCF.Version)

	versionMat.GetMetric(labels).SetValueFloat64(instance, 1.0)
	versionMat.GetMetric(uptime).SetValueFloat64(instance, uptimeSeconds)
}

var filenameRegex = regexp.MustCompile(`(?m)Filename\s+:\s+(.*?)$`)
var generatorRegex = regexp.MustCompile(`Generator:\s+([^\s_]+)`)
var versionRegexes = []*regexp.Regexp{
	regexp.MustCompile(`Version\s+:\s+(.*?)$`),
	regexp.MustCompile(`Generator version:\s+([^\s_]+)`),
	regexp.MustCompile(`version\s+(.*?)\s+`),
}

type rcf struct {
	Filename string
	Version  string
}

func parseRCF(banner string) rcf {

	var anRCF rcf

	parseFilename(banner, &anRCF)
	parseVersion(banner, &anRCF)

	return anRCF
}

func parseVersion(banner string, r *rcf) {
	matches := generatorRegex.FindStringSubmatch(banner)
	if len(matches) == 2 {
		r.Version = matches[1]
		return
	}

	for _, regex := range versionRegexes {
		matches = regex.FindStringSubmatch(banner)
		if len(matches) == 2 {
			r.Version = matches[1]
			break
		}
	}
}

func parseFilename(banner string, r *rcf) {
	matches := filenameRegex.FindStringSubmatch(banner)
	if len(matches) == 2 {
		r.Filename = matches[1]
	}
}
