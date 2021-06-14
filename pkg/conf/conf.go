/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */
package main


//package conf

import (
	"fmt"
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
		_ = LoadHarvestConfig(configPath)
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

func LoadHarvestConfig(configPath string) error {
	contents, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Printf("error reading config file=[%s] %+v\n", configPath, err)
		return err
	}
	err = yaml.Unmarshal(contents, &Config)
	if err != nil {
		fmt.Printf("error unmarshalling config file=[%s] %+v\n", configPath, err)
		return err
	}
	return nil
}

func SafeConfig(n *node.Node, fp string) error {
	return tree.Export(n, "yaml", fp)
}

func GetExporters(config_fp string) (*node.Node, error) {
	var err error
	var config, exporters *node.Node

	if config, err = LoadConfig(config_fp); err != nil {
		return nil, err
	}

	if exporters = config.GetChildS("Exporters"); exporters == nil {
		err = errors.New(errors.ERR_CONFIG, "[Exporters] section not found")
		return nil, err
	}

	return exporters, nil
}

func GetPollerNames(config_fp string) ([]string, error) {

	var poller_names []string
	var config, pollers *node.Node
	var err error

	if config, err = LoadConfig(config_fp); err != nil {
		return poller_names, err
	}

	if pollers = config.GetChildS("Pollers"); pollers == nil {
		return poller_names, errors.New(errors.ERR_CONFIG, "[Pollers] section not found")
	}

	poller_names = make([]string, 0)

	for _, p := range pollers.GetChildren() {
		poller_names = append(poller_names, p.GetNameS())
	}

	return poller_names, nil
}

func GetPollers(config_fp string) (*node.Node, error) {
	var config, pollers, defaults *node.Node
	var err error

	if config, err = LoadConfig(config_fp); err != nil {
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

func GetPoller(config_fp, poller_name string) (*node.Node, error) {
	var err error
	var pollers, poller *node.Node

	if pollers, err = GetPollers(config_fp); err == nil {
		if poller = pollers.GetChildS(poller_name); poller == nil {
			err = errors.New(errors.ERR_CONFIG, "poller ["+poller_name+"] not found")
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

func GetHarvestPidPath() string {
	var pidPath string
	if pidPath = os.Getenv("HARVEST_PIDS"); pidPath == "" {
		pidPath = "/var/run/harvest/"
	}
	return pidPath
}

func main() {
	configFp := "/home/rahulg2/code/github/harvest/harvest.yml"
	if Config == (HarvestConfig{}) {
		LoadHarvestConfig(configFp)
	}

	for k,_ := range *Config.Pollers {
		GetPrometheusExporterPorts(k, configFp)
	}
}

/*
This method returns port configured in prometheus exporter for given poller
If there are more than 1 exporter configured for a poller then return string will have ports as comma seperated
*/
func GetPrometheusExporterPorts(pollerName string, configFp string) (int, error) {
	var port int

	if Config == (HarvestConfig{}) {
		LoadHarvestConfig(configFp)
	}
	if len(promPortRangeMapping) == 0 {
		LoadPrometheusExporterPortRangeMapping(configFp)
	}
	exporters := (*Config.Pollers)[pollerName].Exporters

	if exporters != nil && len(*exporters) > 0 {
		for _, e := range *exporters {
			exporter := (*Config.Exporters)[e]
			if *exporter.Type == "Prometheus" {
				if exporter.PortRange != nil {
					//fmt.Println(exporter.PortRange)
					ports := promPortRangeMapping[e]
					for p, _ := range ports {
						if util.CheckPortAvailable(*exporter.Addr, strconv.Itoa(p)) {
							port = p
							break
						}
					}
					for k, _ := range ports {
						if k == port {
							delete(ports,k)
							break
						}
					}
					fmt.Printf("chosen port %d \n", port)
					return port, nil
				} else if *exporter.Port != 0 {
					fmt.Printf("port---- %d \n", *exporter.Port)
					port = *exporter.Port
					return port, nil
				}
			}
			continue
		}

	}
	return port, errors.New(errors.ERR_CONFIG, "No free port found for poller "+pollerName)
}

var promPortRangeMapping = make(map[string]map[int]struct{})

func LoadPrometheusExporterPortRangeMapping(configFp string) {
	if Config == (HarvestConfig{}) {
		LoadHarvestConfig(configFp)
	}
	exporters := *Config.Exporters
	for k, v := range exporters {
		if *v.Type == "Prometheus" {
			if v.PortRange != nil {
				portRange := v.PortRange // [2000-2030]
				var ports = make(map[int]struct{})
				r := regexp.MustCompile(`\((\d+)-(\d+)\)`)
				matches := r.FindStringSubmatch(*portRange)
				//fmt.Println(matches)
				if len(matches) > 0 {
					start, _ := strconv.Atoi(matches[1])
					end, _ := strconv.Atoi(matches[2])
					for i := start; i <= end; i++ {
						ports[i] = struct{}{}
					}
				}
				promPortRangeMapping[k] = ports
			}
		}
	}
}

// Returns unique type of exporters for the poller
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
			exporterType := definedExporters.GetChildS(ec).GetChildContentS("exporter")
			exporterMap[exporterType] = ec
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
}

type Exporter struct {
	Port              *int      `yaml:"port,omitempty"`
	PortRange         *string   `yaml:"port_range,omitempty"`
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
}

type HarvestConfig struct {
	Tools     *Tools               `yaml:"Tools,omitempty"`
	Exporters *map[string]Exporter `yaml:"Exporters,omitempty"`
	Pollers   *map[string]Poller   `yaml:"Pollers,omitempty"`
	Defaults  *Poller              `yaml:"Defaults,omitempty"`
}
