
{{/*
harvest.validations.pollerPort.empty
Fail if poller's prom port is empty
Usage: {{ include "harvest.validations.pollerPort.empty" (dict "pollerName" $name "poller" $p ) }}
*/}}
{{- define "harvest.validations.pollerPort.empty" -}}
{{ $_ := required (printf " \n ERROR: %s prom_port empty! Please provide a non-empty and free port for poller's prom_port key." .pollerName) .poller.prom_port  }}
{{- end -}}
