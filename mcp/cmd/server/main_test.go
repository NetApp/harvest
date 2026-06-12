package main

import "testing"

func TestApplyClusterFilter(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		cluster      string
		clusterMatch string
		want         string
	}{
		{
			name:  "no filter - returns query unchanged",
			query: "up",
			want:  "up",
		},
		{
			name:    "bare metric name - appends label selector",
			query:   "up",
			cluster: "cluster1",
			want:    `up{cluster="cluster1"}`,
		},
		{
			name:    "metric with expression - inserts selector before space",
			query:   "volume_size_used_percent > 95",
			cluster: "cluster1",
			want:    `volume_size_used_percent{cluster="cluster1"} > 95`,
		},
		{
			name:    "metric with existing labels - appends inside braces",
			query:   `volume_size_used_percent{svm="vs1"} > 95`,
			cluster: "cluster1",
			want:    `volume_size_used_percent{svm="vs1", cluster="cluster1"} > 95`,
		},
		{
			name:    "metric with labels and no trailing expression - appends inside braces",
			query:   `license_labels{license="snapmirror"}`,
			cluster: "cluster1",
			want:    `license_labels{license="snapmirror", cluster="cluster1"}`,
		},
		{
			name:    "selector-only query - appends inside braces",
			query:   `{__name__=~"health_.*"}`,
			cluster: "cluster1",
			want:    `{__name__=~"health_.*", cluster="cluster1"}`,
		},
		{
			name:         "clusterMatch - uses regex matcher",
			query:        "up",
			clusterMatch: "prod.*",
			want:         `up{cluster=~"prod.*"}`,
		},
		{
			name:         "clusterMatch with existing labels - appends regex matcher inside braces",
			query:        `volume_size_used_percent{svm="vs1"} > 95`,
			clusterMatch: "prod.*",
			want:         `volume_size_used_percent{svm="vs1", cluster=~"prod.*"} > 95`,
		},
		{
			name:         "clusterMatch with labels and no trailing expression - appends regex matcher inside braces",
			query:        `license_labels{license="snapmirror"}`,
			clusterMatch: "prod.*",
			want:         `license_labels{license="snapmirror", cluster=~"prod.*"}`,
		},
		{
			name:         "cluster takes precedence over clusterMatch",
			query:        "up",
			cluster:      "cluster1",
			clusterMatch: "prod.*",
			want:         `up{cluster="cluster1"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyClusterFilter(tt.query, tt.cluster, tt.clusterMatch)
			if got != tt.want {
				t.Errorf("applyClusterFilter(%q, %q, %q)\n  got  %q\n  want %q",
					tt.query, tt.cluster, tt.clusterMatch, got, tt.want)
			}
		})
	}
}
