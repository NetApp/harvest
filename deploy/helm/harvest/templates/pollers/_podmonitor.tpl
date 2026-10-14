{{/*
harvest.poller.podmonitor
-------------------------
Renders a PodMonitor for a single Harvest poller.

Inputs (dict):
  .root       : chart root context (.)
  .pollerName : key under harvestConfig.Pollers (e.g., "cluster-a")
  .poller     : the values map for that poller

Values:
  monitoring.scrape.podMonitor.enabled                 (bool, default: true)
  monitoring.scrape.podMonitor.defaultInterval         (string, e.g. "1m")
  monitoring.scrape.podMonitor.defaultScrapeTimeout    (string, e.g. "50s")
  monitoring.scrape.podMonitor.labels                  (map) extra labels
  monitoring.scrape.podMonitor.annotations             (map) extra annotations

  # Per-poller overrides (optional)
  harvestConfig.Pollers.<name>.podMonitor:
    enabled: false           # skip this poller; cannot turn a PodMonitor on when
                             # monitoring.scrape.podMonitor.enabled is false
    path: /metrics
    portName: http
    interval: 1m
    scrapeTimeout: 50s
    honorLabels: false
    relabelings: []          # extra PodMonitor 'relabelings', added after the chart's instance relabeling
    metricRelabelings: []    # PodMonitor 'metricRelabelings'
*/}}
{{- define "harvest.poller.podmonitor" -}}
{{- $root := .root -}}
{{- $name := .pollerName -}}
{{- $p := .poller -}}
{{- $slug := include "harvest.pollerSlug" $name -}}
{{- $pm := (get $p "podMonitor") | default dict -}}
{{- $globalEnabled := true -}}
{{- if hasKey $root.Values.monitoring.scrape.podMonitor "enabled" -}}
{{- $globalEnabled = $root.Values.monitoring.scrape.podMonitor.enabled -}}
{{- end -}}
{{- $enabled := $globalEnabled -}}
{{- if hasKey $pm "enabled" -}}
{{- $enabled = (get $pm "enabled") -}}
{{- end -}}
{{- if $enabled }}
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: {{ include "harvest.pollerFullName" (dict "root" $root "name" $name) }}
  labels:
    {{- include "harvest.labels" $root | nindent 4 }}
    app.kubernetes.io/component: poller-{{ $slug }}
    {{- with $root.Values.monitoring.scrape.podMonitor.labels }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
  {{- with $root.Values.monitoring.scrape.podMonitor.annotations }}
  annotations: {{- toYaml . | nindent 4 }}
  {{- end }}
spec:
  # Scrape pods in THIS namespace (where the pollers run)
  namespaceSelector:
    matchNames:
      - {{ $root.Release.Namespace }}
  selector:
    matchLabels:
      {{- include "harvest.selectorLabels" $root | nindent 6 }}
      app.kubernetes.io/component: poller-{{ $slug }}
  podMetricsEndpoints:
    - path: {{ default "/metrics" (get $pm "path") | quote }}
      port: {{ default "http" (get $pm "portName") | quote }}
      interval: {{ default ($root.Values.monitoring.scrape.podMonitor.defaultInterval | default "1m") (get $pm "interval") | quote }}
      scrapeTimeout: {{ default ($root.Values.monitoring.scrape.podMonitor.defaultScrapeTimeout | default "50s") (get $pm "scrapeTimeout") | quote }}
      {{- if hasKey $pm "honorLabels" }}
      honorLabels: {{ get $pm "honorLabels" }}
      {{- end }}
      relabelings:
        # The instance label is the poller name, so it stays the same when the pod IP changes.
        # Relabelings below this one run after it and can change instance again.
        - sourceLabels: [__meta_kubernetes_pod_annotation_harvest_netapp_io_poller]
          targetLabel: instance
      {{- with (get $pm "relabelings") }}
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with (get $pm "metricRelabelings") }}
      metricRelabelings:
        {{- toYaml . | nindent 8 }}
      {{- end }}
{{- end }}
{{- end }}
