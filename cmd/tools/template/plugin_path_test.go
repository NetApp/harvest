package template

import "testing"

func TestToPluginPathSharedPlugins(t *testing.T) {
	tests := []struct {
		path       string
		pluginName string
		want       string
	}{
		{
			path:       "../../../conf/rest/9.12.0/qos_policy_adaptive.yaml",
			pluginName: "qospolicyadaptive",
			want:       "../../../cmd/collectors/plugins/qospolicyadaptive/qospolicyadaptive.go",
		},
		{
			path:       "conf/zapi/cdot/9.8.0/qos_workload.yaml",
			pluginName: "workload",
			want:       "cmd/collectors/plugins/workload/workload.go",
		},
		{
			path:       "conf/eseriesperf/workload.yaml",
			pluginName: "workload",
			want:       "cmd/collectors/eseriesperf/plugins/workload/workload.go",
		},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			if got := toPluginPath(test.path, test.pluginName); got != test.want {
				t.Fatalf("toPluginPath() = %q, want %q", got, test.want)
			}
		})
	}
}
