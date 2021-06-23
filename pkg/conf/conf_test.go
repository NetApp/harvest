package conf

import (
	"fmt"
	"testing"
)

func TestGetPrometheusExporterPorts(t *testing.T) {
	path := "../../cmd/tools/doctor/testdata/testConfig.yml"
	err := LoadHarvestConfig(path)
	// Test without checking
	ValidatePortInUse = true
	if err != nil {
		panic(err)
	}
	type args struct {
		pollerName string
	}

	type test struct {
		name    string
		args    args
		wantErr bool
	}
	tests := []test{
		{"success", args{pollerName: "unix-01"}, false},
		{"failure_no_free_port", args{pollerName: "cluster-02"}, true},
		{"failure_poller_name_does_not_exist", args{pollerName: "test1"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetPrometheusExporterPorts(tt.args.pollerName)
			fmt.Println(got)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPrometheusExporterPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got == 0 {
				t.Errorf("GetPrometheusExporterPorts() got = %v, want %s", got, "non zero value")
			}
		})
	}
}
