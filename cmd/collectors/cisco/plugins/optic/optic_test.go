package optic

import (
	"github.com/google/go-cmp/cmp"
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"os"
	"path/filepath"
	"testing"
)

func TestNewOpticModels_9K(t *testing.T) {
	tests := []struct {
		name string
		want Model
	}{
		{name: "case0_empty", want: Model{Name: ""}},
		{name: "case1_empty", want: Model{Name: ""}},
		{name: "Ethernet1/15", want: Model{Name: "Ethernet1/15", RxPower: -3.72, TxPower: -2.38}},
		{name: "case3_empty", want: Model{Name: ""}},
		{name: "Ethernet1/22/4", want: Model{Name: "Ethernet1/22/4", RxPower: 0, TxPower: -2.80}},
	}

	testOpticModel(t, "N9K-C9336C-FX2_10.3.4-show_interface_transceiver_details.json", tests)
}

func testOpticModel(t *testing.T, testpath string, tests []struct {
	name string
	want Model
}) {
	// Read the file from the testdata directory
	filename := filepath.Join("testdata", testpath)
	data, err := os.ReadFile(filename)
	assert.Nil(t, err)

	result := gjson.ParseBytes(data)
	jsons := result.Get("TABLE_interface.ROW_interface").Array()

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewOpticModel(jsons[i])
			diff1 := cmp.Diff(tt.want, got)
			assert.Equal(t, diff1, "")
		})
	}
}

func TestNewOpticModels_5K(t *testing.T) {
	tests := []struct {
		name string
		want Model
	}{
		{
			name: "case0_empty", want: Model{Name: ""},
		},
		{
			name: "Ethernet1/25", want: Model{Name: "Ethernet1/25", RxPower: -2.76, TxPower: -2.29},
		},
		{
			name: "case2_empty", want: Model{Name: ""},
		},
	}

	testOpticModel(t, "N5K-show_interface_transceiver_details.json", tests)
}
