package grid

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

const pollerName = "test"

// newGrid builds the plugin and runs Init for real. Init resolves the poller's
// address from the loaded harvest config, so a config has to be in place first.
//
// The config is loaded through conf.TestLoadHarvestConfig, matching the sibling
// storagegrid collector test. Note this populates the process-wide conf.Config;
// that is the established pattern in this repo and is safe while no test calls
// t.Parallel().
func newGrid(t *testing.T, uuid string) *Grid {
	t.Helper()

	conf.TestLoadHarvestConfig("testdata/config.yml")

	opts := options.New()
	opts.Poller = pollerName
	params := node.NewS("Grid")
	p := plugin.New("StorageGrid", opts, params, nil, "grid", nil)

	g := &Grid{AbstractPlugin: p}
	if err := g.Init(conf.Remote{}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	g.SetRemote(conf.Remote{UUID: uuid})
	return g
}

// TestRunSetsAddrAndSystemID covers GitHub issue #4374
// "[StorageGRID] Add a new metric with `addr` value". The grid.yaml template
// declares addr and system_id as instance_labels and notes that the Grid plugin
// sets them; the feature shipped with no test, so nothing caught a regression
// in either label.
func TestRunSetsAddrAndSystemID(t *testing.T) {
	g := newGrid(t, "grid-uuid-1")

	data := matrix.New("Grid", "grid", "grid")
	inst, err := data.NewInstance("grid-1")
	assert.Nil(t, err)

	got, meta, err := g.Run(map[string]*matrix.Matrix{"grid": data})

	assert.Nil(t, err)
	assert.Nil(t, got)
	assert.Nil(t, meta)
	assert.Equal(t, inst.GetLabel("addr"), "10.1.2.3")
	assert.Equal(t, inst.GetLabel("system_id"), "grid-uuid-1")
}

// TestRunLabelsEveryInstance pins that the labels go on all instances, not just
// the first. storagegrid_labels is a per-instance metric, so a partial pass
// would leave some series without the addr the dashboards join on.
func TestRunLabelsEveryInstance(t *testing.T) {
	g := newGrid(t, "grid-uuid-1")

	data := matrix.New("Grid", "grid", "grid")
	keys := []string{"grid-1", "grid-2", "grid-3"}
	instances := make([]*matrix.Instance, 0, len(keys))
	for _, key := range keys {
		inst, err := data.NewInstance(key)
		assert.Nil(t, err)
		instances = append(instances, inst)
	}

	_, _, err := g.Run(map[string]*matrix.Matrix{"grid": data})
	assert.Nil(t, err)

	for _, inst := range instances {
		assert.Equal(t, inst.GetLabel("addr"), "10.1.2.3")
		assert.Equal(t, inst.GetLabel("system_id"), "grid-uuid-1")
	}
}

// TestRunWithEmptyRemoteUUID pins the pre-discovery case: before the client has
// resolved the grid's systemId, system_id is empty rather than stale or absent.
func TestRunWithEmptyRemoteUUID(t *testing.T) {
	g := newGrid(t, "")

	data := matrix.New("Grid", "grid", "grid")
	inst, err := data.NewInstance("grid-1")
	assert.Nil(t, err)

	_, _, err = g.Run(map[string]*matrix.Matrix{"grid": data})
	assert.Nil(t, err)

	assert.Equal(t, inst.GetLabel("addr"), "10.1.2.3")
	assert.Equal(t, inst.GetLabel("system_id"), "")
}

// TestRunNoInstances pins that an empty matrix is a no-op, not an error. The
// grid template can legitimately return nothing on an unreachable grid.
func TestRunNoInstances(t *testing.T) {
	g := newGrid(t, "grid-uuid-1")

	data := matrix.New("Grid", "grid", "grid")

	got, _, err := g.Run(map[string]*matrix.Matrix{"grid": data})

	assert.Nil(t, err)
	assert.Nil(t, got)
}
