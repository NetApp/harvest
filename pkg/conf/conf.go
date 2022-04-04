/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package conf

import (
	"fmt"
	"github.com/imdario/mergo"
	"goharvest2/pkg/constant"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var Config = HarvestConfig{}
var configRead = false
var ValidatePortInUse = false

const (
	DefaultApiVersion = "1.3"
	DefaultTimeout    = "10s"
)

// TestLoadHarvestConfig is used by testing code to reload a new config
func TestLoadHarvestConfig(configPath string) {
	configRead = false
	err := LoadHarvestConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config at=[%s] err=%+v\n", configPath, err)
	}
}

func ConfigPath(path string) string {
	// env var has higher precedence than --config cmdline arg
	fp := os.Getenv("HARVEST_CONFIG")
	if fp != "" {
		return fp
	}
	return path
}

func LoadHarvestConfig(configPath string) error {
	if configRead {
		return nil
	}
	configPath = ConfigPath(configPath)
	contents, err := ioutil.ReadFile(configPath)

	if err != nil {
		fmt.Printf("error reading config file=[%s] %+v\n", configPath, err)
		return err
	}
	err = yaml.Unmarshal(contents, &Config)
	configRead = true
	if err != nil {
		fmt.Printf("error unmarshalling config file=[%s] %+v\n", configPath, err)
		return err
	}
	// Until https://github.com/go-yaml/yaml/issues/717 is fixed
	// read the yaml again to determine poller order
	orderedConfig := OrderedConfig{}
	err = yaml.Unmarshal(contents, &orderedConfig)
	if err != nil {
		return err
	}
	Config.PollersOrdered = orderedConfig.Pollers.namesInOrder
	for i, name := range Config.PollersOrdered {
		Config.Pollers[name].promIndex = i
	}

	// Merge pollers and defaults
	pollers := Config.Pollers
	defaults := Config.Defaults

	if pollers == nil {
		return errors.New(errors.ERR_CONFIG, "[Pollers] section not found")
	} else if defaults != nil {
		for _, p := range pollers {
			p.Union(defaults)
		}
	}
	return nil
}

func ReadCredentialsFile(credPath string, p *Poller) error {
	contents, err := ioutil.ReadFile(credPath)

	if p == nil {
		return nil
	}
	if err != nil {
		return err
	}
	var credConfig HarvestConfig
	err = yaml.Unmarshal(contents, &credConfig)
	if err != nil {
		return err
	}

	credPoller := credConfig.Pollers[p.Name]
	if credPoller == nil {
		// when the poller is not listed in the file, check if there is a default, and if so, use it
		if credConfig.Defaults != nil {
			credPoller = credConfig.Defaults
		} else {
			return errors.New(errors.INVALID_PARAM, "poller not found in credentials file")
		}
	}
	if credPoller.SslKey != "" {
		p.SslKey = credPoller.SslKey
	}
	if credPoller.SslCert != "" {
		p.SslCert = credPoller.SslCert
	}
	if credPoller.CaCertPath != "" {
		p.CaCertPath = credPoller.CaCertPath
	}
	if credPoller.Username != "" {
		p.Username = credPoller.Username
	}
	if credPoller.Password != "" {
		p.Password = credPoller.Password
	}
	return nil
}

func PollerNamed(name string) (*Poller, error) {
	poller, ok := Config.Pollers[name]
	if !ok {
		return nil, errors.New(errors.ERR_CONFIG, "poller ["+name+"] not found")
	}
	poller.Name = name
	return poller, nil
}

// GetDefaultHarvestConfigPath returns the absolute path of the default harvest config file.
func GetDefaultHarvestConfigPath() string {
	var configPath string
	configFileName := constant.ConfigFileName
	if configPath = os.Getenv("HARVEST_CONF"); configPath == "" {
		configPath = path.Join(GetHarvestHomePath(), configFileName)
	} else {
		configPath = path.Join(configPath, configFileName)
	}
	return configPath
}

// GetHarvestHomePath returns the value of the env var HARVEST_CONF or ./
func GetHarvestHomePath() string {
	harvestConf := os.Getenv("HARVEST_CONF")
	if harvestConf == "" {
		return "./"
	}
	if !strings.HasSuffix(harvestConf, "/") {
		harvestConf += "/"
	}
	return harvestConf
}

func GetHarvestLogPath() string {
	logPath := os.Getenv("HARVEST_LOGS")
	if logPath == "" {
		return "/var/log/harvest/"
	}
	return logPath
}

// GetPrometheusExporterPorts returns the Prometheus port for the given poller
func GetPrometheusExporterPorts(pollerName string) (int, error) {
	var port int
	var isPrometheusExporterConfigured bool

	if len(promPortRangeMapping) == 0 {
		loadPrometheusExporterPortRangeMapping()
	}
	poller := Config.Pollers[pollerName]
	if poller == nil {
		return 0, errors.New(errors.ERR_CONFIG, "Poller does not exist "+pollerName)
	}

	exporters := poller.Exporters
	if len(exporters) > 0 {
		for _, e := range exporters {
			exporter := Config.Exporters[e]
			if exporter.Type == "Prometheus" {
				isPrometheusExporterConfigured = true
				if exporter.PortRange != nil {
					ports := promPortRangeMapping[e]
					preferredPort := exporter.PortRange.Min + poller.promIndex
					_, isFree := ports.freePorts[preferredPort]
					if isFree {
						port = preferredPort
						delete(ports.freePorts, preferredPort)
						break
					}
					for k := range ports.freePorts {
						port = k
						delete(ports.freePorts, k)
						break
					}
				} else if exporter.Port != nil && *exporter.Port != 0 {
					port = *exporter.Port
					break
				}
			}
		}
	}
	if port == 0 && isPrometheusExporterConfigured {
		return port, errors.New(errors.ERR_CONFIG, "No free port found for poller "+pollerName)
	} else {
		return port, nil
	}
}

type PortMap struct {
	portSet   []int
	freePorts map[int]struct{}
}

func PortMapFromRange(address string, portRange *IntRange) PortMap {
	portMap := PortMap{}
	portMap.freePorts = make(map[int]struct{})
	start := portRange.Min
	end := portRange.Max
	for i := start; i <= end; i++ {
		portMap.portSet = append(portMap.portSet, i)
		if ValidatePortInUse {
			portMap.freePorts[i] = struct{}{}
		}
	}
	if !ValidatePortInUse {
		portMap.freePorts = util.CheckFreePorts(address, portMap.portSet)
	}
	return portMap
}

var promPortRangeMapping = make(map[string]PortMap)

func loadPrometheusExporterPortRangeMapping() {
	for k, v := range Config.Exporters {
		if v.Type == "Prometheus" {
			if v.PortRange != nil {
				// we only care about free ports on the localhost
				promPortRangeMapping[k] = PortMapFromRange("localhost", v.PortRange)
			}
		}
	}
}

type IntRange struct {
	Min int
	Max int
}

var rangeRegex, _ = regexp.Compile(`(\d+)\s*-\s*(\d+)`)

func (i *IntRange) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode && node.ShortTag() == "!!str" {
		matches := rangeRegex.FindStringSubmatch(node.Value)
		if len(matches) == 3 {
			min, err1 := strconv.Atoi(matches[1])
			max, err2 := strconv.Atoi(matches[2])
			if err1 != nil {
				return err1
			}
			if err2 != nil {
				return err2
			}
			i.Min = min
			i.Max = max
		}
	}
	return nil
}

// GetUniqueExporters returns the unique set of exporter types from the list of export names
// For example: If 2 prometheus exporters are configured for a poller, the last one is returned
func GetUniqueExporters(exporterNames []string) []string {
	var resultExporters []string
	definedExporters := Config.Exporters
	exporterMap := make(map[string]string)
	for _, ec := range exporterNames {
		e, ok := definedExporters[ec]
		if ok {
			exporterMap[e.Type] = ec
		}
	}

	for _, value := range exporterMap {
		resultExporters = append(resultExporters, value)
	}
	return resultExporters
}

type TLS struct {
	CertFile string `yaml:"cert_file,omitempty"`
	KeyFile  string `yaml:"key_file,omitempty"`
}

type Httpsd struct {
	Listen    string `yaml:"listen,omitempty"`
	AuthBasic struct {
		Username string `yaml:"username,omitempty"`
		Password string `yaml:"password,omitempty"`
	} `yaml:"auth_basic,omitempty"`
	TLS         TLS    `yaml:"tls,omitempty"`
	HeartBeat   string `yaml:"heart_beat,omitempty"`
	ExpireAfter string `yaml:"expire_after,omitempty"`
}

type Admin struct {
	Httpsd Httpsd `yaml:"httpsd,omitempty"`
}

type Tools struct {
	GrafanaApiToken string `yaml:"grafana_api_token,omitempty"`
	AsupDisabled    bool   `yaml:"autosupport_disabled,omitempty"`
}

type Collector struct {
	Name      string    `yaml:"-"`
	Templates *[]string `yaml:"-"`
}

type Poller struct {
	Addr            string                `yaml:"addr,omitempty"`
	ApiVersion      string                `yaml:"api_version,omitempty"`
	ApiVfiler       string                `yaml:"api_vfiler,omitempty"`
	AuthStyle       string                `yaml:"auth_style,omitempty"`
	CaCertPath      string                `yaml:"ca_cert,omitempty"`
	ClientTimeout   string                `yaml:"client_timeout,omitempty"`
	Collectors      []Collector           `yaml:"collectors,omitempty"`
	Datacenter      string                `yaml:"datacenter,omitempty"`
	Exporters       []string              `yaml:"exporters,omitempty"`
	IsKfs           bool                  `yaml:"is_kfs,omitempty"`
	Labels          *[]*map[string]string `yaml:"labels,omitempty"`
	LogMaxBytes     int64                 `yaml:"log_max_bytes,omitempty"`
	LogMaxFiles     int                   `yaml:"log_max_files,omitempty"`
	LogSet          *[]string             `yaml:"log,omitempty"`
	Password        string                `yaml:"password,omitempty"`
	PollerSchedule  string                `yaml:"poller_schedule,omitempty"`
	SslCert         string                `yaml:"ssl_cert,omitempty"`
	SslKey          string                `yaml:"ssl_key,omitempty"`
	UseInsecureTls  *bool                 `yaml:"use_insecure_tls,omitempty"`
	Username        string                `yaml:"username,omitempty"`
	CredentialsFile string                `yaml:"credentials_file,omitempty"`
	promIndex       int
	Name            string
}

func (p *Poller) Union(defaults *Poller) {
	// this is needed because of how mergo handles boolean zero values
	isInsecureNil := true
	var pUseInsecureTls bool
	pIsKfs := p.IsKfs
	if p.UseInsecureTls != nil {
		isInsecureNil = false
		pUseInsecureTls = *p.UseInsecureTls
	}
	_ = mergo.Merge(p, defaults)
	if !isInsecureNil {
		p.UseInsecureTls = &pUseInsecureTls
	}
	p.IsKfs = pIsKfs
}

// ZapiPoller creates a poller out of a node, this is a bridge between the node and struct-based code
// Used by ZAPI based code
func ZapiPoller(n *node.Node) Poller {
	var p Poller

	if Config.Defaults != nil {
		p = *Config.Defaults
	} else {
		p = Poller{}
	}
	p.Name = n.GetChildContentS("poller_name")
	if apiVersion := n.GetChildContentS("api_version"); apiVersion != "" {
		p.ApiVersion = apiVersion
	} else {
		if p.ApiVersion == "" {
			p.ApiVersion = DefaultApiVersion
		}
	}
	if vfiler := n.GetChildContentS("api_vfiler"); vfiler != "" {
		p.ApiVfiler = vfiler
	}
	if addr := n.GetChildContentS("addr"); addr != "" {
		p.Addr = addr
	}
	isKfs := n.GetChildContentS("is_kfs")
	if isKfs == "true" {
		p.IsKfs = true
	} else if isKfs == "false" {
		p.IsKfs = false
	}
	if x := n.GetChildContentS("use_insecure_tls"); x != "" {
		if insecureTLS, err := strconv.ParseBool(x); err == nil {
			// err can be ignored since conf was already validated
			p.UseInsecureTls = &insecureTLS
		}
	}
	if authStyle := n.GetChildContentS("auth_style"); authStyle != "" {
		p.AuthStyle = authStyle
	}
	if sslCert := n.GetChildContentS("ssl_cert"); sslCert != "" {
		p.SslCert = sslCert
	}
	if sslKey := n.GetChildContentS("ssl_key"); sslKey != "" {
		p.SslKey = sslKey
	}
	if caCert := n.GetChildContentS("ca_cert"); caCert != "" {
		p.CaCertPath = caCert
	}
	if username := n.GetChildContentS("username"); username != "" {
		p.Username = username
	}
	if password := n.GetChildContentS("password"); password != "" {
		p.Password = password
	}
	if credentialsFile := n.GetChildContentS("credentials_file"); credentialsFile != "" {
		p.CredentialsFile = credentialsFile
	}
	if clientTimeout := n.GetChildContentS("client_timeout"); clientTimeout != "" {
		p.ClientTimeout = clientTimeout
	} else {
		if p.ClientTimeout == "" {
			p.ClientTimeout = DefaultTimeout
		}
	}
	return p
}

type Exporter struct {
	Port              *int      `yaml:"port,omitempty"`
	PortRange         *IntRange `yaml:"port_range,omitempty"`
	Type              string    `yaml:"exporter,omitempty"`
	Addr              *string   `yaml:"addr,omitempty"`
	Url               *string   `yaml:"url,omitempty"`
	LocalHttpAddr     string    `yaml:"local_http_addr,omitempty"`
	GlobalPrefix      *string   `yaml:"global_prefix,omitempty"`
	AllowedAddrs      *[]string `yaml:"allow_addrs,omitempty"`
	AllowedAddrsRegex *[]string `yaml:"allow_addrs_regex,omitempty"`
	CacheMaxKeep      *string   `yaml:"cache_max_keep,omitempty"`
	ShouldAddMetaTags *bool     `yaml:"add_meta_tags,omitempty"`

	// Prometheus specific
	HeartBeatUrl string `yaml:"heart_beat_url,omitempty"`
	SortLabels   bool   `yaml:"sort_labels,omitempty"`

	// InfluxDB specific
	Bucket        *string `yaml:"bucket,omitempty"`
	Org           *string `yaml:"org,omitempty"`
	Token         *string `yaml:"token,omitempty"`
	Precision     *string `yaml:"precision,omitempty"`
	ClientTimeout *string `yaml:"client_timeout,omitempty"`
	Version       *string `yaml:"version,omitempty"`
}

type Pollers struct {
	namesInOrder []string
}

var defaultTemplate = &[]string{"default.yaml", "custom.yaml"}

func (c *Collector) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode && n.ShortTag() == "!!str" {
		c.Name = n.Value
		c.Templates = defaultTemplate
	} else if n.Kind == yaml.MappingNode && len(n.Content) == 2 {
		c.Name = n.Content[0].Value
		var subs []string
		c.Templates = &subs
		seq := n.Content[1]
		for _, n2 := range seq.Content {
			subs = append(subs, n2.Value)
		}
	}
	return nil
}

func (i *Pollers) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.MappingNode {
		var namesInOrder []string
		for _, n := range node.Content {
			if n.Kind == yaml.ScalarNode && n.ShortTag() == "!!str" {
				namesInOrder = append(namesInOrder, n.Value)
			}
		}
		i.namesInOrder = namesInOrder
	}
	return nil
}

type OrderedConfig struct {
	Pollers Pollers `yaml:"Pollers,omitempty"`
}

type HarvestConfig struct {
	Tools          *Tools              `yaml:"Tools,omitempty"`
	Exporters      map[string]Exporter `yaml:"Exporters,omitempty"`
	Pollers        map[string]*Poller  `yaml:"Pollers,omitempty"`
	Defaults       *Poller             `yaml:"Defaults,omitempty"`
	Admin          Admin               `yaml:"Admin,omitempty"`
	PollersOrdered []string            // poller names in same order as yaml config
}
