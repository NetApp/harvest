{{/* templates/_deployment.tpl */}}

{{/*
harvest.poller.deployment
-------------------------
Renders a Deployment for a single Harvest poller.

Inputs (dict):
  .root       : chart root context (.)
  .pollerName : key under harvestConfig.Pollers (e.g., "cluster-a")
  .poller     : the poller's config map — Harvest fields plus chart-only keys
                (extraEnvVars, collectorsExtensions, podMonitor) stripped from harvest.yml

Notes:
  - replicas default to 1 to avoid duplicate metrics (identical exporters).
  - collectorsExtensions.<CollectorType>.objects mounts a per-collector custom.yaml
    at /opt/harvest/conf/<collectortype-lower>/custom.yaml.
  - the config checksum annotation triggers rollout when config changes.
*/}}
{{- define "harvest.poller.deployment" -}}
{{- $root := .root -}}
{{- $name := .pollerName -}}
{{- $p := .poller -}}
{{- $exts := ($p.collectorsExtensions | default dict) -}}

apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "harvest.fullname" $root }}-poller-{{ $name }}
  labels:
    {{- include "harvest.labels" $root | nindent 4 }}
    app.kubernetes.io/component: poller-{{ $name }}
spec:
  replicas: {{ $root.Values.poller.replicaCount }}
  {{- if $root.Values.poller.updateStrategy }}
  strategy: {{- toYaml $root.Values.poller.updateStrategy | nindent 4 }}
  {{- end }}
  selector:
    matchLabels:
      {{- include "harvest.selectorLabels" $root | nindent 6 }}
      app.kubernetes.io/component: poller-{{ $name }}
  template:
    metadata:
      annotations:
        {{- with $root.Values.poller.podAnnotations }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
        checksum/config: {{ include "harvest.configChecksum" $root }}
      labels:
        {{- include "harvest.labels" $root | nindent 8 }}
        app.kubernetes.io/component: poller-{{ $name }}
    spec:
      automountServiceAccountToken: {{ $root.Values.poller.automountServiceAccountToken }}
      {{- with $root.Values.image.pullSecrets }}
      imagePullSecrets:
        {{- range . }}
        - name: {{ . }}
        {{- end }}
      {{- end }}
      {{- with $root.Values.poller.affinity }}
      affinity: {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $root.Values.poller.nodeSelector }}
      nodeSelector: {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- with $root.Values.poller.tolerations }}
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
      {{- with $root.Values.poller.terminationGracePeriodSeconds }}
      terminationGracePeriodSeconds: {{ . }}
      {{- end }}
      containers:
        - name: poller-{{ $name }}
          image: {{ include "harvest.image" $root }}
          imagePullPolicy: {{ $root.Values.image.pullPolicy }}
          {{- with $root.Values.poller.containerSecurityContext }}
          securityContext: {{- toYaml . | nindent 12 }}
          {{- end }}
          args:
            - --poller
            - "{{ $name }}"
            - --promPort
            - "{{ $p.prom_port }}"
            - --config
            - /opt/harvest/harvest.yml
          {{- /* global env first, then this poller's own env so it wins on conflicts */ -}}
          {{- $envVars := concat ($root.Values.poller.extraEnvVars | default list) ($p.extraEnvVars | default list) }}
          {{- with $envVars }}
          env:
            {{- toYaml . | nindent 12 }}
          {{- end }}

          {{- with $root.Values.poller.resources }}
          resources: {{- toYaml . | nindent 12 }}
          {{- end }}
          ports:
            - name: http
              containerPort: {{ $p.prom_port }}
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

            {{- /* For each collectorType that defines 'objects', mount a custom.yaml
                  that augments the built-in default.yaml for that collector.
                  Example: Rest     -> /opt/harvest/conf/rest/custom.yaml
                           RestPerf -> /opt/harvest/conf/restperf/custom.yaml
            */ -}}
            {{- range $collectorType, $ext := $exts }}
            {{- if (get $ext "objects") }}
            - name: {{ printf "custom-%s-%s" $name (lower $collectorType) }}
              mountPath: {{ printf "/opt/harvest/conf/%s/custom.yaml" (lower $collectorType) }}
              subPath: custom.yaml
            {{- end }}
            {{- end }}
      volumes:
        - name: harvest-config
          configMap:
            name: {{ include "harvest.configMapName" $root }}

        {{- /* One ConfigMap per collectorType that has 'objects', generated by a separate template:
              <release>-poller-<pollerName>-<collectorType-lower>-custom
              Each contains a 'custom.yaml' with:
                objects:
                  <ObjectName>:
                    - <object-template.yaml>
        */ -}}
        {{- range $collectorType, $ext := $exts }}
        {{- if (get $ext "objects") }}
        - name: {{ printf "custom-%s-%s" $name (lower $collectorType) }}
          configMap:
            name: {{ include "harvest.fullname" $root }}-poller-{{ $name }}-{{ lower $collectorType }}-custom
        {{- end }}
        {{- end }}
{{- end -}}
