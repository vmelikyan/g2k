{{- define "g2k.name" -}}
{{- .Chart.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "g2k.fullname" -}}
{{- include "g2k.name" . }}-{{ .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create image name with tag
*/}}
{{- define "g2k.image" -}}
{{- $imageConfig := .imageConfig -}}
{{- $chartAppVersion := .chartAppVersion -}}
{{- $defaultTag := .defaultTag | default $chartAppVersion -}}
{{- $tag := $imageConfig.tag | default $defaultTag -}}
{{- if and (not $imageConfig.tag) (not (hasPrefix "v" $tag)) (not (eq $tag "latest")) -}}
{{- $tag = printf "v%s" $tag -}}
{{- end -}}
{{- printf "%s:%s" $imageConfig.repository $tag -}}
{{- end -}}

{{/*
Create g2krelay image
*/}}
{{- define "g2k.g2krelay.image" -}}
{{- include "g2k.image" (dict "imageConfig" .Values.g2krelay.image "chartAppVersion" .Chart.AppVersion) -}}
{{- end -}}

{{/*
Create g2krepeater image
*/}}
{{- define "g2k.g2krepeater.image" -}}
{{- $imageConfig := .imageConfig | default dict -}}
{{- $repository := $imageConfig.repository | default "vmelikyan/g2krepeater" -}}
{{- $tag := $imageConfig.tag | default .chartAppVersion -}}
{{- if and (not $imageConfig.tag) (not (hasPrefix "v" $tag)) (not (eq $tag "latest")) -}}
{{- $tag = printf "v%s" $tag -}}
{{- end -}}
{{- printf "%s:%s" $repository $tag -}}
{{- end -}}


{{/*
Get Kafka broker address - either from Redpanda subchart or external brokers
*/}}
{{- define "g2k.kafkaBrokers" -}}
{{- if .Values.global.kafka.enabled -}}
{{- printf "%s.%s.svc.cluster.local:9093" .Release.Name .Release.Namespace -}}
{{- else if .Values.global.kafka.externalBrokers -}}
{{- .Values.global.kafka.externalBrokers -}}
{{- else -}}
{{- fail "Either global.kafka.enabled must be true or global.kafka.externalBrokers must be specified" -}}
{{- end -}}
{{- end -}}

