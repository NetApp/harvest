{{/*
harvest.validations.DNS1123
Fail if the value is not a valid Kubernetes DNS-1123 label: 63 chars or fewer,
lowercase alphanumeric and '-', starting and ending alphanumeric.

Usage:
{{- include "harvest.validations.DNS1123" $<variable> -}}
*/}}
{{- define "harvest.validations.DNS1123" -}}
{{- if or (gt (len .) 63) (not (regexMatch "^[a-z0-9]([a-z0-9-]*[a-z0-9])?$" .)) -}}
{{- fail (printf "\n ERROR: invalid Kubernetes name %q: use 63 or fewer chars of lowercase a-z, 0-9 and '-', starting and ending alphanumeric. Shorten the poller name or set fullnameOverride." .) -}}
{{- end -}}
{{- end -}}

{{/*
harvest.validate
Fail fast on invalid values. Included at the top of every manifest template because Helm
renders sub-directories first, so a template may otherwise hit a nil value before a
dedicated validation file would run. Renders nothing.

Usage:
{{- include "harvest.validate" . -}}
*/}}
{{- define "harvest.validate" -}}
{{- $root := . -}}
{{- $scrape := .Values.monitoring.scrape -}}
{{- $adminSD := $scrape.adminSD.enabled -}}

{{- /* PodMonitor and admin SD would scrape every poller twice. Neither is fine: scrape the pods yourself. */ -}}
{{- if and $adminSD $scrape.podMonitor.enabled -}}
{{- fail "\n ERROR: monitoring.scrape.adminSD.enabled and monitoring.scrape.podMonitor.enabled are both true. Pick one scrape method (or disable both to scrape the pods yourself)." -}}
{{- end -}}

{{- /* Two pollers for the same cluster publish duplicate series; the admin SD registry is in memory. */ -}}
{{- if not (has (int .Values.poller.replicaCount) (list 0 1)) -}}
{{- fail "\n ERROR: poller.replicaCount must be 0 or 1. Several replicas of one poller collect and publish duplicate metrics." -}}
{{- end -}}
{{- if and $adminSD (ne (int .Values.admin.replicaCount) 1) -}}
{{- fail "\n ERROR: admin.replicaCount must be 1. The admin service-discovery registry is kept in memory, replicas would each know only some pollers." -}}
{{- end -}}

{{- /* Both scrape methods need a Prometheus exporter in the generated harvest.yml. */ -}}
{{- if and (not .Values.config.existingConfigMap) (or $adminSD $scrape.podMonitor.enabled) -}}
{{- $found := false -}}
{{- range $expName, $exp := (.Values.harvestConfig.Exporters | default dict) -}}
{{- if and (kindIs "map" $exp) (eq (toString (get $exp "exporter")) "Prometheus") -}}
{{- $found = true -}}
{{- end -}}
{{- end -}}
{{- if not $found -}}
{{- fail "\n ERROR: harvestConfig.Exporters must define an exporter with `exporter: Prometheus` when a PodMonitor or admin service discovery is used." -}}
{{- end -}}
{{- end -}}

{{- /* Admin SD */ -}}
{{- if $adminSD -}}
{{- $port := .Values.admin.containerPorts.httpsd -}}
{{- if not .Values.config.existingConfigMap -}}
{{- $adminCfg := .Values.harvestConfig.Admin | default dict -}}
{{- $listen := (($adminCfg.httpsd | default dict).listen) | default "" -}}
{{- if and $listen (not (hasSuffix (printf ":%v" $port) $listen)) -}}
{{- fail (printf "\n ERROR: harvestConfig.Admin.httpsd.listen is %q but the admin container and Service use port %v (admin.containerPorts.httpsd). Make them match or remove listen." $listen $port) -}}
{{- end -}}
{{- else -}}
{{- /* Helm can read the ConfigMap on install/upgrade (not on helm template/lint). */ -}}
{{- $cm := lookup "v1" "ConfigMap" .Release.Namespace .Values.config.existingConfigMap -}}
{{- if $cm -}}
{{- $yml := dig "data" "harvest.yml" "" $cm -}}
{{- if not (and (contains "heart_beat_url" $yml) (contains "listen" $yml)) -}}
{{- fail (printf "\n ERROR: ConfigMap %q is used with admin service discovery but its harvest.yml lacks Admin.httpsd.listen or exporters[].heart_beat_url. Add both to that harvest.yml." .Values.config.existingConfigMap) -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- /* Pollers */ -}}
{{- $slugs := dict -}}
{{- range $name, $p := (.Values.harvestConfig.Pollers | default dict) -}}
{{- if not (kindIs "map" $p) -}}
{{- fail (printf "\n ERROR: poller %q is empty or not a map. Give it Harvest poller settings such as addr, username and password." $name) -}}
{{- end -}}
{{- $slug := include "harvest.pollerSlug" $name -}}
{{- if not $slug -}}
{{- fail (printf "\n ERROR: poller name %q has no letters or digits to build a Kubernetes name from." $name) -}}
{{- end -}}
{{- if hasKey $slugs $slug -}}
{{- fail (printf "\n ERROR: pollers %q and %q both map to the Kubernetes name %q. Rename one of them." (get $slugs $slug) $name $slug) -}}
{{- end -}}
{{- $_ := set $slugs $slug $name -}}
{{- /* The chart names its own volumes harvest-config and custom-<collector>; a user volume with the same name is rejected by Kubernetes. */ -}}
{{- $reserved := list "harvest-config" -}}
{{- range $collectorType, $ext := ($p.collectorsExtensions | default dict) -}}
{{- if $ext -}}
{{- $reserved = append $reserved (printf "custom-%s" (lower $collectorType)) -}}
{{- end -}}
{{- end -}}
{{- range $v := concat ($root.Values.poller.extraVolumes | default list) ($p.extraVolumes | default list) -}}
{{- if has (get $v "name") $reserved -}}
{{- fail (printf "\n ERROR: poller %q: extraVolumes name %q is used by the chart. Use another name (reserved: %s)." $name (get $v "name") (join ", " $reserved)) -}}
{{- end -}}
{{- end -}}
{{- include "harvest.validations.DNS1123" (include "harvest.pollerFullName" (dict "root" $root "name" $name)) -}}
{{- $port := atoi (include "harvest.pollerPort" (dict "root" $root "poller" $p)) -}}
{{- if or (lt $port 1) (gt $port 65535) -}}
{{- fail (printf "\n ERROR: poller %q has an invalid prom_port. Set prom_port (or poller.defaultPromPort) to a number from 1 to 65535." $name) -}}
{{- end -}}
{{- end -}}
{{- end -}}
