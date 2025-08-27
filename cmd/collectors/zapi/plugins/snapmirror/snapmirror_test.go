package snapmirror

import (
	"github.com/netapp/harvest/v2/assert"
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
	testAutomatedFailoverDuplex(t, instance)
	testOtherPolicyType(t, instance)
	testWithNoPolicyType(t, instance)
}

// Test cases for protectedBy and protectionSourceType
func testWithoutGroupType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "")
}

func testSvmdr(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "vserver")
	instance.SetLabel("destination_volume", "")
	instance.SetLabel("source_volume", "")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "storage_vm")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "storage_vm")
}

func testConstituentVolumeWithinSvmdr(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "vserver")
	instance.SetLabel("destination_volume", "destvol")
	instance.SetLabel("source_volume", "sourcevol")
	instance.SetLabel("destination_location", "test1")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "storage_vm")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "volume")
}

func testCg(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "consistencygroup")
	instance.SetLabel("destination_location", "test123:/cg/")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "cg")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "cg")
}

func testConstituentVolumeWithinCg(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "consistencygroup")
	instance.SetLabel("destination_location", "test123")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "cg")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "volume")
}

func testNegativeCase1(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "infinitevol")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "volume")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "not_mapped")
}

func testNegativeCase2(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "vserver")
	instance.SetLabel("destination_volume", "destvol")
	instance.SetLabel("source_volume", "sourcevol")
	instance.SetLabel("destination_location", "test123:")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "volume")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "not_mapped")
}

func testGroupTypeNone(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "none")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "volume")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "volume")
}

func testGroupTypeFlexgroup(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("group_type", "flexgroup")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("protectedBy"), "volume")
	assert.Equal(t, instance.GetLabel("protectionSourceType"), "volume")
}

// Test cases for derived_relationship_type
func testStrictSyncMirror(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "strict_sync_mirror")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "sync_mirror_strict")
}

func testSyncMirror(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "sync_mirror")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "sync_mirror")
}

func testMirrorVault(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "mirror_vault")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "mirror_vault")
}

func testAutomatedFailover(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "automated_failover")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "automated_failover")
}

func testAutomatedFailoverDuplex(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "")
	instance.SetLabel("policy_type", "automated_failover_duplex")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "automated_failover_duplex")
}

func testOtherPolicyType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "vault")
	instance.SetLabel("policy_type", "vault")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "vault")
}

func testWithNoPolicyType(t *testing.T, instance *matrix.Instance) {
	instance.SetLabel("relationship_type", "extended_data_protection")
	instance.SetLabel("policy_type", "")
	collectors.UpdateProtectedFields(instance)

	assert.Equal(t, instance.GetLabel("derived_relationship_type"), "extended_data_protection")
}
