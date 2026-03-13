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
  monitoring.scrape.podMonitor.defaultInterval         (string, e.g. "30s")
  monitoring.scrape.podMonitor.defaultScrapeTimeout    (string, e.g. "20s")
  monitoring.scrape.podMonitor.labels                  (map) extra labels
  monitoring.scrape.podMonitor.annotations             (map) extra annotations

  # Per-poller overrides (optional)
  harvestConfig.Pollers.<name>.podMonitor:
    enabled: true|false
    path: /metrics
    portName: http
    interval: 30s
    scrapeTimeout: 20s
    honorLabels: false
    relabelings: []          # PodMonitor 'relabelings'
    metricRelabelings: []    # PodMonitor 'metricRelabelings'
*/}}
{{- define "harvest.poller.podmonitor" -}}
{{- $root := .root -}}
{{- $name := .pollerName -}}
{{- $p := .poller -}}
{{- $pm := (get $p "podMonitor") | default dict -}}
{{- $enabled := (coalesce (get $pm "enabled") ($root.Values.monitoring.scrape.podMonitor.enabled | default true)) -}}
{{- if $enabled }}
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: {{ include "harvest.fullname" $root }}-poller-{{ $name }}
  labels:
    {{- include "harvest.labels" $root | nindent 4 }}
    app.kubernetes.io/component: poller-{{ $name }}
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
      app.kubernetes.io/component: poller-{{ $name }}
  podMetricsEndpoints:
    - path: {{ default "/metrics" (get $pm "path") | quote }}
      port: {{ default "http" (get $pm "portName") | quote }}
      interval: {{ default ($root.Values.monitoring.scrape.podMonitor.defaultInterval | default "30s") (get $pm "interval") | quote }}
      scrapeTimeout: {{ default ($root.Values.monitoring.scrape.podMonitor.defaultScrapeTimeout | default "20s") (get $pm "scrapeTimeout") | quote }}
      {{- with (get $pm "honorLabels") }}
      honorLabels: {{ . }}
      {{- end }}
      {{- with (get $pm "relabelings") }}
      relabelings:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with (get $pm "metricRelabelings") }}
      metricRelabelings:
        {{- toYaml . | nindent 8 }}
      {{- end }}
{{- end }}
{{- end }}
