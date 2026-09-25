package netroute

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"testing"
)

func newNetRoute(t *testing.T) *NetRoute {
	t.Helper()
	params := node.NewS("NetRoute")
	parentParams := node.NewS("parent")
	parentParams.NewChildS("export_options", "")
	n := &NetRoute{AbstractPlugin: plugin.New("Rest", options.New(), params, parentParams, "net_route", nil)}
	n.SLogger = slog.Default()
	assert.Nil(t, n.Init(conf.Remote{}))
	return n
}

func newRouteData() *matrix.Matrix {
	data := matrix.New("net_route", "net_route", "net_route")
	data.SetGlobalLabel("cluster", "cluster1")
	for i := range 50 {
		uuid := "route-" + strconv.Itoa(i)
		instance, _ := data.NewInstance(uuid)
		instance.SetLabel("uuid", uuid)
		instance.SetLabel("svm", "svm1")
		instance.SetLabel("interfaces", `{"name":["lif1","lif2"],"address":["10.0.0.1","10.0.0.2"]}`)
	}
	return data
}

func runKeys(t *testing.T, n *NetRoute, data *matrix.Matrix) map[string]string {
	t.Helper()
	output, _, err := n.Run(map[string]*matrix.Matrix{data.Object: data})
	assert.Nil(t, err)
	assert.Equal(t, len(output), 1)

	keys := make(map[string]string)
	for key, instance := range output[0].GetInstances() {
		_, hasIndex := instance.GetLabels()["index"]
		assert.False(t, hasIndex)
		keys[key] = instance.GetLabel("route_uuid") + "/" + instance.GetLabel("name")
	}
	return keys
}

func TestRunKeysAreDeterministic(t *testing.T) {
	data := newRouteData()

	first := runKeys(t, newNetRoute(t), data)
	assert.Equal(t, len(first), 100)

	// Run several more times, with fresh and reused plugins, to exercise different map iteration orders.
	reused := newNetRoute(t)
	for range 10 {
		for _, n := range []*NetRoute{newNetRoute(t), reused} {
			got := runKeys(t, n, data)
			assert.Equal(t, slices.Sorted(maps.Keys(got)), slices.Sorted(maps.Keys(first)))
			for key, want := range first {
				assert.Equal(t, got[key], want)
			}
		}
	}

	assert.Equal(t, first["route-7_lif2"], "route-7/lif2")
}

func TestExportOptions(t *testing.T) {
	n := newNetRoute(t)
	exportOptions := n.data.GetExportOptions()
	assert.Equal(t, exportOptions.GetChildS("instance_keys").GetAllChildContentS(), []string{"route_uuid"})
	assert.Equal(t, exportOptions.GetChildS("instance_labels").GetAllChildContentS(), instanceLabels)
}
