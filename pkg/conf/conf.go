/*
 * Copyright NetApp Inc, 2021 All rights reserved
 */

package conf

import (
	"errors"
	"fmt"
	"github.com/netapp/harvest/v2/pkg/errs"
	"github.com/netapp/harvest/v2/pkg/logging"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/netapp/harvest/v2/third_party/mergo"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

var (
	Config            = HarvestConfig{}
	configRead        = false
	readMu            = sync.Mutex{}
	credentialModTime = int64(0)
	credConfig        HarvestConfig
)

const (
	DefaultAPIVersion = "1.3"
	DefaultTimeout    = "30s"
	DefaultConfPath   = "conf"
	HarvestYML        = "harvest.yml"
	BasicAuth         = "basic_auth"
	CertificateAuth   = "certificate_auth"
	HomeEnvVar        = "HARVEST_CONF"
)

// TestLoadHarvestConfig loads a new config - used by testing code
func TestLoadHarvestConfig(configPath string) {
	configRead = false
	Config = HarvestConfig{}
	promPortRangeMapping = make(map[string]PortMap)
	_, err := LoadHarvestConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config at=[%s] err=%+v\n", configPath, err)
	}
}

func ConfigPath(path string) string {
	// Harvest uses the following precedence order. Each item takes precedence over the
	// item below it. All paths are relative to `HARVEST_CONF` environment variable
	// 1. `--config` command line flag
	// 2. `HARVEST_CONFIG` environment variable
	// 3. no command line argument and no environment variable, use the default path (HarvestYML)
	if path != HarvestYML && path != "./"+HarvestYML {
		return Path(path)
	}
	fp := os.Getenv("HARVEST_CONFIG")
	if fp != "" {
		path = fp
	}
	return Path(path)
}

func LoadHarvestConfig(configPath string) (string, error) {
	var (
		contents   []byte
		duplicates []error
		err        error
	)

	configPath = ConfigPath(configPath)
	if configRead {
		return configPath, nil
	}
	contents, err = os.ReadFile(configPath)

	if err != nil {
		return "", fmt.Errorf("error reading %s err=%w", configPath, err)
	}
	err = DecodeConfig(contents)
	if err != nil {
		fmt.Printf("error unmarshalling config file=[%s] %+v\n", configPath, err)
		return "", err
	}

	for _, pat := range Config.PollerFiles {
		fs, err := filepath.Glob(pat)
		if err != nil {
			return "", fmt.Errorf("error retrieving poller_files path=%s err=%w", pat, err)
		}

		sort.Strings(fs)

		if len(fs) == 0 {
			fmt.Printf("add 0 poller(s) from poller_file=%s because no matching paths\n", pat)
			continue
		}

		for _, filename := range fs {
			fsContents, err := os.ReadFile(filename)
			if err != nil {
				return "", fmt.Errorf("error reading poller_file=%s err=%w", filename, err)
			}
			cfg, err := unmarshalConfig(fsContents)
			if err != nil {
				return "", fmt.Errorf("error unmarshalling poller_file=%s err=%w", filename, err)
			}
			for _, pName := range cfg.PollersOrdered {
				_, ok := Config.Pollers[pName]
				if ok {
					duplicates = append(duplicates, fmt.Errorf("poller name=%s from poller_file=%s is not unique", pName, filename))
					continue
				}
				// Merge poller and defaults
				child := cfg.Pollers[pName]
				if Config.Defaults != nil {
					child.Union(Config.Defaults)
				}
				Config.Pollers[pName] = child
				Config.PollersOrdered = append(Config.PollersOrdered, pName)
			}
			fmt.Printf("add %d poller(s) from poller_file=%s\n", len(cfg.PollersOrdered), filename)
		}
	}

	if len(duplicates) > 0 {
		return "", errors.Join(duplicates...)
	}

	// After processing all the configuration files, check if the Config.Pollers parameter is still empty.
	if len(Config.Pollers) == 0 {
		return "", errs.New(errs.ErrConfig, "[Pollers] section not found")
	}

	// Fix promIndex for combined pollers
	for i, name := range Config.PollersOrdered {
		Config.Pollers[name].promIndex = i
	}

	fixupExporters()
	return configPath, nil
}

func fixupExporters() {
	for _, pollerName := range Config.PollersOrdered {
		poller := Config.Pollers[pollerName]
		for i, e := range poller.ExporterDefs {
			exporterName := e.name
			if exporterName == "" {
				// This is an embedded exporter, synthesize a name for it
				exporterName = fmt.Sprintf("%s-%d", pollerName, i)
				Config.Exporters[exporterName] = e.Exporter
			}

			poller.Exporters = append(poller.Exporters, exporterName)
		}
	}
}

func unmarshalConfig(contents []byte) (*HarvestConfig, error) {
	var (
		cfg           HarvestConfig
		orderedConfig OrderedConfig
		err           error
	)

	contents, err = ExpandVars(contents)
	if err != nil {
		return nil, fmt.Errorf("error expanding vars: %w", err)
	}

	err = yaml.Unmarshal(contents, &cfg)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// Read the yaml again to determine poller order
	err = yaml.Unmarshal(contents, &orderedConfig)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling ordered config: %w", err)
	}
	cfg.PollersOrdered = orderedConfig.Pollers.namesInOrder
	for i, name := range Config.PollersOrdered {
		Config.Pollers[name].promIndex = i
	}

	return &cfg, nil
}

func DecodeConfig(contents []byte) error {
	cfg, err := unmarshalConfig(contents)
	configRead = true
	if err != nil {
		return fmt.Errorf("error unmarshalling config err: %w", err)
	}
	Config = *cfg

	// Initialize Config.Pollers if it's nil
	if Config.Pollers == nil {
		Config.Pollers = make(map[string]*Poller)
	}
	// Merge pollers and defaults
	pollers := Config.Pollers
	defaults := Config.Defaults

	// Iterate through the pollers check if any are nil and if so create an empty poller
	// This happens when the poller is listed in your config file, but has no configuration
	for name, p := range pollers {
		if p == nil {
			p = &Poller{Name: name}
			pollers[name] = p
		}
	}

	if defaults != nil {
		for _, p := range pollers {
			p.Union(defaults)
		}
	}
	return nil
}

func ReadCredentialFile(credPath string, p *Poller) error {
	fileChanged, err := hasFileChanged(credPath)
	if err != nil {
		return err
	}
	if fileChanged {
		logging.Get().Info().Str("credPath", credPath).Msg("reading credentials")
		contents, err := os.ReadFile(credPath)
		if err != nil {
			abs, err2 := filepath.Abs(credPath)
			if err2 != nil {
				abs = credPath
			}
			return fmt.Errorf("failed to read file=%s error: %w", abs, err)
		}
		err = yaml.Unmarshal(contents, &credConfig)
		if err != nil {
			return err
		}
	}
	if p == nil {
		return nil
	}

	credPoller := credConfig.Pollers[p.Name]
	if credPoller == nil {
		// when the poller is not listed in the file, check if there is a default, and if so, use it
		if credConfig.Defaults != nil {
			credPoller = credConfig.Defaults
		} else {
			return errs.New(errs.ErrInvalidParam, "poller not found in credentials file")
		}
	}

	// Merge the poller and defaults from the credential file
	if credConfig.Defaults != nil {
		_ = mergo.Merge(credPoller, credConfig.Defaults)
	}

	if p.SslKey == "" {
		p.SslKey = credPoller.SslKey
	}
	if p.SslCert == "" {
		p.SslCert = credPoller.SslCert
	}
	if p.CaCertPath == "" {
		p.CaCertPath = credPoller.CaCertPath
	}
	if p.Username == "" {
		p.Username = credPoller.Username
	}
	if p.Password == "" {
		p.Password = credPoller.Password
	}
	return nil
}

func hasFileChanged(path string) (bool, error) {
	readMu.Lock()
	defer readMu.Unlock()
	stat, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("failed to stat file=%s error: %w", path, err)
	}
	if stat.ModTime().Unix() > credentialModTime {
		credentialModTime = stat.ModTime().Unix()
		return true, nil
	}
	return false, nil
}

func PollerNamed(name string) (*Poller, error) {
	poller, ok := Config.Pollers[name]
	if !ok {
		return nil, errs.New(errs.ErrConfig, "poller ["+name+"] not found")
	}
	poller.Name = name
	return poller, nil
}

// Path returns a path based on aPath and the HARVEST_CONF environment variable.
// If aPath is absolute, it is returned unchanged.
// When the HARVEST_CONF environment variable is set, a new path is returned relative to HARVEST_CONF.
// Otherwise, a new path is returned relative to the current working directory.
func Path(aPath string) string {
	confDir := os.Getenv(HomeEnvVar)
	if aPath == "" {
		return confDir
	}
	if filepath.IsAbs(aPath) {
		return aPath
	}
	if strings.HasPrefix(aPath, confDir) {
		return aPath
	}
	return filepath.Join(confDir, aPath)
}

// GetLastPromPort returns the Prometheus port for the given poller
// If multiple Prometheus exporters are configured for a poller, the port for the last exporter is returned.
func GetLastPromPort(pollerName string, validatePortInUse bool) (int, error) {
	var port int
	var isPrometheusExporterConfigured bool

	if len(promPortRangeMapping) == 0 {
		loadPrometheusExporterPortRangeMapping(validatePortInUse)
	}
	poller := Config.Pollers[pollerName]
	if poller == nil {
		return 0, errs.New(errs.ErrConfig, "Poller does not exist "+pollerName)
	}

	exporters := poller.Exporters
	for i := len(exporters) - 1; i >= 0; i-- {
		e := exporters[i]
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
	if port == 0 && isPrometheusExporterConfigured {
		return port, errs.New(errs.ErrConfig, "No free port found for poller "+pollerName)
	}
	return port, nil
}

type PortMap struct {
	portSet   []int
	freePorts map[int]struct{}
}

func PortMapFromRange(address string, portRange *IntRange, validatePortInUse bool) PortMap {
	portMap := PortMap{}
	portMap.freePorts = make(map[int]struct{})
	start := portRange.Min
	end := portRange.Max
	for i := start; i <= end; i++ {
		portMap.portSet = append(portMap.portSet, i)
		if validatePortInUse {
			portMap.freePorts[i] = struct{}{}
		}
	}
	if !validatePortInUse {
		portMap.freePorts = util.CheckFreePorts(address, portMap.portSet)
	}
	return portMap
}

var promPortRangeMapping = make(map[string]PortMap)

func loadPrometheusExporterPortRangeMapping(validatePortInUse bool) {
	for k, v := range Config.Exporters {
		if v.Type == "Prometheus" {
			if v.PortRange != nil {
				// we only care about free ports on the localhost
				promPortRangeMapping[k] = PortMapFromRange("localhost", v.PortRange, validatePortInUse)
			}
		}
	}
}

type IntRange struct {
	Min int
	Max int
}

var rangeRegex = regexp.MustCompile(`(\d+)\s*-\s*(\d+)`)

func (i *IntRange) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode && n.ShortTag() == "!!str" {
		matches := rangeRegex.FindStringSubmatch(n.Value)
		if len(matches) == 3 {
			minVal, err1 := strconv.Atoi(matches[1])
			maxVal, err2 := strconv.Atoi(matches[2])
			if err1 != nil {
				return err1
			}
			if err2 != nil {
				return err2
			}
			i.Min = minVal
			i.Max = maxVal
		}
	}
	return nil
}

// GetUniqueExporters returns the unique set of exporter types from the list of export names.
// For example, if two Prometheus exporters are configured for a poller, the last one is returned.
// Multiple InfluxDB exporters are allowed.
func GetUniqueExporters(exporterNames []string) []string {
	var resultExporters []string
	exporterMap := make(map[string][]string)

	for _, ec := range exporterNames {
		e, ok := Config.Exporters[ec]
		if ok {
			exporterMap[e.Type] = append(exporterMap[e.Type], ec)
		}
	}

	for eType, value := range exporterMap {
		if eType == "Prometheus" {
			// if there are multiple prometheus exporters, only the last one is used
			resultExporters = append(resultExporters, value[len(value)-1])
			continue
		}
		resultExporters = append(resultExporters, value...)
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
	GrafanaAPIToken string `yaml:"grafana_api_token,omitempty"`
	AsupDisabled    bool   `yaml:"autosupport_disabled,omitempty"`
}

type Collector struct {
	Name      string    `yaml:"-"`
	Templates *[]string `yaml:"-"`
}

type CredentialsScript struct {
	Path     string `yaml:"path,omitempty"`
	Schedule string `yaml:"schedule,omitempty"`
	Timeout  string `yaml:"timeout,omitempty"`
}

type CertificateScript struct {
	Path    string `yaml:"path,omitempty"`
	Timeout string `yaml:"timeout,omitempty"`
}

type ExportDef struct {
	name string
	Exporter
}

func (e *ExportDef) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.MappingNode {
		var aExporter *Exporter
		err := n.Decode(&aExporter)
		if err != nil {
			return fmt.Errorf("error unmarshalling embedded exporter: %w", err)
		}
		e.Exporter = *aExporter
	} else if n.Kind == yaml.ScalarNode && n.ShortTag() == "!!str" {
		e.name = n.Value
	}
	return nil
}

type Poller struct {
	Addr              string               `yaml:"addr,omitempty"`
	APIVersion        string               `yaml:"api_version,omitempty"`
	APIVfiler         string               `yaml:"api_vfiler,omitempty"`
	AuthStyle         string               `yaml:"auth_style,omitempty"`
	CaCertPath        string               `yaml:"ca_cert,omitempty"`
	ClientTimeout     string               `yaml:"client_timeout,omitempty"`
	Collectors        []Collector          `yaml:"collectors,omitempty"`
	CredentialsFile   string               `yaml:"credentials_file,omitempty"`
	CredentialsScript CredentialsScript    `yaml:"credentials_script,omitempty"`
	CertificateScript CertificateScript    `yaml:"certificate_script,omitempty"`
	Datacenter        string               `yaml:"datacenter,omitempty"`
	ExporterDefs      []ExportDef          `yaml:"exporters,omitempty"`
	IsKfs             bool                 `yaml:"is_kfs,omitempty"`
	Labels            *[]map[string]string `yaml:"labels,omitempty"`
	LogMaxBytes       int64                `yaml:"log_max_bytes,omitempty"`
	LogMaxFiles       int                  `yaml:"log_max_files,omitempty"`
	LogSet            *[]string            `yaml:"log,omitempty"`
	Password          string               `yaml:"password,omitempty"`
	PollerSchedule    string               `yaml:"poller_schedule,omitempty"`
	PollerLogSchedule string               `yaml:"poller_log_schedule,omitempty"`
	SslCert           string               `yaml:"ssl_cert,omitempty"`
	SslKey            string               `yaml:"ssl_key,omitempty"`
	TLSMinVersion     string               `yaml:"tls_min_version,omitempty"`
	UseInsecureTLS    *bool                `yaml:"use_insecure_tls,omitempty"`
	Username          string               `yaml:"username,omitempty"`
	PreferZAPI        bool                 `yaml:"prefer_zapi,omitempty"`
	ConfPath          string               `yaml:"conf_path,omitempty"`
	Exporters         []string             `yaml:"-"`
	promIndex         int
	Name              string
}

// Union merges a poller's config with the defaults.
// For all keys in default, copy them to the poller if the poller does not already include them
func (p *Poller) Union(defaults *Poller) {
	// this is needed because of how mergo handles boolean zero values
	isInsecureNil := true
	var pUseInsecureTLS bool
	pIsKfs := p.IsKfs
	if p.UseInsecureTLS != nil {
		isInsecureNil = false
		pUseInsecureTLS = *p.UseInsecureTLS
	}
	// Don't copy auth related fields from defaults to poller, even when the poller is missing those fields.
	// Save a copy of the poller's auth fields and restore after merge
	pPassword := p.Password
	pAuthStyle := p.AuthStyle
	pCredentialsFile := p.CredentialsFile
	pCredentialsScript := p.CredentialsScript.Path
	_ = mergo.Merge(p, defaults)
	if !isInsecureNil {
		p.UseInsecureTLS = &pUseInsecureTLS
	}
	p.IsKfs = pIsKfs
	p.Password = pPassword
	p.AuthStyle = pAuthStyle
	p.CredentialsFile = pCredentialsFile
	p.CredentialsScript.Path = pCredentialsScript
}

// ZapiPoller creates a poller out of a node, this is a bridge between the node and struct-based code
// Used by ZAPI based code
func ZapiPoller(n *node.Node) *Poller {
	var p Poller

	if Config.Defaults != nil {
		p = *Config.Defaults
	} else {
		p = Poller{}
	}
	p.Name = n.GetChildContentS("poller_name")
	if apiVersion := n.GetChildContentS("api_version"); apiVersion != "" {
		p.APIVersion = apiVersion
	} else if p.APIVersion == "" {
		p.APIVersion = DefaultAPIVersion
	}
	if vfiler := n.GetChildContentS("api_vfiler"); vfiler != "" {
		p.APIVfiler = vfiler
	}
	if addr := n.GetChildContentS("addr"); addr != "" {
		p.Addr = addr
	}
	isKfs := n.GetChildContentS("is_kfs")
	p.IsKfs = isKfs == "true"

	if x := n.GetChildContentS("use_insecure_tls"); x != "" {
		if insecureTLS, err := strconv.ParseBool(x); err == nil {
			// err can be ignored since conf was already validated
			p.UseInsecureTLS = &insecureTLS
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
	if credentialsScriptNode := n.GetChildS("credentials_script"); credentialsScriptNode != nil {
		p.CredentialsScript.Path = credentialsScriptNode.GetChildContentS("path")
		p.CredentialsScript.Schedule = credentialsScriptNode.GetChildContentS("schedule")
		p.CredentialsScript.Timeout = credentialsScriptNode.GetChildContentS("timeout")
	}
	if certificateScriptNode := n.GetChildS("certificate_script"); certificateScriptNode != nil {
		p.CertificateScript.Path = certificateScriptNode.GetChildContentS("path")
		p.CertificateScript.Timeout = certificateScriptNode.GetChildContentS("timeout")
	}
	if clientTimeout := n.GetChildContentS("client_timeout"); clientTimeout != "" {
		p.ClientTimeout = clientTimeout
	} else if p.ClientTimeout == "" {
		p.ClientTimeout = DefaultTimeout
	}
	if tlsMinVersion := n.GetChildContentS("tls_min_version"); tlsMinVersion != "" {
		p.TLSMinVersion = tlsMinVersion
	}
	if logSet := n.GetChildS("log"); logSet != nil {
		names := logSet.GetAllChildNamesS()
		p.LogSet = &names
	}
	if confPath := n.GetChildContentS("conf_path"); confPath != "" {
		p.ConfPath = confPath
	}
	return &p
}

type Exporter struct {
	Port              *int      `yaml:"port,omitempty"`
	PortRange         *IntRange `yaml:"port_range,omitempty"`
	Type              string    `yaml:"exporter,omitempty"`
	Addr              *string   `yaml:"addr,omitempty"`
	URL               *string   `yaml:"url,omitempty"`
	LocalHTTPAddr     string    `yaml:"local_http_addr,omitempty"`
	GlobalPrefix      *string   `yaml:"global_prefix,omitempty"`
	AllowedAddrs      *[]string `yaml:"allow_addrs,omitempty"`
	AllowedAddrsRegex *[]string `yaml:"allow_addrs_regex,omitempty"`
	CacheMaxKeep      *string   `yaml:"cache_max_keep,omitempty"`
	ShouldAddMetaTags *bool     `yaml:"add_meta_tags,omitempty"`

	// Prometheus specific
	HeartBeatURL string `yaml:"heart_beat_url,omitempty"`
	SortLabels   bool   `yaml:"sort_labels,omitempty"`
	TLS          TLS    `yaml:"tls,omitempty"`

	// InfluxDB specific
	Bucket        *string `yaml:"bucket,omitempty"`
	Org           *string `yaml:"org,omitempty"`
	Token         *string `yaml:"token,omitempty"`
	Precision     *string `yaml:"precision,omitempty"`
	ClientTimeout *string `yaml:"client_timeout,omitempty"`
	Version       *string `yaml:"version,omitempty"`

	IsTest bool // true when run from unit tests
}

type Pollers struct {
	namesInOrder []string
}

var defaultTemplate = &[]string{"default.yaml", "custom.yaml"}

func NewCollector(name string) Collector {
	return Collector{
		Name:      name,
		Templates: defaultTemplate,
	}
}

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

func (i *Pollers) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.MappingNode {
		var namesInOrder []string
		for _, n := range n.Content {
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
	PollerFiles    []string            `yaml:"Poller_files,omitempty"`
	Defaults       *Poller             `yaml:"Defaults,omitempty"`
	Admin          Admin               `yaml:"Admin,omitempty"`
	PollersOrdered []string            // poller names in same order as yaml config
}
