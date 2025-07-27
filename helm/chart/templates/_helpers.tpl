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
{{- printf "%s:%s" $repository $tag -}}
{{- end -}}

{{/*
Create redpanda image
*/}}
{{- define "g2k.redpanda.image" -}}
{{- include "g2k.image" (dict "imageConfig" .Values.redpanda.image "chartAppVersion" .Chart.AppVersion "defaultTag" "latest") -}}
{{- end -}}

{{/*
Create redpanda console image
*/}}
{{- define "g2k.redpandaConsole.image" -}}
{{- include "g2k.image" (dict "imageConfig" .Values.redpandaConsole.image "chartAppVersion" .Chart.AppVersion "defaultTag" "latest") -}}
{{- end -}}

