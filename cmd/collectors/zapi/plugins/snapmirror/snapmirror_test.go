package snapmirror

import (
	"github.com/netapp/harvest/v2/cmd/collectors"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"testing"
)

func TestProtectedFields(t *testing.T) {
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
}

// Test cases for protectedBy and protectionSourceType
func testWithoutGroupType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "")
	collectors.UpdateProtectedFields(instance)

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
	collectors.UpdateProtectedFields(instance)

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
	collectors.UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "storage_vm" && instance.GetLabel("protectionSourceType") == "volume" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: storage_vm and protectionSourceType= %s, expected: volume", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testCg(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "consistencygroup")
	instance.SetLabel("destination_location", "test123:/cg/")
	collectors.UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "cg" && instance.GetLabel("protectionSourceType") == "cg" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: cg and protectionSourceType= %s, expected: cg", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testConstituentVolumeWithinCg(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "consistencygroup")
	instance.SetLabel("destination_location", "test123")
	collectors.UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "cg" && instance.GetLabel("protectionSourceType") == "volume" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: cg and protectionSourceType= %s, expected: volume", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testNegativeCase1(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "infinitevol")
	collectors.UpdateProtectedFields(instance)

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
	collectors.UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "volume" && instance.GetLabel("protectionSourceType") == "not_mapped" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: volume and protectionSourceType= %s, expected: not_mapped", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testGroupTypeNone(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "none")
	collectors.UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "volume" && instance.GetLabel("protectionSourceType") == "volume" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: volume and protectionSourceType= %s, expected: volume", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testGroupTypeFlexgroup(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "flexgroup")
	collectors.UpdateProtectedFields(instance)

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
	collectors.UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "sync_mirror_strict" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: sync_mirror_strict", instance.GetLabel("derived_relationship_type"))
	}
}

func testSyncMirror(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "sync_mirror")
	collectors.UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "sync_mirror" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: sync_mirror", instance.GetLabel("derived_relationship_type"))
	}
}

func testMirrorVault(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "mirror_vault")
	collectors.UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "mirror_vault" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: mirror_vault", instance.GetLabel("derived_relationship_type"))
	}
}

func testAutomatedFailover(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "automated_failover")
	collectors.UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "sync_mirror" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: sync_mirror", instance.GetLabel("derived_relationship_type"))
	}
}

func testOtherPolicyType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "vault")
	instance.SetLabel("policy_type", "vault")
	collectors.UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "vault" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: vault", instance.GetLabel("derived_relationship_type"))
	}
}

func testWithNoPolicyType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "extended_data_protection")
	instance.SetLabel("policy_type", "")
	collectors.UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "extended_data_protection" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: extended_data_protection", instance.GetLabel("derived_relationship_type"))
	}
}
