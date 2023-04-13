package collectors

import (
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

	if instance.GetLabel("protectedBy") == "" && instance.GetLabel("protectionSourceType") == "" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected empty  and protectionSourceType= %s, expected empty", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testSvmdr(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "vserver")
	instance.SetLabel("destination_volume", "")
	instance.SetLabel("source_volume", "")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "storage_vm" && instance.GetLabel("protectionSourceType") == "storage_vm" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: storage_vm and protectionSourceType= %s, expected: storage_vm", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testConstituentVolumeWithinSvmdr(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "vserver")
	instance.SetLabel("destination_volume", "destvol")
	instance.SetLabel("source_volume", "sourcevol")
	instance.SetLabel("destination_location", "test1")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "storage_vm" && instance.GetLabel("protectionSourceType") == "volume" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: storage_vm and protectionSourceType= %s, expected: volume", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testCg(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "consistencygroup")
	instance.SetLabel("destination_location", "test123:/cg/")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "cg" && instance.GetLabel("protectionSourceType") == "cg" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: cg and protectionSourceType= %s, expected: cg", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testConstituentVolumeWithinCg(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "consistencygroup")
	instance.SetLabel("destination_location", "test123")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "cg" && instance.GetLabel("protectionSourceType") == "volume" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: cg and protectionSourceType= %s, expected: volume", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testNegativeCase1(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "infinitevol")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "volume" && instance.GetLabel("protectionSourceType") == "not_mapped" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: volume and protectionSourceType= %s, expected: not_mapped", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testNegativeCase2(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "vserver")
	instance.SetLabel("destination_volume", "destvol")
	instance.SetLabel("source_volume", "sourcevol")
	instance.SetLabel("destination_location", "test123:")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "volume" && instance.GetLabel("protectionSourceType") == "not_mapped" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: volume and protectionSourceType= %s, expected: not_mapped", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testGroupTypeNone(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "none")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "volume" && instance.GetLabel("protectionSourceType") == "volume" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: volume and protectionSourceType= %s, expected: volume", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testGroupTypeFlexgroup(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "flexgroup")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "volume" && instance.GetLabel("protectionSourceType") == "volume" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: volume and protectionSourceType= %s, expected: volume", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

// Test cases for derived_relationship_type
func testStrictSyncMirror(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "strict_sync_mirror")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "sync_mirror_strict" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: sync_mirror_strict", instance.GetLabel("derived_relationship_type"))
	}
}

func testSyncMirror(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "sync_mirror")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "sync_mirror" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: sync_mirror", instance.GetLabel("derived_relationship_type"))
	}
}

func testMirrorVault(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "mirror_vault")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "mirror_vault" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: mirror_vault", instance.GetLabel("derived_relationship_type"))
	}
}

func testAutomatedFailover(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "automated_failover")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "sync_mirror" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: sync_mirror", instance.GetLabel("derived_relationship_type"))
	}
}

func testOtherPolicyType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "vault")
	instance.SetLabel("policy_type", "vault")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "vault" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: vault", instance.GetLabel("derived_relationship_type"))
	}
}

func testWithNoPolicyType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "extended_data_protection")
	instance.SetLabel("policy_type", "")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "extended_data_protection" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: extended_data_protection", instance.GetLabel("derived_relationship_type"))
	}
}

func testWithNoPolicyTypeNoRelationshipType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: \"\"(empty)", instance.GetLabel("derived_relationship_type"))
	}
}

// Test cases for timestamp comparison with duration
func testOlderTimestampThanDuration(t *testing.T) {
	timestamp := float64(time.Now().Add(-20 * time.Minute).UnixMicro())
	duration := 5 * time.Minute
	isOlder := IsTimestampOlderThanDuration(timestamp, duration)

	if isOlder {
		// OK
	} else {
		t.Errorf("timestamp= %f is older than duration %s", timestamp, duration.String())
	}
}

func testNewerTimestampThanDuration(t *testing.T) {
	timestamp := float64(time.Now().Add(-1 * time.Hour).UnixMicro())
	duration := 2 * time.Hour
	isOlder := IsTimestampOlderThanDuration(timestamp, duration)

	if !isOlder {
		// OK
	} else {
		t.Errorf("timestamp= %f is newer than duration %s", timestamp, duration.String())
	}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDataInterval() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Seconds() != tt.want {
				t.Errorf("GetDataInterval() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func generateScheduleParam(duration string) *node.Node {
	root := node.NewS("root")
	param := root.NewChildS("schedule", "")
	param.NewChildS("data", duration)
	return root
}
