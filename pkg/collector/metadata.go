package collector

import "sync/atomic"

type Metadata struct {
	BytesRx         atomic.Uint64
	NumCalls        atomic.Uint64
	PluginInstances atomic.Uint64
}

func (m *Metadata) Reset() {
	m.BytesRx.Store(0)
	m.NumCalls.Store(0)
	m.PluginInstances.Store(0)
}
