package collectors

import (
	"goharvest2/pkg/matrix"
	"testing"
)

func TestUpdateProtectedFields(t *testing.T) {
	instance := matrix.NewInstance(0)

	// Test cases for protectedBy and protectionSourceType
	testWithoutGroupType(instance, t)
	testSvmdr(instance, t)
	testConstituentVolumeWithinSvmdr(instance, t)
	testCg(instance, t)
	testConstituentVolumeWithinCg(instance, t)
	testNegativeCase1(instance, t)
	testNegativeCase2(instance, t)
	testGroupTypeNone(instance, t)
	testGroupTypeFlexgroup(instance, t)

	// Test cases for derived_relationship_type
	testStrictSyncMirror(instance, t)
	testSyncMirror(instance, t)
	testMirrorVault(instance, t)
	testAutomatedFailover(instance, t)
	testOtherPolicyType(instance, t)
	testWithNoPolicyType(instance, t)
	testWithNoPolicyTypeNoRelationshipType(instance, t)
}

// Test cases for protectedBy and protectionSourceType
func testWithoutGroupType(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("group_type", "")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "" && instance.GetLabel("protectionSourceType") == "" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected empty  and protectionSourceType= %s, expected empty", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testSvmdr(instance *matrix.Instance, t *testing.T) {
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

func testConstituentVolumeWithinSvmdr(instance *matrix.Instance, t *testing.T) {
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

func testCg(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("group_type", "CONSISTENCYGROUP")
	instance.SetLabel("destination_location", "test123:/cg/")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "cg" && instance.GetLabel("protectionSourceType") == "cg" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: cg and protectionSourceType= %s, expected: cg", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testConstituentVolumeWithinCg(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("group_type", "CONSISTENCYGROUP")
	instance.SetLabel("destination_location", "test123")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "cg" && instance.GetLabel("protectionSourceType") == "volume" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: cg and protectionSourceType= %s, expected: volume", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testNegativeCase1(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("group_type", "infinitevol")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "volume" && instance.GetLabel("protectionSourceType") == "not_mapped" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: volume and protectionSourceType= %s, expected: not_mapped", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testNegativeCase2(instance *matrix.Instance, t *testing.T) {
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

func testGroupTypeNone(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("group_type", "none")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "volume" && instance.GetLabel("protectionSourceType") == "volume" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: volume and protectionSourceType= %s, expected: volume", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

func testGroupTypeFlexgroup(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("group_type", "flexgroup")
	UpdateProtectedFields(instance)

	if instance.GetLabel("protectedBy") == "volume" && instance.GetLabel("protectionSourceType") == "volume" {
		// OK
	} else {
		t.Errorf("Labels protectedBy= %s, expected: volume and protectionSourceType= %s, expected: volume", instance.GetLabel("protectedBy"), instance.GetLabel("protectionSourceType"))
	}
}

// Test cases for derived_relationship_type
func testStrictSyncMirror(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "strict_sync_mirror")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "sync_mirror_strict" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: sync_mirror_strict", instance.GetLabel("derived_relationship_type"))
	}
}

func testSyncMirror(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "sync_mirror")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "sync_mirror" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: sync_mirror", instance.GetLabel("derived_relationship_type"))
	}
}

func testMirrorVault(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "mirror_vault")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "mirror_vault" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: mirror_vault", instance.GetLabel("derived_relationship_type"))
	}
}

func testAutomatedFailover(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "automated_failover")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "sync_mirror" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: sync_mirror", instance.GetLabel("derived_relationship_type"))
	}
}

func testOtherPolicyType(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("relationship_type", "vault")
	instance.SetLabel("policy_type", "vault")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "vault" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: vault", instance.GetLabel("derived_relationship_type"))
	}
}

func testWithNoPolicyType(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("relationship_type", "extended_data_protection")
	instance.SetLabel("policy_type", "")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "extended_data_protection" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: extended_data_protection", instance.GetLabel("derived_relationship_type"))
	}
}

func testWithNoPolicyTypeNoRelationshipType(instance *matrix.Instance, t *testing.T) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "")
	UpdateProtectedFields(instance)

	if instance.GetLabel("derived_relationship_type") == "" {
		// OK
	} else {
		t.Errorf("Labels derived_relationship_type= %s, expected: \"\"(empty)", instance.GetLabel("derived_relationship_type"))
	}
}
