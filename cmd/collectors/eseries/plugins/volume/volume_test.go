package volume

import (
	"testing"

	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/cmd/poller/plugin"
	"github.com/netapp/harvest/v2/pkg/matrix"
)

func TestFormatWWID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard_wwid",
			input:    "6D039EA000DCD9780000268E695FC9BE",
			expected: "6D:03:9E:A0:00:DC:D9:78:00:00:26:8E:69:5F:C9:BE",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "",
		},
		{
			name:     "short_wwid",
			input:    "AB12",
			expected: "AB:12",
		},
		{
			name:     "odd_length",
			input:    "ABC",
			expected: "AB:C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatWWID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVolumePlugin_Run(t *testing.T) {
	mat := matrix.New("eseries_volume", "volume", "volume")

	instance, err := mat.NewInstance("vol1")
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}
	instance.SetLabel("wwid", "6D039EA000DCD9780000268E695FC9BE")

	instance2, err := mat.NewInstance("vol2")
	if err != nil {
		t.Fatalf("failed to create instance: %v", err)
	}
	instance2.SetLabel("wwid", "AABBCCDDEEFF00112233445566778899")

	p := &Volume{
		AbstractPlugin: &plugin.AbstractPlugin{
			Object: "eseries_volume",
		},
	}

	dataMap := map[string]*matrix.Matrix{
		"eseries_volume": mat,
	}

	_, _, err = p.Run(dataMap)
	if err != nil {
		t.Fatalf("plugin run failed: %v", err)
	}

	wwid1 := instance.GetLabel("wwid")
	assert.Equal(t, wwid1, "6D:03:9E:A0:00:DC:D9:78:00:00:26:8E:69:5F:C9:BE")

	wwid2 := instance2.GetLabel("wwid")
	assert.Equal(t, wwid2, "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99")
}
