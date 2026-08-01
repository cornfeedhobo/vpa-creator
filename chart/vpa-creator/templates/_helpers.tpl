{{/*
Expand the name of the chart.
*/}}
{{- define "vpa-creator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "vpa-creator.fullname" -}}
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

{{/*
Resolve the namespace for namespaced resources.
*/}}
{{- define "vpa-creator.namespace" -}}
{{- default .Release.Namespace .Values.namespace.name -}}
{{- end -}}

{{/*
Create chart labels.
*/}}
{{- define "vpa-creator.labels" -}}
helm.sh/chart: {{ include "vpa-creator.chart" . }}
{{ include "vpa-creator.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "vpa-creator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Selector labels.
*/}}
{{- define "vpa-creator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "vpa-creator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
control-plane: controller-manager
{{- end -}}

{{/*
Service account name.
*/}}
{{- define "vpa-creator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (printf "%s-controller-manager" (include "vpa-creator.fullname" .)) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{- define "vpa-creator.managerName" -}}
{{- printf "%s-controller-manager" (include "vpa-creator.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "vpa-creator.metricsServiceName" -}}
{{- printf "%s-controller-manager-metrics-service" (include "vpa-creator.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "vpa-creator.managerRoleName" -}}
{{- printf "%s-manager-role" (include "vpa-creator.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "vpa-creator.leaderElectionRoleName" -}}
{{- printf "%s-leader-election-role" (include "vpa-creator.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "vpa-creator.metricsAuthRoleName" -}}
{{- printf "%s-metrics-auth-role" (include "vpa-creator.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "vpa-creator.metricsReaderRoleName" -}}
{{- printf "%s-metrics-reader" (include "vpa-creator.fullname" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}
