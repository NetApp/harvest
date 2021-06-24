package conf

import (
	"fmt"
	"reflect"
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

func TestPollerStructDefaults(t *testing.T) {
	path := "../../cmd/tools/doctor/testdata/testConfig.yml"
	err := LoadHarvestConfig(path)
	if err != nil {
		panic(err)
	}
	t.Run("poller exporters", func(t *testing.T) {
		poller, err := GetPoller2(path, "zeros")
		if err != nil {
			panic(err)
		}
		// the poller does not define exporters but defaults does
		if poller.Exporters == nil {
			t.Fatalf(`expected exporters to not be nil, but it was`)
		}
		if len(*poller.Exporters) != 1 {
			t.Fatalf(`expected 1 exporters but got %v`, *poller.Exporters)
		}
		expected := []string{"prometheusrange"}
		if !reflect.DeepEqual(*poller.Exporters, expected) {
			t.Fatalf(`expected collectors to be %v but was %v`, expected, *poller.Exporters)
		}
	})

	t.Run("poller collector", func(t *testing.T) {
		poller, err := GetPoller2(path, "cluster-01")
		if err != nil {
			panic(err)
		}
		// the poller does not define collectors but defaults does
		if poller.Collectors == nil {
			t.Fatalf(`expected collectors to not be nil, but it was`)
		}
		if len(*poller.Collectors) != 2 {
			t.Fatalf(`expected 2 collectors but got %v`, *poller.Collectors)
		}
		expected := []string{"Zapi", "ZapiPerf"}
		if !reflect.DeepEqual(*poller.Collectors, expected) {
			t.Fatalf(`expected collectors to be %v but was %v`, expected, *poller.Collectors)
		}
	})

	t.Run("poller username", func(t *testing.T) {
		poller, err := GetPoller2(path, "zeros")
		if err != nil {
			panic(err)
		}
		// the poller does not define a username but defaults does
		if *poller.Username != "myuser" {
			t.Fatalf(`expected username to be [myuser] but was [%v]`, *poller.Username)
		}
	})
}
