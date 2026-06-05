package rest

import (
	"github.com/netapp/harvest/v2/assert"
	"github.com/netapp/harvest/v2/third_party/tidwall/gjson"
	"os"
	"testing"
)

func Test_aristaVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"EOS 4.18", "4.18.5M", "4.18.5"},
		{"EOS 4.30", "4.30.2F", "4.30.2"},
		{"EOS 4.27 maint", "4.27.3M-29413748.demo", "4.27.3"},
		{"no match", "unknown", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, aristaVersion(tt.input), tt.want)
		})
	}
}

func TestParseVersionResponse(t *testing.T) {
	data, err := os.ReadFile("testdata/show_version.json")
	assert.Nil(t, err)

	result := gjson.ParseBytes(data).Get("result.0")
	assert.Equal(t, result.Get("modelName").String(), "DCS-7010T-48-R")
	assert.Equal(t, result.Get("serialNumber").String(), "JPE17011101")
	assert.Equal(t, aristaVersion(result.Get("version").String()), "4.18.5")
}

func TestParseHostnameResponse(t *testing.T) {
	data, err := os.ReadFile("testdata/show_hostname.json")
	assert.Nil(t, err)

	result := gjson.ParseBytes(data).Get("result.0")
	assert.Equal(t, result.Get("hostname").String(), "sa-tme-flexpod-kk19-7010T-1g")
}

func TestParseErrorResponse(t *testing.T) {
	data, err := os.ReadFile("testdata/banner_error.json")
	assert.Nil(t, err)

	parsed := gjson.ParseBytes(data)
	apiErr := parsed.Get("error")
	assert.Equal(t, apiErr.Exists(), true)
	assert.Equal(t, apiErr.Get("code").Int(), int64(1002))
}
