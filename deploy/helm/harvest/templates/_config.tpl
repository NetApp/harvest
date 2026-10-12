{{/*
harvest.adminSDURL
URL of the admin service-discovery endpoint inside the cluster. Used by pollers as
heart_beat_url and by the ScrapeConfig. https when Admin.httpsd.tls.cert_file is set.
*/}}
{{- define "harvest.adminSDURL" -}}
{{- $scheme := "http" -}}
{{- $adminCfg := .Values.harvestConfig.Admin | default dict -}}
{{- $tls := (($adminCfg.httpsd | default dict).tls) | default dict -}}
{{- if $tls.cert_file -}}
{{- $scheme = "https" -}}
{{- end -}}
{{- printf "%s://%s-admin-sd.%s.svc:%v/api/v1/sd" $scheme (include "harvest.fullname" .) .Release.Namespace .Values.admin.service.ports.httpsd -}}
{{- end -}}

{{/*
harvest.config.effective
------------------------
The harvest.yml that pollers and the admin pod read, as YAML. It is harvestConfig with:
  - chart-only poller keys removed, so they never reach Harvest
  - when admin service discovery is on:
      Admin.httpsd.listen set to :<admin.containerPorts.httpsd> unless the user set it
        (`harvest admin start` exits when listen is empty)
      heart_beat_url set on every exporter of type Prometheus that has none, because a
        bare ":port" listen address makes pollers publish to 127.0.0.1 (themselves)
        instead of the admin Service.
Usage: include "harvest.config.effective" .
*/}}
{{- define "harvest.config.effective" -}}
{{- $cfg := deepCopy .Values.harvestConfig -}}
{{- if .Values.monitoring.scrape.adminSD.enabled -}}
{{- $admin := $cfg.Admin | default dict -}}
{{- $httpsd := $admin.httpsd | default dict -}}
{{- if not $httpsd.listen -}}
{{- $_ := set $httpsd "listen" (printf ":%v" .Values.admin.containerPorts.httpsd) -}}
{{- end -}}
{{- $_ := set $admin "httpsd" $httpsd -}}
{{- $_ := set $cfg "Admin" $admin -}}
{{- $sdURL := include "harvest.adminSDURL" . -}}
{{- range $expName, $exp := ($cfg.Exporters | default dict) -}}
{{- if and (kindIs "map" $exp) (eq (toString (get $exp "exporter")) "Prometheus") (not (get $exp "heart_beat_url")) -}}
{{- $_ := set $exp "heart_beat_url" $sdURL -}}
{{- end -}}
{{- end -}}
{{- end -}}
{{- $keys := splitList " " (include "harvest.pollerChartKeys" .) -}}
{{- range $name, $p := ($cfg.Pollers | default dict) -}}
{{- range $key := $keys -}}
{{- $_ := unset $p $key -}}
{{- end -}}
{{- end -}}
{{- toYaml $cfg -}}
{{- end -}}
