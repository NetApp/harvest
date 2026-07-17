package eseriesperf

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/collectors/eseries"
	"github.com/netapp/harvest/v2/cmd/poller/collector"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/set"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
)

func newEseriesPerfSymbol() *EseriesPerf {
	const object = "SsdCache"
	const path = "ssd_cache.yaml"
	opts := options.New(options.WithConfPath("../../../conf"))
	opts.Poller = pollerName
	opts.HomePath = "testdata"
	opts.IsTest = true
	ac := collector.New("EseriesPerf", object, opts, params(object, path), nil, conf.Remote{Version: "11.80.0"})
	ep := &EseriesPerf{}
	if err := ep.Init(ac); err != nil {
		panic(err)
	}
	return ep
}

func symbolPollData(path string) []gjson.Result {
	return jsonToArrayPerfData(path)
}

func TestSymbolSsdCache_ParseWrappedShape(t *testing.T) {
	ep := newEseriesPerfSymbol()
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	results := symbolPollData("testdata/symbol_ssd_cache1.json")
	count, partials := ep.pollData(mat, results, set.New())

	assert.True(t, count > 0)
	assert.Equal(t, len(mat.GetInstances()), 2) // one per controller
	assert.Equal(t, partials, uint64(0))

	// Instance keyed by synthetic controller ID "a"
	instA := mat.GetInstance("a")
	if instA == nil {
		t.Fatal("expected instance keyed by \"a\"")
	}

	// statistics.reads for controller a
	readsMetric := mat.GetMetric("statistics.reads")
	if readsMetric == nil {
		t.Fatal("statistics.reads metric not found")
	}
	val, ok := readsMetric.GetValueFloat64(instA)
	assert.True(t, ok)
	assert.Equal(t, val, 1129310.0)

	// statistics.availableBytes (raw type) — should be present on both instances
	avail := mat.GetMetric("statistics.availableBytes")
	if avail == nil {
		t.Fatal("statistics.availableBytes metric not found")
	}
	availVal, ok := avail.GetValueFloat64(instA)
	assert.True(t, ok)
	assert.Equal(t, availVal, 1599784091648.0)

	instB := mat.GetInstance("b")
	if instB == nil {
		t.Fatal("expected instance keyed by \"b\"")
	}
	valB, okB := readsMetric.GetValueFloat64(instB)
	assert.True(t, okB)
	assert.Equal(t, valB, 2000000.0)
}

func TestSymbolSsdCache_CacheAbsentFirstPoll(t *testing.T) {
	ep := newEseriesPerfSymbol()
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	count, partials := ep.pollData(mat, []gjson.Result{}, set.New())

	assert.Equal(t, count, uint64(0))
	assert.Equal(t, partials, uint64(0))
	assert.Equal(t, len(mat.GetInstances()), 0)
}

func TestSymbolSsdCache_CacheCreatedOverPoll(t *testing.T) {
	ep := newEseriesPerfSymbol()
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	// Poll 1: no cache
	count1, _ := ep.pollData(mat, []gjson.Result{}, set.New())
	assert.Equal(t, count1, uint64(0))
	assert.Equal(t, len(mat.GetInstances()), 0)

	// Poll 2: cache appears
	count2, _ := ep.pollData(mat, symbolPollData("testdata/symbol_ssd_cache1.json"), set.New())
	assert.True(t, count2 > 0)
	assert.Equal(t, len(mat.GetInstances()), 2) // one per controller

	instA := mat.GetInstance("a")
	if instA == nil {
		t.Fatal("expected instance \"a\" after cache creation poll")
	}
}

func TestSymbolSsdCache_CacheRemovedOverPoll(t *testing.T) {
	ep := newEseriesPerfSymbol()
	mat := ep.Matrix[ep.Object]
	mat.SetGlobalLabel("array_id", "test-system")
	mat.SetGlobalLabel("array", "test")

	// Poll 1: cache present
	ep.pollData(mat, symbolPollData("testdata/symbol_ssd_cache1.json"), set.New())
	assert.Equal(t, len(mat.GetInstances()), 2) // one per controller

	// Poll 2: cache removed — record which instances existed
	oldInstances := set.New()
	for key := range mat.GetInstances() {
		oldInstances.Add(key)
	}

	prevMat := mat.Clone()
	curMat := prevMat.CloneForCollection()
	curMat.Reset()

	// No results → pollData touches no instances, so oldInstances retains the stale key
	ep.pollData(curMat, []gjson.Result{}, oldInstances)

	// Simulate PollData's stale removal
	for key := range oldInstances.Iter() {
		curMat.RemoveInstance(key)
	}

	assert.Equal(t, len(curMat.GetInstances()), 0)
	assert.Nil(t, curMat.GetInstance("a"))
	assert.Nil(t, curMat.GetInstance("b"))
}

func buildEPWithVersion(version string) *EseriesPerf {
	ep := new(EseriesPerf)
	ep.ESeries = new(eseries.ESeries)
	ep.AbstractCollector = &collector.AbstractCollector{}
	ep.Remote = conf.Remote{Version: version}
	return ep
}

func TestIsLegacyFlashCache(t *testing.T) {
	tests := []struct {
		version  string
		expected bool
	}{
		{"11.70.0", true},
		{"11.80.0", true},
		{"11.90.0", true},
		{"11.99.9", true},
		{"12.00.0", false},
		{"12.10.0", false},
		{"13.00.0", false},
		{"", false},
		{"garbage", false},
		{"not-a-version", false},
	}
	for _, tt := range tests {
		t.Run("version="+tt.version, func(t *testing.T) {
			assert.Equal(t, isLegacyFlashCache(buildEPWithVersion(tt.version)), tt.expected)
		})
	}
}
