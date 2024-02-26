package conf

import (
	"testing"
)

func Test_expandVars(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		setEnvs map[string]string
	}{
		{name: "empty", in: "", want: ""},
		{name: "no env var", in: "hi", want: "hi"},
		{name: "missing an underscore", in: "$_env{NETAPP_U2}", want: "$_env{NETAPP_U2}"},

		{name: "empty var", in: "$__env{ENV}", want: ""},
		{name: "empty var2", in: "${ENV}", want: ""},

		{name: "env1", in: "$__env{VAR1} is cool", want: "Harvest is cool", setEnvs: map[string]string{"VAR1": "Harvest"}},
		{name: "dup", in: "$__env{VAR1} and $__env{VAR1}", want: "Harvest and Harvest", setEnvs: map[string]string{"VAR1": "Harvest"}},
		{name: "both", in: "a $__env{VAR1} $__env{VAR2}", want: "a b c", setEnvs: map[string]string{"VAR1": "b", "VAR2": "c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.setEnvs {
				t.Setenv(k, v)
			}
			g, err := ExpandVars([]byte(tt.in))
			got := string(g)
			if err != nil {
				t.Errorf("ExpandVars() error %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("ExpandVars() got %v, want %v", got, tt.want)
			}
		})
	}
}
