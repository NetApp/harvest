{{/* templates/_deployment.tpl */}}

{{/*
harvest.poller.deployment
-------------------------
Renders a Deployment for a single Harvest poller.

Inputs (dict):
  .root       : chart root context (.)
  .pollerName : key under harvestConfig.Pollers (e.g., "cluster-a")
  .poller     : the poller's config map — Harvest fields plus chart-only keys
                (extraEnvVars, collectorsExtensions, podMonitor,
                extraVolumes, extraVolumeMounts) stripped from harvest.yml
  .cfg        : the effective harvest config, computed once; empty with config.existingConfigMap
  .shared     : checksum of every section of that config except Pollers

Notes:
  - Kubernetes names use the poller slug (lowercase, DNS-1123 safe); --poller keeps the
    original name so it matches harvest.yml.
  - replicas default to 1 to avoid duplicate metrics (identical exporters).
  - collectorsExtensions.<CollectorType> mounts a per-collector custom.yaml
    at /opt/harvest/conf/<collectortype-lower>/custom.yaml.
  - the per-poller config checksum annotation rolls only this poller when its config changes.
  - resources, nodeSelector, tolerations, affinity and podAnnotations come from poller.* and
    apply to every poller.
*/}}
{{- define "harvest.poller.deployment" -}}
{{- $root := .root -}}
{{- $name := .pollerName -}}
{{- $p := .poller -}}
{{- $slug := include "harvest.pollerSlug" $name -}}
{{- $fullName := include "harvest.pollerFullName" (dict "root" $root "name" $name) -}}
{{- $port := include "harvest.pollerPort" (dict "root" $root "poller" $p) -}}
{{- $exts := ($p.collectorsExtensions | default dict) -}}
{{- $resources := $root.Values.poller.resources -}}
{{- $nodeSelector := $root.Values.poller.nodeSelector -}}
{{- $tolerations := $root.Values.poller.tolerations -}}
{{- $affinity := $root.Values.poller.affinity -}}
{{- /* harvest.netapp.io/poller is the original poller name. The PodMonitor uses it as the Prometheus instance label. The chart's keys win over poller.podAnnotations. */ -}}
{{- $ann := deepCopy ($root.Values.poller.podAnnotations | default dict) -}}
{{- $_ := set $ann "checksum/config" (include "harvest.pollerChecksum" (dict "root" $root "name" $name "poller" $p "cfg" .cfg "shared" .shared)) -}}
{{- $_ := set $ann "harvest.netapp.io/poller" $name -}}
{{- $extraVolumes := concat ($root.Values.poller.extraVolumes | default list) ($p.extraVolumes | default list) -}}
{{- $extraMounts := concat ($root.Values.poller.extraVolumeMounts | default list) ($p.extraVolumeMounts | default list) -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $fullName }}
  labels:
    {{- include "harvest.labels" $root | nindent 4 }}
    app.kubernetes.io/component: poller-{{ $slug }}
spec:
  replicas: {{ $root.Values.poller.replicaCount }}
  {{- if $root.Values.poller.updateStrategy }}
  strategy: {{- toYaml $root.Values.poller.updateStrategy | nindent 4 }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "harvest.selectorLabels" $root | nindent 6 }}
      app.kubernetes.io/component: poller-{{ $slug }}
  template:
    metadata:
      annotations:
        {{- toYaml $ann | nindent 8 }}
      labels:
        {{- include "harvest.labels" $root | nindent 8 }}
        app.kubernetes.io/component: poller-{{ $slug }}
    spec:
      automountServiceAccountToken: {{ $root.Values.poller.automountServiceAccountToken }}
      {{- with $root.Values.image.pullSecrets }}
      imagePullSecrets:
        {{- range . }}
        - name: {{ . }}
        {{- end }}
      {{- end }}
      {{- with $affinity }}
      affinity: {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $nodeSelector }}
      nodeSelector: {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $tolerations }}
      tolerations: {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $root.Values.poller.priorityClassName }}
      priorityClassName: {{ . | quote }}
      {{- end }}
      {{- with $root.Values.poller.topologySpreadConstraints }}
      topologySpreadConstraints: {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $root.Values.poller.podSecurityContext }}
      securityContext: {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- $tgps := $root.Values.poller.terminationGracePeriodSeconds -}}
      {{- if not (kindIs "invalid" $tgps) }}
      terminationGracePeriodSeconds: {{ $tgps }}
      {{- end }}
      containers:
        - name: poller-{{ $slug }}
          image: {{ include "harvest.image" $root }}
          imagePullPolicy: {{ $root.Values.image.pullPolicy }}
          {{- with $root.Values.poller.containerSecurityContext }}
          securityContext: {{- toYaml . | nindent 12 }}
          {{- end }}
          args:
            - --poller
            - {{ $name | quote }}
            - --promPort
            - {{ $port | quote }}
            - --config
            - /opt/harvest/harvest.yml
          {{- /* global env first, then this poller's own env so it wins on conflicts */ -}}
          {{- $envVars := concat ($root.Values.poller.extraEnvVars | default list) ($p.extraEnvVars | default list) }}
          {{- with $envVars }}
          env:
            {{- toYaml . | nindent 12 }}
          {{- end }}

          {{- with $resources }}
          resources: {{- toYaml . | nindent 12 }}
          {{- end }}
          ports:
            - name: http
              containerPort: {{ $port }}
              protocol: TCP

          {{- if $root.Values.poller.livenessProbe.enabled }}
          livenessProbe: {{- toYaml (omit $root.Values.poller.livenessProbe "enabled") | nindent 12 }}
          {{- end }}
          {{- if $root.Values.poller.readinessProbe.enabled }}
          readinessProbe: {{- toYaml (omit $root.Values.poller.readinessProbe "enabled") | nindent 12 }}
          {{- end }}
          {{- if $root.Values.poller.startupProbe.enabled }}
          startupProbe: {{- toYaml (omit $root.Values.poller.startupProbe "enabled") | nindent 12 }}
          {{- end }}
          {{- with $root.Values.poller.lifecycleHooks }}
          lifecycle: {{- toYaml . | nindent 12 }}
          {{- end }}
          volumeMounts:
            - name: harvest-config
              mountPath: /opt/harvest/harvest.yml
              subPath: harvest.yml
            {{- /* For each collectorType that defines collectorsExtensions, mount a custom.yaml
                  that augments the built-in default.yaml for that collector.
                  Example: Rest     -> /opt/harvest/conf/rest/custom.yaml
                           RestPerf -> /opt/harvest/conf/restperf/custom.yaml
            */ -}}
            {{- range $collectorType, $ext := $exts }}
            {{- if $ext }}
            - name: {{ printf "custom-%s" (lower $collectorType) }}
              mountPath: {{ printf "/opt/harvest/conf/%s/custom.yaml" (lower $collectorType) }}
              subPath: custom.yaml
            {{- end }}
            {{- end }}
            {{- with $extraMounts }}
            {{- toYaml . | nindent 12 }}
            {{- end }}
      volumes:
        - name: harvest-config
          configMap:
            name: {{ include "harvest.configMapName" $root }}
        {{- /* One ConfigMap per collectorType with a non-empty extension, generated by a
              separate template: <release>-poller-<slug>-<collectorType-lower>-custom
              Its custom.yaml augments the collector's default.yaml (objects, schedule, jitter, etc.).
        */ -}}
        {{- range $collectorType, $ext := $exts }}
        {{- if $ext }}
        - name: {{ printf "custom-%s" (lower $collectorType) }}
          configMap:
            name: {{ $fullName }}-{{ lower $collectorType }}-custom
        {{- end }}
        {{- end }}
        {{- with $extraVolumes }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
{{- end -}}
