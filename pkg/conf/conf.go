/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package conf

import (
	"fmt"
	"github.com/imdario/mergo"
	"goharvest2/pkg/constant"
	"goharvest2/pkg/errors"
	"goharvest2/pkg/tree"
	"goharvest2/pkg/tree/node"
	"goharvest2/pkg/util"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
)

// LoadConfig loads the config info from harvest.yml
func LoadConfig(configPath string) (*node.Node, error) {
	configNode, err := tree.Import("yaml", configPath)
	if configNode != nil {
		// Load HarvestConfig to rewrite passwords - eventually all the code will be refactored to use HarvestConfig.
		// This is needed because the current yaml parser does not handle password with special characters.
		// E.g abc#123, that's because the # is interpreted as the beginning of a comment. The code below overwrites
		// the incorrect password with the correct one by using a better yaml parser for each Poller and Default section
		err := LoadHarvestConfig(configPath)
		if err != nil {
			return nil, err
		}
		pollers := configNode.GetChildS("Pollers")
		if pollers != nil {
			for _, poller := range pollers.GetChildren() {
				password := poller.GetChildContentS("password")
				pollerStruct := (*Config.Pollers)[poller.GetNameS()]
				if pollerStruct.Password != "" && pollerStruct.Password != password {
					poller.SetChildContentS("password", pollerStruct.Password)
				}
			}
		}
		// Check Defaults also
		defaultNode := configNode.GetChildS("Defaults")
		if defaultNode != nil {
			password := defaultNode.GetChildContentS("password")
			defaultStruct := *Config.Defaults
			if defaultStruct.Password != "" && defaultStruct.Password != password {
				defaultNode.SetChildContentS("password", defaultStruct.Password)
			}
		}
	}
	return configNode, err
}

var Config = HarvestConfig{}
var configRead = false
var ValidatePortInUse = false

func LoadHarvestConfig(configPath string) error {
	if configRead {
		return nil
	}
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
	return nil
}

func SafeConfig(n *node.Node, fp string) error {
	return tree.Export(n, "yaml", fp)
}

func GetExporters2(configFp string) (map[string]Exporter, error) {
	err := LoadHarvestConfig(configFp)
	if err != nil {
		return nil, err
	}
	exporters := Config.Exporters

	if exporters == nil {
		err = errors.New(errors.ERR_CONFIG, "[Exporters] section not found")
		return nil, err
	}

	return *exporters, nil
}

func GetExporters(configFp string) (*node.Node, error) {
	var err error
	var config, exporters *node.Node

	if config, err = LoadConfig(configFp); err != nil {
		return nil, err
	}

	if exporters = config.GetChildS("Exporters"); exporters == nil {
		err = errors.New(errors.ERR_CONFIG, "[Exporters] section not found")
		return nil, err
	}

	return exporters, nil
}

func GetPollerNames(configFp string) ([]string, error) {

	var pollerNames []string
	var config, pollers *node.Node
	var err error

	if config, err = LoadConfig(configFp); err != nil {
		return pollerNames, err
	}

	if pollers = config.GetChildS("Pollers"); pollers == nil {
		return pollerNames, errors.New(errors.ERR_CONFIG, "[Pollers] section not found")
	}

	pollerNames = make([]string, 0)

	for _, p := range pollers.GetChildren() {
		pollerNames = append(pollerNames, p.GetNameS())
	}

	return pollerNames, nil
}

func GetPollers2(configFp string) (map[string]*Poller, error) {
	err := LoadHarvestConfig(configFp)
	if err != nil {
		return nil, err
	}
	pollers := Config.Pollers
	defaults := Config.Defaults

	if pollers == nil {
		return nil, errors.New(errors.ERR_CONFIG, "[Pollers] section not found")
	} else if defaults != nil { // optional
		for _, p := range *pollers {
			p.Union(defaults)
		}
	}
	return *pollers, nil
}

func GetPollers(configFp string) (*node.Node, error) {
	var config, pollers, defaults *node.Node
	var err error

	if config, err = LoadConfig(configFp); err != nil {
		return nil, err
	}

	pollers = config.GetChildS("Pollers")
	defaults = config.GetChildS("Defaults")

	if pollers == nil {
		err = errors.New(errors.ERR_CONFIG, "[Pollers] section not found")
	} else if defaults != nil { // optional
		for _, p := range pollers.GetChildren() {
			p.Union(defaults)
		}
	}
	return pollers, err
}

func GetPoller2(configFp, pollerName string) (*Poller, error) {
	pollers, err := GetPollers2(configFp)
	if err != nil {
		return nil, err
	}
	poller, ok := pollers[pollerName]
	if !ok {
		return nil, errors.New(errors.ERR_CONFIG, "poller ["+pollerName+"] not found")
	}
	return poller, nil
}

func GetPoller(configFp, pollerName string) (*node.Node, error) {
	var err error
	var pollers, poller *node.Node

	if pollers, err = GetPollers(configFp); err == nil {
		if poller = pollers.GetChildS(pollerName); poller == nil {
			err = errors.New(errors.ERR_CONFIG, "poller ["+pollerName+"] not found")
		}
	}
	return poller, err
}

/*GetDefaultHarvestConfigPath*/
//This method is used to return the default absolute path of harvest config file.
func GetDefaultHarvestConfigPath() (string, error) {
	var configPath string
	var err error
	configFileName := constant.ConfigFileName
	if configPath = os.Getenv("HARVEST_CONF"); configPath == "" {
		var homePath string
		homePath = GetHarvestHomePath()
		configPath = path.Join(homePath, configFileName)
	} else {
		configPath = path.Join(configPath, configFileName)
	}
	return configPath, err
}

/*GetHarvestHomePath*/
//This method is used to return current working directory
func GetHarvestHomePath() string {
	return "./"
}

func GetHarvestLogPath() string {
	var logPath string
	if logPath = os.Getenv("HARVEST_LOGS"); logPath == "" {
		logPath = "/var/log/harvest/"
	}
	return logPath
}

/*
GetPrometheusExporterPorts returns port configured in prometheus exporter for given poller
*/
func GetPrometheusExporterPorts(pollerName string) (int, error) {
	var port int
	var isPrometheusExporterConfigured bool

	if len(promPortRangeMapping) == 0 {
		loadPrometheusExporterPortRangeMapping()
	}
	poller := (*Config.Pollers)[pollerName]
	if poller == nil {
		return 0, errors.New(errors.ERR_CONFIG, "Poller does not exist "+pollerName)
	}
	exporters := poller.Exporters
	if exporters != nil && len(*exporters) > 0 {
		for _, e := range *exporters {
			exporter := (*Config.Exporters)[e]
			if exporter.Type != nil && *exporter.Type == "Prometheus" {
				isPrometheusExporterConfigured = true
				if exporter.PortRange != nil {
					ports := promPortRangeMapping[e]
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
			continue
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
	exporters := *Config.Exporters
	for k, v := range exporters {
		if *v.Type == "Prometheus" {
			if v.PortRange != nil {
				promPortRangeMapping[k] = PortMapFromRange(*v.Addr, v.PortRange)
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

// GetUniqueExporters returns unique type of exporters for the poller
// For example: If 2 prometheus exporters are configured for a poller then last one defined is returned
func GetUniqueExporters(p *node.Node, configFp string) ([]string, error) {
	var resultExporters []string
	exporters := p.GetChildS("exporters")
	if exporters != nil {
		exportChildren := exporters.GetAllChildContentS()
		definedExporters, err := GetExporters(configFp)
		if err != nil {
			return nil, err
		}
		exporterMap := make(map[string]string)
		for _, ec := range exportChildren {
			e := definedExporters.GetChildS(ec)
			if e != nil {
				exporterType := e.GetChildContentS("exporter")
				exporterMap[exporterType] = ec
			}
		}

		for _, value := range exporterMap {
			resultExporters = append(resultExporters, value)
		}
	}
	return resultExporters, nil
}

// Pointers need to be used for struct members where you need
// to distinguish missing values from zero values. See
// https://github.com/go-yaml/yaml/issues/113 for details
// The downside of making all members pointers is accessing
// the values requires more dereferencing - see doctor_test.go

type Consul struct {
	Host        *string   `yaml:"host,omitempty"`
	ServiceName *string   `yaml:"service_name,omitempty"`
	Tags        *[]string `yaml:"tags,omitempty"`
}

type Tools struct {
	GrafanaApiToken *string `yaml:"grafana_api_token,omitempty"`
}

type Poller struct {
	Datacenter     *string   `yaml:"datacenter,omitempty"`
	Addr           *string   `yaml:"addr,omitempty"`
	AuthStyle      *string   `yaml:"auth_style,omitempty"`
	Username       *string   `yaml:"username,omitempty"`
	Password       string    `yaml:"password,omitempty"`
	UseInsecureTls *bool     `yaml:"use_insecure_tls,omitempty"`
	SslCert        *string   `yaml:"ssl_cert,omitempty"`
	SslKey         *string   `yaml:"ssl_key,omitempty"`
	LogMaxBytes    *int64    `yaml:"log_max_bytes,omitempty"`
	LogMaxFiles    *int      `yaml:"log_max_files,omitempty"`
	Exporters      *[]string `yaml:"exporters,omitempty"`
	Collectors     *[]string `yaml:"collectors,omitempty"`
	IsKfs          *bool     `yaml:"is_kfs,omitempty"`
	PollerSchedule *string   `yaml:"poller_schedule,omitempty"`
}

func (p *Poller) Union(defaults *Poller) {
	_ = mergo.Merge(p, defaults)
}

type Exporter struct {
	Port              *int      `yaml:"port,omitempty"`
	PortRange         *IntRange `yaml:"port_range,omitempty"`
	Type              *string   `yaml:"exporter,omitempty"`
	Addr              *string   `yaml:"addr,omitempty"`
	Url               *string   `yaml:"url,omitempty"`
	LocalHttpAddr     *string   `yaml:"local_http_addr,omitempty"`
	GlobalPrefix      *string   `yaml:"global_prefix,omitempty"`
	AllowedAddrs      *[]string `yaml:"allow_addrs,omitempty"`
	AllowedAddrsRegex *[]string `yaml:"allow_addrs_regex,omitempty"`
	CacheMaxKeep      *string   `yaml:"cache_max_keep,omitempty"`
	ShouldAddMetaTags *bool     `yaml:"add_meta_tags,omitempty"`
	Consul            *Consul   `yaml:"consul,omitempty"`

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
	Tools          *Tools               `yaml:"Tools,omitempty"`
	Exporters      *map[string]Exporter `yaml:"Exporters,omitempty"`
	Pollers        *map[string]*Poller  `yaml:"Pollers,omitempty"`
	Defaults       *Poller              `yaml:"Defaults,omitempty"`
	PollersOrdered []string             // poller names in same order as yaml config
}
