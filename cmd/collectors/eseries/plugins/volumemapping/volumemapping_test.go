package volumemapping

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
	"github.com/netapp/harvest/v2/pkg/tree/node"
)

func createMockPlugin() *VolumeMapping {
	params := node.NewS("VolumeMapping")
	p := plugin.New("ESeries", nil, params, nil, "eseries_volume", nil)
	_ = p.InitAbc()
	return &VolumeMapping{AbstractPlugin: p}
}

func createVolumeInstance(t *testing.T) *matrix.Instance {
	t.Helper()
	mat := matrix.New("test", "eseries_volume", "eseries_volume")
	inst, err := mat.NewInstance("vol-1")
	if err != nil {
		t.Fatalf("failed to create volume instance: %v", err)
	}
	inst.SetLabel("volume", "vol-1")
	return inst
}

func TestAddLunAndHostLabels_Basic(t *testing.T) {
	v := createMockPlugin()
	inst := createVolumeInstance(t)
	inst.SetLabel("list_of_mappings",
		`{"lun":"1","mapRef":"host-ref-1","type":"host"},{"lun":"2","mapRef":"cluster-ref-1","type":"cluster"}`)

	hostNames := map[string]string{"host-ref-1": "esxi-01"}
	hostClusterNames := map[string]string{"cluster-ref-1": "vsphere-cluster"}

	v.addLunAndHostLabels(inst, hostNames, hostClusterNames)

	assert.Equal(t, inst.GetLabel("luns"), "1,2")
	assert.Equal(t, inst.GetLabel("hosts"), "esxi-01,vsphere-cluster")
	assert.Equal(t, inst.GetLabel("mapping_types"), "host,cluster")
}

func TestAddLunAndHostLabels_EmptyOrMalformed(t *testing.T) {
	tests := []struct {
		name           string
		listOfMappings string
	}{
		{"empty label", ""},
		{"malformed label", "not-json-at-all"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := createMockPlugin()
			inst := createVolumeInstance(t)
			if tt.listOfMappings != "" {
				inst.SetLabel("list_of_mappings", tt.listOfMappings)
			}

			v.addLunAndHostLabels(inst, map[string]string{}, map[string]string{})

			assert.Equal(t, inst.GetLabel("luns"), "")
			assert.Equal(t, inst.GetLabel("hosts"), "")
			assert.Equal(t, inst.GetLabel("mapping_types"), "")
		})
	}
}

func TestAddWorkloadLabel_Resolved(t *testing.T) {
	v := createMockPlugin()
	inst := createVolumeInstance(t)
	inst.SetLabel("metadata",
		`{"key":"createdBy","value":"SANtricity System Manager"},{"key":"workloadId","value":"wl-1"}`)

	v.addWorkloadLabel(inst, map[string]string{"wl-1": "database-workload"})

	assert.Equal(t, inst.GetLabel("workload"), "database-workload")
}

func TestAddWorkloadLabel_UnresolvedFallsBackToID(t *testing.T) {
	v := createMockPlugin()
	inst := createVolumeInstance(t)
	inst.SetLabel("metadata", `{"key":"workloadId","value":"wl-unknown"}`)

	v.addWorkloadLabel(inst, map[string]string{})

	assert.Equal(t, inst.GetLabel("workload"), "wl-unknown")
}

func TestAddWorkloadLabel_EmptyOrMalformed(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
	}{
		{"empty label", ""},
		{"malformed label", "not-json-at-all"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := createMockPlugin()
			inst := createVolumeInstance(t)
			if tt.metadata != "" {
				inst.SetLabel("metadata", tt.metadata)
			}

			v.addWorkloadLabel(inst, map[string]string{})

			assert.Equal(t, inst.GetLabel("workload"), "")
		})
	}
}
