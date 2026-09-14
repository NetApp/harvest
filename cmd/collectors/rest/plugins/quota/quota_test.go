package quota

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/options"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/conf"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

// newQuota builds the plugin the way a poller does. This plugin has no client,
// so Init runs for real. Pass qtreeMetrics to enable the qtree-prefixed clone.
func newQuota(t *testing.T, qtreeMetrics bool) *Quota {
	t.Helper()
	params := node.NewS("Quota")
	if qtreeMetrics {
		params.NewChildS("qtreeMetrics", "")
	}
	p := plugin.New("Rest", options.New(), params, nil, "quota", nil)
	q := &Quota{AbstractPlugin: p}
	if err := q.Init(conf.Remote{}); err != nil {
		t.Fatalf("Init: %v", err)
	}
	return q
}

// quotaMatrix builds a quota matrix with the space metrics the REST template
// collects.
func quotaMatrix(t *testing.T) (*matrix.Matrix, *matrix.Instance) {
	t.Helper()
	data := matrix.New("Quota", "quota", "quota")
	inst, err := data.NewInstance("quota-1")
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	return data, inst
}

func newMetric(t *testing.T, data *matrix.Matrix, name string) *matrix.Metric {
	t.Helper()
	m, err := data.NewMetricFloat64(name)
	if err != nil {
		t.Fatalf("NewMetricFloat64(%s): %v", name, err)
	}
	return m
}

// TestRunAddsThresholdMetric covers the REST/ZAPI parity metric. The REST quota
// template has no threshold counter, so the plugin synthesizes one; without it
// the ZAPI and REST quota dashboards disagree.
func TestRunAddsThresholdMetric(t *testing.T) {
	q := newQuota(t, false)
	data, _ := quotaMatrix(t)

	assert.Nil(t, data.GetMetric("threshold"))

	_, _, err := q.Run(map[string]*matrix.Matrix{"quota": data})
	assert.Nil(t, err)

	assert.NotNil(t, data.GetMetric("threshold"))
}

// TestSpaceLimitsConvertedToKibibytes covers GitHub issue #3070
// "Quota dashboard should use kibibytes instead of kilobytes". ONTAP reports
// these in bytes; the plugin divides by 1024 and labels the unit so the
// dashboard's formatting matches.
func TestSpaceLimitsConvertedToKibibytes(t *testing.T) {
	q := newQuota(t, false)
	data, inst := quotaMatrix(t)

	hard := newMetric(t, data, "space.hard_limit")
	soft := newMetric(t, data, "space.soft_limit")
	used := newMetric(t, data, "space.used.total")

	hard.SetValueFloat64(inst, 10240)
	soft.SetValueFloat64(inst, 5120)
	used.SetValueFloat64(inst, 2048)

	_, _, err := q.Run(map[string]*matrix.Matrix{"quota": data})
	assert.Nil(t, err)

	gotHard, ok := data.GetMetric("space.hard_limit").GetValueFloat64(inst)
	assert.True(t, ok)
	assert.Equal(t, gotHard, 10.0)

	gotSoft, ok := data.GetMetric("space.soft_limit").GetValueFloat64(inst)
	assert.True(t, ok)
	assert.Equal(t, gotSoft, 5.0)

	gotUsed, ok := data.GetMetric("space.used.total").GetValueFloat64(inst)
	assert.True(t, ok)
	assert.Equal(t, gotUsed, 2.0)

	assert.Equal(t, data.GetMetric("space.hard_limit").GetLabel("unit"), "kibibytes")
	assert.Equal(t, data.GetMetric("space.soft_limit").GetLabel("unit"), "kibibytes")
	assert.Equal(t, data.GetMetric("space.used.total").GetLabel("unit"), "kibibytes")
}

// TestSoftLimitPopulatesThreshold pins that the synthesized threshold metric is
// filled from space.soft_limit, in kibibytes, with the unit label set.
//
// handlingQuotaMetrics iterates data.GetMetrics(), which includes the threshold
// metric it just created, and writes into threshold from inside that loop. Go
// randomizes map iteration order, but the result converges either way, so a
// single run is sufficient:
//   - threshold visited first: it has no value, so it is set to the -1
//     sentinel, and space.soft_limit later overwrites it with the converted
//     value.
//   - space.soft_limit visited first: it sets threshold to the converted value,
//     and when threshold is reached its own value reads back unchanged, since
//     "threshold" is not one of the space.* names that get converted.
func TestSoftLimitPopulatesThreshold(t *testing.T) {
	q := newQuota(t, false)
	data, inst := quotaMatrix(t)

	soft := newMetric(t, data, "space.soft_limit")
	soft.SetValueFloat64(inst, 8192)

	_, _, err := q.Run(map[string]*matrix.Matrix{"quota": data})
	assert.Nil(t, err)

	threshold := data.GetMetric("threshold")
	assert.NotNil(t, threshold)
	got, ok := threshold.GetValueFloat64(inst)
	assert.True(t, ok)
	assert.Equal(t, got, 8.0)
	assert.Equal(t, threshold.GetLabel("unit"), "kibibytes")
}

// TestUnlimitedQuotaIsMinusOne covers the sentinel the dashboards use to render
// "unlimited": a metric with no value from ONTAP becomes -1 rather than being
// left absent, so the panel shows a value instead of a gap.
func TestUnlimitedQuotaIsMinusOne(t *testing.T) {
	q := newQuota(t, false)
	data, inst := quotaMatrix(t)

	// Declared but never given a value, which is how ONTAP reports "no limit".
	newMetric(t, data, "space.hard_limit")
	newMetric(t, data, "files.hard_limit")

	_, _, err := q.Run(map[string]*matrix.Matrix{"quota": data})
	assert.Nil(t, err)

	gotSpace, ok := data.GetMetric("space.hard_limit").GetValueFloat64(inst)
	assert.True(t, ok)
	assert.Equal(t, gotSpace, -1.0)

	gotFiles, ok := data.GetMetric("files.hard_limit").GetValueFloat64(inst)
	assert.True(t, ok)
	assert.Equal(t, gotFiles, -1.0)

	// An unset metric must not acquire a byte unit it was never converted into.
	assert.Equal(t, data.GetMetric("space.hard_limit").GetLabel("unit"), "")
}

// TestUserAndGroupLabels covers the label derivation behind GitHub issues #2867
// "Quota metrics should remove index from label" and #2951 "Harvest quota
// dashboard should include svm, qtree, user, and group columns". Note that for
// a group quota the id comes from the userId field, not a groupId field.
func TestUserAndGroupLabels(t *testing.T) {
	tests := []struct {
		name        string
		quotaType   string
		userName    string
		userID      string
		groupName   string
		wantUser    string
		wantUserID  string
		wantGroup   string
		wantGroupID string
	}{
		{
			name: "user quota", quotaType: "user",
			userName: "alice", userID: "1001",
			wantUser: "alice", wantUserID: "1001",
		},
		{
			name: "group quota", quotaType: "group",
			groupName: "eng", userID: "2002",
			wantGroup: "eng", wantGroupID: "2002",
		},
		{
			name: "tree quota sets neither", quotaType: "tree",
			userName: "alice", userID: "1001", groupName: "eng",
		},
		{
			name: "empty type sets neither", quotaType: "",
			userName: "alice", userID: "1001",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			q := newQuota(t, false)
			data, inst := quotaMatrix(t)

			inst.SetLabel("type", tc.quotaType)
			inst.SetLabel("userName", tc.userName)
			inst.SetLabel("userId", tc.userID)
			inst.SetLabel("groupName", tc.groupName)

			_, _, err := q.Run(map[string]*matrix.Matrix{"quota": data})
			assert.Nil(t, err)

			assert.Equal(t, inst.GetLabel("user"), tc.wantUser)
			assert.Equal(t, inst.GetLabel("user_id"), tc.wantUserID)
			assert.Equal(t, inst.GetLabel("group"), tc.wantGroup)
			assert.Equal(t, inst.GetLabel("group_id"), tc.wantGroupID)
		})
	}
}

// TestQtreeMetricsClone covers the backward-compatibility clone: the same data
// is republished under the qtree object so the older qtree-prefixed metric names
// keep working.
func TestQtreeMetricsClone(t *testing.T) {
	q := newQuota(t, true)
	data, inst := quotaMatrix(t)

	soft := newMetric(t, data, "space.soft_limit")
	soft.SetValueFloat64(inst, 4096)

	got, _, err := q.Run(map[string]*matrix.Matrix{"quota": data})
	assert.Nil(t, err)

	assert.Equal(t, len(got), 1)
	clone := got[0]
	assert.Equal(t, clone.Object, "qtree")
	assert.Equal(t, clone.Identifier, "qtree")
	// The clone carries the converted values, not the raw bytes.
	cloneInst := clone.GetInstance("quota-1")
	assert.NotNil(t, cloneInst)
	gotSoft, ok := clone.GetMetric("space.soft_limit").GetValueFloat64(cloneInst)
	assert.True(t, ok)
	assert.Equal(t, gotSoft, 4.0)
}

// TestQtreeMetricsDisabledReturnsNothing is the control for the clone path.
func TestQtreeMetricsDisabledReturnsNothing(t *testing.T) {
	q := newQuota(t, false)
	data, _ := quotaMatrix(t)

	got, meta, err := q.Run(map[string]*matrix.Matrix{"quota": data})

	assert.Nil(t, err)
	assert.Nil(t, got)
	assert.Nil(t, meta)
}

// TestNonExportableInstanceIsSkipped pins that a filtered-out quota is not
// given the -1 sentinel, which would otherwise publish a series for a row the
// template meant to drop.
func TestNonExportableInstanceIsSkipped(t *testing.T) {
	q := newQuota(t, false)
	data, inst := quotaMatrix(t)
	inst.SetExportable(false)

	hard := newMetric(t, data, "space.hard_limit")

	_, _, err := q.Run(map[string]*matrix.Matrix{"quota": data})
	assert.Nil(t, err)

	_, ok := hard.GetValueFloat64(inst)
	assert.False(t, ok)
}
