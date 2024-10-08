package joinrest

import (
	"github.com/netapp/harvest/v2/cmd/collectors/storagegrid/rest"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/slogx"
	"github.com/netapp/harvest/v2/pkg/util"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
	"log/slog"
	"strings"
)

type JoinRest struct {
	*plugin.AbstractPlugin
	client       *rest.Client
	translateMap map[string]join
	timesCalled  int
	resourcesMap map[string]resourceMap
}

func New(p *plugin.AbstractPlugin) plugin.Plugin {
	return &JoinRest{AbstractPlugin: p}
}

const joinTemplate = `
plugins:
  - JoinRest:
      - rest: grid/traffic-classes/policies
        metrics:
          - storagegrid_private_load_balancer_storage_request_count
          - storagegrid_private_load_balancer_storage_rx_bytes
          - storagegrid_private_load_balancer_storage_tx_bytes
          - storagegrid_private_load_balancer_storage_request_time
          - storagegrid_private_load_balancer_storage_request_body_bytes_bucket
        join_rest: id
        with_prom: policy_id
        label_rest: name
        label_prom: policy
`

func (t *JoinRest) Init() error {
	var err error

	if err := t.InitAbc(); err != nil {
		return err
	}
	if err := t.initClient(); err != nil {
		return err
	}

	t.resourcesMap = make(map[string]resourceMap)

	// Read hidden plugin
	t.translateMap = make(map[string]join)
	decoder := yaml.NewDecoder(strings.NewReader(joinTemplate))
	var tm translatePlugin
	err = decoder.Decode(&tm)
	if err != nil {
		t.SLogger.Error("Failed to decode joinTemplate", slogx.Err(err))
		return err
	}
	for _, p := range tm.Plugins {
		for _, j := range p.Translate {
			for _, metric := range j.Metrics {
				t.translateMap[metric] = j
			}
		}
	}

	// Update caches every 6m
	s := t.Params.NewChildS("schedule", "")
	s.NewChildS("data", "6m")

	// Refresh the cache after the plugin is called n times
	t.timesCalled = t.SetPluginInterval()
	return nil
}

func (t *JoinRest) Run(dataMap map[string]*matrix.Matrix) ([]*matrix.Matrix, *util.Metadata, error) {
	t.client.Metadata.Reset()

	if t.timesCalled >= t.PluginInvocationRate {
		// refresh cache
		t.timesCalled = 0

		for _, model := range t.translateMap {
			bytes, err := t.client.GetGridRest(model.Rest)
			if err != nil {
				t.SLogger.Error(
					"Failed to collect records from REST",
					slogx.Err(err),
					slog.String("rest", model.Rest),
				)
				continue
			}
			t.updateCache(model, &bytes)
		}
	}

	for metricName, model := range t.translateMap {
		m, ok := dataMap[metricName]
		if !ok {
			continue
		}
		cache, ok := t.resourcesMap[model.Rest]
		if !ok {
			t.SLogger.Warn(
				"Cache does not have resources for REST",
				slog.String("metricName", metricName),
				slog.String("rest", model.Rest),
			)
			continue
		}
		for _, instance := range m.GetInstances() {
			label := instance.GetLabel(model.WithProm)
			if label == "" {
				t.SLogger.Debug(
					"Instance label for withProm is empty. Ignoring",
					slog.String("metricName", metricName),
					slog.String("withProm", model.WithProm),
					slog.String("rest", model.Rest),
				)
				continue
			}
			newLabel, ok := cache[label]
			if !ok {
				t.SLogger.Debug(
					"Cache does not contain label. Ignoring",
					slog.String("metricName", metricName),
					slog.String("withProm", model.WithProm),
					slog.String("label", label),
					slog.String("rest", model.Rest),
				)
				continue
			}
			instance.SetLabel(model.LabelProm, newLabel)
		}
	}
	t.timesCalled++
	return nil, t.client.Metadata, nil
}

func (t *JoinRest) updateCache(model join, bytes *[]byte) {
	results := gjson.ParseBytes(*bytes)
	keys := results.Get("data.#." + model.JoinRest).Array()
	vals := results.Get("data.#." + model.LabelRest).Array()
	if len(keys) != len(vals) {
		t.SLogger.Error(
			"Data sizes are different lengths",
			slog.String("restKey", model.JoinRest),
			slog.String("restVal", model.LabelRest),
		)
		return
	}
	for i, k := range keys {
		m, ok := t.resourcesMap[model.Rest]
		if !ok {
			m = make(map[string]string)
			t.resourcesMap[model.Rest] = m
		}
		m[k.String()] = vals[i].String()
	}
}

func (t *JoinRest) initClient() error {
	var err error

	if t.client, err = rest.NewClient(t.Options.Poller, t.Params.GetChildContentS("client_timeout"), t.Auth); err != nil {
		return err
	}

	if err := t.client.Init(5); err != nil {
		return err
	}
	t.client.TraceLogSet(t.Name, t.Params)

	return nil
}

type resourceMap map[string]string

type join struct {
	Rest      string   `yaml:"rest"`
	Metrics   []string `yaml:"metrics"`
	JoinRest  string   `yaml:"join_rest"`
	WithProm  string   `yaml:"with_prom"`
	LabelRest string   `yaml:"label_rest"`
	LabelProm string   `yaml:"label_prom"`
}

type plugins struct {
	Translate []join `yaml:"JoinRest"`
}

type translatePlugin struct {
	Plugins []plugins `yaml:"plugins"`
}
