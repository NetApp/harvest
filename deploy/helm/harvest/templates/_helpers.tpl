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
Constructs the image endpoint from image.registry, image.repository and image.tag.
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
harvest.pollerSlug
Kubernetes-safe form of a poller name: lowercase, every run of characters outside
[a-z0-9-] becomes '-', and leading/trailing '-' are trimmed.
The original name is still used for --poller and in harvest.yml.
Usage: include "harvest.pollerSlug" $name
*/}}
{{- define "harvest.pollerSlug" -}}
{{- regexReplaceAll "[^a-z0-9-]+" (lower .) "-" | trimAll "-" -}}
{{- end -}}

{{/*
harvest.pollerFullName
Name shared by a poller's Deployment and PodMonitor: <fullname>-poller-<slug>
Usage: include "harvest.pollerFullName" (dict "root" $ "name" $name)
*/}}
{{- define "harvest.pollerFullName" -}}
{{- printf "%s-poller-%s" (include "harvest.fullname" .root) (include "harvest.pollerSlug" .name) -}}
{{- end -}}

{{/*
harvest.pollerPort
The poller's prom_port, falling back to poller.defaultPromPort. Returns "0" when the
value is not a number so validation can reject it.
Usage: include "harvest.pollerPort" (dict "root" $ "poller" $p)
*/}}
{{- define "harvest.pollerPort" -}}
{{- $port := .root.Values.poller.defaultPromPort -}}
{{- if hasKey .poller "prom_port" -}}
{{- $port = .poller.prom_port -}}
{{- end -}}
{{- atoi (toString $port) -}}
{{- end -}}

{{/*
harvest.pollerChartKeys is the list of poller keys that only the chart understands.
They are removed before harvest.yml is rendered. Keep in sync with the harvestConfig.Pollers comment in values.yaml.
*/}}
{{- define "harvest.pollerChartKeys" -}}
extraEnvVars collectorsExtensions podMonitor extraVolumes extraVolumeMounts
{{- end -}}

{{/*
harvest.pollerChecksum
Changes only when something that affects this poller changes: the shared checksum (every section
of harvest.yml except Pollers), this poller's own entry, and its collector extensions. Editing
one poller therefore restarts only that poller. The caller computes the effective config and the
shared checksum once. With config.existingConfigMap the file is not visible to Helm, so bump
config.checksum to roll pods.
Usage: include "harvest.pollerChecksum" (dict "root" $ "name" $name "poller" $p "cfg" $cfg "shared" $shared)
*/}}
{{- define "harvest.pollerChecksum" -}}
{{- $parts := dict "shared" .shared "extensions" (.poller.collectorsExtensions | default dict) "checksum" (.root.Values.config.checksum | default "") -}}
{{- $pollers := get .cfg "Pollers" -}}
{{- if $pollers -}}
{{- $_ := set $parts "poller" (get $pollers .name) -}}
{{- end -}}
{{- toYaml $parts | sha256sum -}}
{{- end -}}

{{/*
harvest.adminChecksum
Like harvest.pollerChecksum, for the admin pod: only the shared sections matter.
*/}}
{{- define "harvest.adminChecksum" -}}
{{- $parts := dict "checksum" (.Values.config.checksum | default "") -}}
{{- if not .Values.config.existingConfigMap -}}
{{- $cfg := include "harvest.config.effective" . | fromYaml -}}
{{- $_ := set $parts "shared" (omit $cfg "Pollers") -}}
{{- end -}}
{{- toYaml $parts | sha256sum -}}
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
