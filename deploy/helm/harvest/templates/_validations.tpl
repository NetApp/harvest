
{{/*
harvest.validations.pollerPort.empty
Fail if poller's prom port is empty

Usage: 
{{ include "harvest.validations.pollerPort.empty" (dict "pollerName" $name "poller" $p ) }}

*/}}
{{- define "harvest.validations.pollerPort.empty" -}}
{{ $_ := required (printf " \n ERROR: %s prom_port empty! Please provide a non-empty and free port for poller's prom_port key." .pollerName) .poller.prom_port  }}
{{- end -}}


{{/*
harvest.validateDNS1123
Fail if the value is not a valid Kubernetes DNS-1123 label: 63 chars or fewer,
lowercase alphanumeric and '-', starting and ending alphanumeric.

Usage:
{{- include "harvest.validations.DNS1123" $<variable> -}}

*/}}
{{- define "harvest.validations.DNS1123" -}}
{{- if or (gt (len .) 63) (not (regexMatch "^[a-z0-9]([a-z0-9-]*[a-z0-9])?$" .)) -}}
{{- fail (printf "invalid Kubernetes name %q: use 63 or fewer chars of lowercase a-z, 0-9 and '-', starting and ending alphanumeric" .) -}}
{{- end -}}
{{- end -}}
