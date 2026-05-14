{{- define "shortcut.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "shortcut.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "shortcut.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "shortcut.labels" -}}
helm.sh/chart: {{ include "shortcut.chart" . }}
{{ include "shortcut.selectorLabels" . }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "shortcut.selectorLabels" -}}
app.kubernetes.io/name: {{ include "shortcut.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "shortcut.postgresHost" -}}
{{- if .Values.postgresql.enabled -}}
{{- printf "%s-postgresql" .Release.Name -}}
{{- else -}}
{{- .Values.shortcut.externalDatabases.postgres.host -}}
{{- end -}}
{{- end -}}

{{- define "shortcut.mongoHost" -}}
{{- if .Values.mongodb.enabled -}}
{{- printf "%s-mongodb" .Release.Name -}}
{{- else -}}
{{- .Values.shortcut.externalDatabases.mongo.host -}}
{{- end -}}
{{- end -}}

{{- define "shortcut.valkeyHost" -}}
{{- if .Values.valkey.enabled -}}
{{- printf "%s-valkey-primary" .Release.Name -}}
{{- else -}}
{{- .Values.shortcut.externalDatabases.valkey.host -}}
{{- end -}}
{{- end -}}
