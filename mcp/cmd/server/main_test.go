package main

import "testing"

func TestInjectLabelSelector(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		selector string
		want     string
	}{
		{
			name:     "empty selector - returns query unchanged",
			query:    "up",
			selector: "",
			want:     "up",
		},
		{
			name:     "bare metric - wraps with selector",
			query:    "up",
			selector: `job="harvest"`,
			want:     `up{job="harvest"}`,
		},
		{
			name:     "metric with expression - inserts before space",
			query:    "volume_size_used_percent > 95",
			selector: `job="harvest"`,
			want:     `volume_size_used_percent{job="harvest"} > 95`,
		},
		{
			name:     "metric with existing labels - appends inside braces",
			query:    `volume_size_used_percent{svm="vs1"} > 95`,
			selector: `job="harvest"`,
			want:     `volume_size_used_percent{svm="vs1",job="harvest"} > 95`,
		},
		{
			name:     "selector-only query - appends inside braces",
			query:    `{__name__=~"health_.*"}`,
			selector: `job="harvest"`,
			want:     `{__name__=~"health_.*",job="harvest"}`,
		},
		{
			name:     "metric with labels no trailing expression",
			query:    `license_labels{license="snapmirror"}`,
			selector: `job="harvest"`,
			want:     `license_labels{license="snapmirror",job="harvest"}`,
		},
		{
			name:     "regex selector",
			query:    "node_uptime",
			selector: `job=~"harvest|storagegrid"`,
			want:     `node_uptime{job=~"harvest|storagegrid"}`,
		},
		{
			name:     "multiple label selectors",
			query:    "volume_read_ops > 1000",
			selector: `job="harvest", datacenter="us-east"`,
			want:     `volume_read_ops{job="harvest",datacenter="us-east"} > 1000`,
		},
		{
			name:     "aggregation by clause - injects into inner VectorSelector",
			query:    "count by (cluster) (cluster_new_status)",
			selector: `cluster="c1"`,
			want:     `count(cluster_new_status{cluster="c1"}) by(cluster)`,
		},
		{
			name:     "nested function call - injects into inner VectorSelector",
			query:    "sum(rate(http_requests_total[5m]))",
			selector: `job="harvest"`,
			want:     `sum(rate(http_requests_total{job="harvest"}[5m]))`,
		},
		{
			name:     "topk aggregation - injects into inner VectorSelector",
			query:    "topk(5, volume_size_used_percent)",
			selector: `job="harvest"`,
			want:     `topk(5, volume_size_used_percent{job="harvest"})`,
		},
		{
			name:     "count_values aggregation - injects into inner VectorSelector",
			query:    `count_values("version", build_info)`,
			selector: `job="harvest"`,
			want:     `count_values("version", build_info{job="harvest"})`,
		},
		{
			name:     "empty selector query - injects into bare braces (ListMetrics fallback)",
			query:    `{}`,
			selector: `job="harvest"`,
			want:     `{job="harvest"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := injectFilters(tt.query, parseSelector(tt.selector))
			if got != tt.want {
				t.Errorf("injectFilters(%q, parseSelector(%q))\n  got  %q\n  want %q",
					tt.query, tt.selector, got, tt.want)
			}
		})
	}
}

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
			name:    "cluster - builds exact matcher",
			query:   "up",
			cluster: "cluster1",
			want:    `up{cluster="cluster1"}`,
		},
		{
			name:         "clusterMatch - builds regex matcher",
			query:        "up",
			clusterMatch: "prod.*",
			want:         `up{cluster=~"prod.*"}`,
		},
		{
			name:         "cluster takes precedence over clusterMatch",
			query:        "up",
			cluster:      "cluster1",
			clusterMatch: "prod.*",
			want:         `up{cluster="cluster1"}`,
		},
		{
			name:    "delegates injection - metric with existing labels",
			query:   `volume_size_used_percent{svm="vs1"} > 95`,
			cluster: "cluster1",
			want:    `volume_size_used_percent{svm="vs1",cluster="cluster1"} > 95`,
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

func TestCombinedFilters(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		labelFilter  string
		cluster      string
		clusterMatch string
		want         string
	}{
		{
			name:        "label filter only - no cluster",
			query:       "up",
			labelFilter: `job="harvest"`,
			want:        `up{job="harvest"}`,
		},
		{
			name:    "cluster only - no label filter",
			query:   "up",
			cluster: "cluster1",
			want:    `up{cluster="cluster1"}`,
		},
		{
			name:        "both on bare metric",
			query:       "up",
			labelFilter: `job="harvest"`,
			cluster:     "cluster1",
			want:        `up{job="harvest",cluster="cluster1"}`,
		},
		{
			name:        "both on metric with existing labels",
			query:       `volume_size_used_percent{svm="vs1"} > 95`,
			labelFilter: `job="harvest"`,
			cluster:     "cluster1",
			want:        `volume_size_used_percent{svm="vs1",job="harvest",cluster="cluster1"} > 95`,
		},
		{
			name:        "both on selector-only query (health alerts pattern)",
			query:       `{__name__=~"health_.*"}`,
			labelFilter: `job="harvest"`,
			cluster:     "cluster1",
			want:        `{__name__=~"health_.*",job="harvest",cluster="cluster1"}`,
		},
		{
			name:         "label filter + clusterMatch regex",
			query:        "node_uptime",
			labelFilter:  `job="harvest"`,
			clusterMatch: "prod.*",
			want:         `node_uptime{job="harvest",cluster=~"prod.*"}`,
		},
		{
			name:        "multiple label selectors + cluster",
			query:       "volume_read_ops > 1000",
			labelFilter: `job="harvest", datacenter="us-east"`,
			cluster:     "cluster1",
			want:        `volume_read_ops{job="harvest",datacenter="us-east",cluster="cluster1"} > 1000`,
		},
		{
			name:  "no filters - passthrough",
			query: "volume_size_used_percent > 95",
			want:  "volume_size_used_percent > 95",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := injectFilters(tt.query, parseSelector(tt.labelFilter))
			got := applyClusterFilter(q, tt.cluster, tt.clusterMatch)
			if got != tt.want {
				t.Errorf("combined(%q, labelFilter=%q, cluster=%q, clusterMatch=%q)\n  got  %q\n  want %q",
					tt.query, tt.labelFilter, tt.cluster, tt.clusterMatch, got, tt.want)
			}
		})
	}
}
