{{/*
harvest.fullname
Returns the full name of the release, truncated to 63 chars.
*/}}
{{- define "harvest.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{/*
harvest.chart
Returns the chart name and version label value.
*/}}
{{- define "harvest.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
harvest.image
Constructs the image endpoint, defaults to Chart's AppVersion.
*/}}

{{- define "harvest.image" -}}
{{- $registry := .Values.image.registry -}}
{{- $repo := .Values.image.repository -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- end -}}


{{/*
harvest.labels
Standard Kubernetes labels.
Usage: include "harvest.labels" .
*/}}
{{- define "harvest.labels" -}}
helm.sh/chart: {{ include "harvest.chart" . }}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
harvest.selectorLabels
Minimal labels for use in selector/matchLabels.
Usage: include "harvest.selectorLabels" .
*/}}
{{- define "harvest.selectorLabels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
harvest.configChecksum
sha256 of harvestConfig (incl. per-poller chart keys), so pods roll when config changes.
*/}}
{{- define "harvest.configChecksum" -}}
{{- toYaml .Values.harvestConfig | sha256sum -}}
{{- end -}}

{{/*
harvest.configMapName
The config.existingConfigMap when set, otherwise the chart-generated ConfigMap.
*/}}
{{- define "harvest.configMapName" -}}
{{- if .Values.config.existingConfigMap -}}
{{- .Values.config.existingConfigMap -}}
{{- else -}}
{{- printf "%s-config" (include "harvest.fullname" .) -}}
{{- end -}}
{{- end -}}
