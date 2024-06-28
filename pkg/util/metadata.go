package util

type Metadata struct {
	BytesRx       uint64
	NumCalls      uint64
	PluginObjects uint64
	PluginMetrics uint64
	PluginAPID    uint64
	PluginParseD  uint64
}

func (m *Metadata) Reset() {
	m.BytesRx = 0
	m.NumCalls = 0
	m.PluginObjects = 0
	m.PluginMetrics = 0
	m.PluginAPID = 0
	m.PluginParseD = 0
}
