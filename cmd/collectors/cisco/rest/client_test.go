package rest

import (
	"github.com/netapp/harvest/v2/assert"
	"testing"
)

func Test_ciscoVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Nexus 9000 10.3", "10.3(4a)", "10.3.4"},
		{"Nexus 9000 10.4", "10.4(4)", "10.4.4"},
		{"Nexus 5000 7.3", "7.3(8)N1(1)", "7.3.8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, ciscoVersion(tt.input), tt.want)
		})
	}
}
