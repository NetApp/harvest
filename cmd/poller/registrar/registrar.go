package registrar

import (
	"fmt"
	"goharvest2/cmd/poller/collector"
	"goharvest2/cmd/poller/exporter"
	"goharvest2/cmd/poller/plugin"
	"sync"
)

var (
	collectors   = make(map[string]func() collector.Collector)
	exporters    = make(map[string]func() exporter.Exporter)
	plugins      = make(map[string]func() plugin.Plugin)
	collectorsMu sync.RWMutex
	exportersMu  sync.RWMutex
	pluginsMu    sync.RWMutex
)

// RegisterCollector registers collector with name "name" and function
// "foo" which should return a new instance of the collector
func RegisterCollector(name string, foo func() collector.Collector) {

	if foo == nil {
		panic(fmt.Sprintf("can't register collector [%s] with empty nil function", name))
	}

	collectorsMu.Lock()
	defer collectorsMu.Unlock()

	if _, ok := collectors[name]; ok {
		panic(fmt.Sprintf("collector [%s] already registered", name))
	}

	collectors[name] = foo
}

// RegisterExporter registers exporter with name "name" and function
// "foo" which should return a new instance of the exporter
func RegisterExporter(name string, foo func() exporter.Exporter) {
	if foo == nil {
		panic(fmt.Sprintf("can't register exporter [%s] with empty nil function", name))
	}
	exportersMu.Lock()
	defer exportersMu.Unlock()
	if _, ok := exporters[name]; ok {
		panic(fmt.Sprintf("exporter [%s] already registered", name))
	}
	exporters[name] = foo
}

// RegisterPlugin registers plugin with name "name" and function
// "foo" which should return a new instance of the plugin
func RegisterPlugin(name string, foo func() plugin.Plugin) {
	if foo == nil {
		panic(fmt.Sprintf("can't register plugin [%s] with empty nil function", name))
	}
	pluginsMu.Lock()
	defer pluginsMu.Unlock()
	if _, ok := plugins[name]; ok {
		panic(fmt.Sprintf("plugin [%s] already registered", name))
	}
    plugins[name] = foo
}

// GetCollector returns the new() function of the collector if it
// is registered, otherwise nil
func GetCollector(name string) func() collector.Collector {
	collectorsMu.RLock()
	defer collectorsMu.RUnlock()
	return collectors[name]
}

// GetExporter returns the new() function of the exporter if it
// is registered, otherwise nil
func GetExporter(name string) func() exporter.Exporter {
	exportersMu.RLock()
	defer exportersMu.RUnlock()
	return exporters[name]
}

// GetPlugin returns the new() function of the plugin if it
// is registered, otherwise nil
func GetPlugin(name string) func() plugin.Plugin {
	pluginsMu.RLock()
	defer pluginsMu.RUnlock()
	return plugins[name]
}
