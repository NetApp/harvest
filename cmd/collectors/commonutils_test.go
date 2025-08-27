package collectors

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
	"testing"
	"time"
)

func TestUpdateProtectedFields(t *testing.T) {
	instance := matrix.NewInstance(0)

	// Test cases for protectedBy and protectionSourceType
	testWithoutGroupType(t, instance)
	testSvmdr(t, instance)
	testConstituentVolumeWithinSvmdr(t, instance)
	testCg(t, instance)
	testConstituentVolumeWithinCg(t, instance)
	testNegativeCase1(t, instance)
	testNegativeCase2(t, instance)
	testGroupTypeNone(t, instance)
	testGroupTypeFlexgroup(t, instance)

	// Test cases for derived_relationship_type
	testStrictSyncMirror(t, instance)
	testSyncMirror(t, instance)
	testMirrorVault(t, instance)
	testAutomatedFailover(t, instance)
	testAutomatedFailoverDuplex(t, instance)
	testOtherPolicyType(t, instance)
	testWithNoPolicyType(t, instance)
	testWithNoPolicyTypeNoRelationshipType(t, instance)
}

func TestIsTimestampOlderThanDuration(t *testing.T) {
	// Test cases for timestamp comparison with Duration
	testOlderTimestampThanDuration(t)
	testNewerTimestampThanDuration(t)
}

// Test cases for protectedBy and protectionSourceType
func testWithoutGroupType(t *testing.T, instance *matrix.Instance) {
	t.Helper()
	instance.SetLabel("group_type", "")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "")
}

func testSvmdr(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "vserver")
	instance.SetLabel("destination_volume", "")
	instance.SetLabel("source_volume", "")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "storage_vm")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "storage_vm")
}

func testConstituentVolumeWithinSvmdr(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "vserver")
	instance.SetLabel("destination_volume", "destvol")
	instance.SetLabel("source_volume", "sourcevol")
	instance.SetLabel("destination_location", "test1")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "storage_vm")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "volume")
}

func testCg(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "consistencygroup")
	instance.SetLabel("destination_location", "test123:/cg/")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "cg")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "cg")
}

func testConstituentVolumeWithinCg(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "consistencygroup")
	instance.SetLabel("destination_location", "test123")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "cg")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "volume")
}

func testNegativeCase1(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "infinitevol")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "volume")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "not_mapped")
}

func testNegativeCase2(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "vserver")
	instance.SetLabel("destination_volume", "destvol")
	instance.SetLabel("source_volume", "sourcevol")
	instance.SetLabel("destination_location", "test123:")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "volume")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "not_mapped")
}

func testGroupTypeNone(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "none")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "volume")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "volume")
}

func testGroupTypeFlexgroup(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "flexgroup")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "volume")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "volume")
}

// Test cases for derived_relationship_type
func testStrictSyncMirror(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "strict_sync_mirror")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "sync_mirror_strict")
}

func testSyncMirror(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "sync_mirror")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "sync_mirror")
}

func testMirrorVault(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "mirror_vault")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "mirror_vault")
}

func testAutomatedFailover(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "automated_failover")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "automated_failover")
}

func testAutomatedFailoverDuplex(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "automated_failover_duplex")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "automated_failover_duplex")
}

func testOtherPolicyType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "vault")
	instance.SetLabel("policy_type", "vault")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "vault")
}

func testWithNoPolicyType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "extended_data_protection")
	instance.SetLabel("policy_type", "")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "extended_data_protection")
}

func testWithNoPolicyTypeNoRelationshipType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "")
	UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "")
}

// Test cases for timestamp comparison with duration
func testOlderTimestampThanDuration(t *testing.T) {
	timestamp := float64(time.Now().Add(-20 * time.Minute).UnixMicro())
	duration := 5 * time.Minute
	isOlder := IsTimestampOlderThanDuration(time.Now(), timestamp, duration)

	assert.True(t, isOlder)
}

func testNewerTimestampThanDuration(t *testing.T) {
	timestamp := float64(time.Now().Add(-1 * time.Hour).UnixMicro())
	duration := 2 * time.Hour
	isOlder := IsTimestampOlderThanDuration(time.Now(), timestamp, duration)

	assert.False(t, isOlder)
}

func TestGetDataInterval(t *testing.T) {
	defaultDataPollDuration := 3 * time.Minute
	type args struct {
		param           *node.Node
		defaultInterval time.Duration
	}

	type test struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}
	tests := []test{
		{"success_return_poller_schedule", args{param: generateScheduleParam("4m"), defaultInterval: defaultDataPollDuration}, 240, false},
		{"error_return_default_schedule", args{param: generateScheduleParam("4ma"), defaultInterval: defaultDataPollDuration}, 180, true},
		{"return_default_schedule", args{param: generateScheduleParam(""), defaultInterval: defaultDataPollDuration}, 180, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetDataInterval(tt.args.param, tt.args.defaultInterval)
			if err != nil {
				assert.True(t, tt.wantErr)
				return
			}
			assert.Equal(t, got.Seconds(), tt.want)
		})
	}
}

func generateScheduleParam(duration string) *node.Node {
	root := node.NewS("root")
	param := root.NewChildS("schedule", "")
	param.NewChildS("data", duration)
	return root
}

func TestLagTimeBasedOnLastTransferSize(t *testing.T) {
	type test struct {
		name              string
		instance          string
		healthy           string
		schedule          string
		lastTransferError string
		relationshipID    string
		lastBytesValue    float64
		lagTimeValue      float64
		want              float64
	}
	tests := []test{
		{name: "TestBytesNonZero", instance: "InstanceA", healthy: "true", schedule: "5min", lastTransferError: "", relationshipID: "4885136b-4c31-11ec-95a6-00a09865cd13", lastBytesValue: 3479, lagTimeValue: 172652, want: 172652},
		{name: "TestLastErrorNonEmpty", instance: "InstanceB", healthy: "true", schedule: "my_daily", lastTransferError: "The specified Snapshot copy is older than the base Snapshot copy on source volume", relationshipID: "7c4d92e1-6828-11ea-893b-00a09865cd13", lastBytesValue: 0, lagTimeValue: 172223, want: 172223},
		{name: "TestScheduleEmpty", instance: "InstanceC", healthy: "true", schedule: "", lastTransferError: "", relationshipID: "62422099-9c61-11e8-b2ff-00a09865fe59", lastBytesValue: 0, lagTimeValue: 738, want: 738},
		{name: "TestNonHealthy", instance: "InstanceD", healthy: "false", schedule: "daily", lastTransferError: "", relationshipID: "31421f02-7703-11e9-b59e-00a09865cd13", lastBytesValue: 0, lagTimeValue: 892, want: 892},
		{name: "TestLagTimeChangedTo0", instance: "InstanceE", healthy: "true", schedule: "hourly", lastTransferError: "", relationshipID: "b4da7644-76e0-11e9-b59e-00a09865cd13", lastBytesValue: 0, lagTimeValue: 2736, want: 0},
	}

	data := matrix.New("SnapMirrorTest", "SnapMirror", "SnapMirror")
	lastTransferSizeMetric, err := data.NewMetricUint8("last_transfer_size")
	assert.Nil(t, err)
	lagTimeMetric, err := data.NewMetricUint8("lag_time")
	assert.Nil(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := populateInstance(data, tt.instance, tt.healthy, tt.schedule, tt.lastTransferError, tt.relationshipID, lastTransferSizeMetric, tt.lastBytesValue, lagTimeMetric, tt.lagTimeValue)
			UpdateLagTime(instance, lastTransferSizeMetric, lagTimeMetric)
			actualValue, _ := lagTimeMetric.GetValueFloat64(instance)
			assert.Equal(t, actualValue, tt.want)
		})
	}
}

func populateInstance(data *matrix.Matrix, instanceName, healthy, schedule, lastError, relationshipID string, lastTransferSizeMetric *matrix.Metric, bytesData float64, lagTimeMetric *matrix.Metric, lagTime float64) *matrix.Instance {
	instance, err := data.NewInstance(instanceName)
	if err != nil {
		panic(err)
	}
	instance.SetLabel("healthy", healthy)
	instance.SetLabel("schedule", schedule)
	instance.SetLabel("last_transfer_error", lastError)
	instance.SetLabel("relationship_id", relationshipID)

	lastTransferSizeMetric.SetValueFloat64(instance, bytesData)
	lagTimeMetric.SetValueFloat64(instance, lagTime)
	return instance
}

func Test_SplitVscanName(t *testing.T) {
	tests := []struct {
		name      string
		ontapName string
		svm       string
		scanner   string
		node      string
		isValid   bool
		isZapi    bool
	}{
		{
			name:      "valid",
			ontapName: "svm:scanner:node",
			svm:       "svm",
			scanner:   "scanner",
			node:      "node",
			isValid:   true,
			isZapi:    true,
		},
		{
			name:      "ipv6",
			ontapName: "moon-ad:2a03:1e80:a15:60c::1:2a5:moon-02",
			svm:       "moon-ad",
			scanner:   "2a03:1e80:a15:60c::1:2a5",
			node:      "moon-02",
			isValid:   true,
			isZapi:    true,
		},
		{
			name:      "invalid zero colon",
			ontapName: "svm",
			svm:       "",
			scanner:   "",
			node:      "",
			isValid:   false,
			isZapi:    true,
		},
		{
			name:      "invalid one colon",
			ontapName: "svm:scanner",
			svm:       "",
			scanner:   "",
			node:      "",
			isValid:   false,
			isZapi:    true,
		},
		{
			name:      "rest",
			ontapName: "node:svm:scanner",
			svm:       "svm",
			scanner:   "scanner",
			node:      "node",
			isValid:   true,
			isZapi:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanerNames, ok := SplitVscanName(tt.ontapName, tt.isZapi)

			assert.Equal(t, scanerNames.Svm, tt.svm)
			assert.Equal(t, scanerNames.Scanner, tt.scanner)
			assert.Equal(t, scanerNames.Node, tt.node)
			assert.Equal(t, ok, tt.isValid)
		})
	}
}

func Test_HandleDuration(t *testing.T) {

	type test struct {
		timeFieldValue string
		want           float64
	}

	var tests = []test{
		{
			timeFieldValue: "PT54S",
			want:           54,
		},
		{
			timeFieldValue: "PT48M",
			want:           2880,
		},
		{
			timeFieldValue: "P428DT22H45M19S",
			want:           37061119,
		},
		{
			timeFieldValue: "PT8H35M42S",
			want:           30942,
		},
	}

	for _, tt := range tests {
		t.Run(tt.timeFieldValue, func(t *testing.T) {
			assert.Equal(t, HandleDuration(tt.timeFieldValue), tt.want)
		})
	}
}

func Test_HandleTimestamp(t *testing.T) {

	type test struct {
		timeFieldValue string
		want           float64
	}

	var tests = []test{
		{
			timeFieldValue: "2020-12-02T18:36:19-08:00",
			want:           1606962979,
		},
		{
			timeFieldValue: "2022-01-31T04:05:02-05:00",
			want:           1643619902,
		},
	}

	for _, tt := range tests {
		t.Run(tt.timeFieldValue, func(t *testing.T) {
			assert.Equal(t, HandleTimestamp(tt.timeFieldValue), tt.want)
		})
	}
}
