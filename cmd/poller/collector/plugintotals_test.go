package collector

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/collector"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

// newTestMetadata makes a metadata matrix. The matrix is the same as the matrix from
// AbstractCollector.InitCache. It has one instance for the "data" task.
func newTestMetadata(t *testing.T) (*matrix.Matrix, *matrix.Instance) {
	t.Helper()

	md := matrix.New("test", "metadata", "metadata")
	for _, name := range []string{"bytesRx", "numCalls", "pluginInstances"} {
		if _, err := md.NewMetricUint64(name); err != nil {
			t.Fatalf("NewMetricUint64(%s): %v", name, err)
		}
	}

	inst, err := md.NewInstance("data")
	if err != nil {
		t.Fatalf("NewInstance(data): %v", err)
	}
	return md, inst
}

// pluginMetadata makes the metadata that a plugin returns from Run.
func pluginMetadata(bytesRx, numCalls, instances uint64) *collector.Metadata {
	md := &collector.Metadata{}
	md.BytesRx.Store(bytesRx)
	md.NumCalls.Store(numCalls)
	md.PluginInstances.Store(instances)
	return md
}

func value(t *testing.T, md *matrix.Matrix, key string, i *matrix.Instance) uint64 {
	t.Helper()
	v, has := md.MustGetMetric(key).GetValueUint64(i)
	assert.True(t, has)
	return v
}

// The collector adds the instance totals from all the plugins for an object.
// Before pluginTotals, the collector called MustSetValueUint64 for each plugin.
// The collector then kept only the total from the last plugin.
func TestPluginTotalsSumsInstancesAcrossPlugins(t *testing.T) {
	md, inst := newTestMetadata(t)

	var totals pluginTotals
	totals.add(pluginMetadata(0, 0, 30))
	totals.add(pluginMetadata(0, 0, 12))
	totals.write(md, inst)

	assert.Equal(t, value(t, md, "pluginInstances", inst), uint64(42))
}

// A plugin can return metadata with no instance total. This plugin must not remove the
// totals from the plugins before it. The RestPerf Disk object shows this condition.
// Its template has these plugins: LabelAgent, Aggregator, Max and Disk.
// The Disk plugin returned &RequestMetadata, but PluginInstances was zero.
func TestPluginTotalsZeroReporterDoesNotClobberPeers(t *testing.T) {
	md, inst := newTestMetadata(t)

	var totals pluginTotals
	totals.add(pluginMetadata(0, 0, 30)) // Aggregator
	totals.add(pluginMetadata(0, 0, 12)) // Max
	totals.add(pluginMetadata(0, 0, 0))  // Disk
	totals.write(md, inst)

	assert.Equal(t, value(t, md, "pluginInstances", inst), uint64(42))
}

// The pluginInstances total must not increase after each poll.
// ResetInstance clears the record flag, but it keeps the value.
// Metric.AddValue* adds to this value and does not read the flag.
// If the code adds this metric, the total is two times too large after each poll.
func TestPluginTotalsDoesNotCompoundAcrossPolls(t *testing.T) {
	md, inst := newTestMetadata(t)

	for poll := range 3 {
		md.ResetInstance("data")

		var totals pluginTotals
		totals.add(pluginMetadata(0, 0, 30))
		totals.add(pluginMetadata(0, 0, 12))
		totals.write(md, inst)

		if got := value(t, md, "pluginInstances", inst); got != 42 {
			t.Errorf("poll %d: pluginInstances = %d, want 42", poll, got)
		}
	}
}

// write adds the bytesRx and numCalls totals to the numbers from PollData.
// PollData counts the API calls of the collector.
func TestPluginTotalsAddsToCollectorOwnCounts(t *testing.T) {
	md, inst := newTestMetadata(t)

	// These are the numbers from PollData for the data poll of the collector.
	md.MustSetValueUint64("bytesRx", inst, 1000)
	md.MustSetValueUint64("numCalls", inst, 4)

	var totals pluginTotals
	totals.add(pluginMetadata(250, 2, 0))
	totals.add(pluginMetadata(750, 3, 0))
	totals.write(md, inst)

	assert.Equal(t, value(t, md, "bytesRx", inst), uint64(2000))
	assert.Equal(t, value(t, md, "numCalls", inst), uint64(9))
}

// Many plugins give no metadata. A nil value must not cause a panic.
func TestPluginTotalsIgnoresNilMetadata(t *testing.T) {
	md, inst := newTestMetadata(t)

	var totals pluginTotals
	totals.add(nil)
	totals.add(pluginMetadata(100, 1, 7))
	totals.add(nil)
	totals.write(md, inst)

	assert.Equal(t, value(t, md, "bytesRx", inst), uint64(100))
	assert.Equal(t, value(t, md, "numCalls", inst), uint64(1))
	assert.Equal(t, value(t, md, "pluginInstances", inst), uint64(7))
}

// If no plugin gives a total, write must put zero in the metric.
// The metric must not keep the value from the last poll.
func TestPluginTotalsWritesZeroWhenNoPluginReports(t *testing.T) {
	md, inst := newTestMetadata(t)

	md.MustSetValueUint64("pluginInstances", inst, 99)
	md.ResetInstance("data")

	var totals pluginTotals
	totals.write(md, inst)

	assert.Equal(t, value(t, md, "pluginInstances", inst), uint64(0))
}
