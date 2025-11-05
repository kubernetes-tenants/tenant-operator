{{/*
Expand the name of the chart.
*/}}
{{- define "tenant-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "tenant-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "tenant-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "tenant-operator.labels" -}}
helm.sh/chart: {{ include "tenant-operator.chart" . }}
{{ include "tenant-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.commonLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "tenant-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tenant-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
control-plane: controller-manager
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "tenant-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "tenant-operator.fullname" . | printf "%s-controller-manager") .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Create the namespace name
*/}}
{{- define "tenant-operator.namespace" -}}
{{- if .Values.namespace.name }}
{{- .Values.namespace.name }}
{{- else }}
{{- .Release.Namespace }}
{{- end }}
{{- end }}

{{/*
Create the image name
*/}}
{{- define "tenant-operator.image" -}}
{{- $registry := .Values.image.registry }}
{{- $repository := .Values.image.repository }}
{{- $tag := .Values.image.tag | default .Chart.AppVersion }}
{{- printf "%s/%s:%s" $registry $repository $tag }}
{{- end }}

{{/*
Common annotations
*/}}
{{- define "tenant-operator.annotations" -}}
{{- with .Values.commonAnnotations }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Manager container name
*/}}
{{- define "tenant-operator.managerContainerName" -}}
manager
{{- end }}

{{/*
Metrics service name
*/}}
{{- define "tenant-operator.metricsServiceName" -}}
{{- include "tenant-operator.fullname" . }}-metrics-service
{{- end }}

{{/*
Webhook service name
*/}}
{{- define "tenant-operator.webhookServiceName" -}}
{{- include "tenant-operator.fullname" . }}-webhook-service
{{- end }}

{{/*
Webhook certificate name
*/}}
{{- define "tenant-operator.webhookCertificateName" -}}
{{- include "tenant-operator.fullname" . }}-serving-cert
{{- end }}

{{/*
Leader election role name
*/}}
{{- define "tenant-operator.leaderElectionRoleName" -}}
{{- include "tenant-operator.fullname" . }}-leader-election-role
{{- end }}

{{/*
Manager role name
*/}}
{{- define "tenant-operator.managerRoleName" -}}
{{- include "tenant-operator.fullname" . }}-manager-role
{{- end }}

{{/*
Metrics reader role name
*/}}
{{- define "tenant-operator.metricsReaderRoleName" -}}
{{- include "tenant-operator.fullname" . }}-metrics-reader
{{- end }}

{{/*
Proxy role name
*/}}
{{- define "tenant-operator.proxyRoleName" -}}
{{- include "tenant-operator.fullname" . }}-proxy-role
{{- end }}

{{/*
Manager role binding name
*/}}
{{- define "tenant-operator.managerRoleBindingName" -}}
{{- include "tenant-operator.fullname" . }}-manager-rolebinding
{{- end }}

{{/*
Leader election role binding name
*/}}
{{- define "tenant-operator.leaderElectionRoleBindingName" -}}
{{- include "tenant-operator.fullname" . }}-leader-election-rolebinding
{{- end }}

{{/*
Proxy role binding name
*/}}
{{- define "tenant-operator.proxyRoleBindingName" -}}
{{- include "tenant-operator.fullname" . }}-proxy-rolebinding
{{- end }}

{{/*
ValidatingWebhookConfiguration name
*/}}
{{- define "tenant-operator.validatingWebhookName" -}}
{{- include "tenant-operator.fullname" . }}-validating-webhook-configuration
{{- end }}

{{/*
MutatingWebhookConfiguration name
*/}}
{{- define "tenant-operator.mutatingWebhookName" -}}
{{- include "tenant-operator.fullname" . }}-mutating-webhook-configuration
{{- end }}

{{/*
ServiceMonitor name
*/}}
{{- define "tenant-operator.serviceMonitorName" -}}
{{- include "tenant-operator.fullname" . }}-metrics-monitor
{{- end }}

{{/*
Validate webhook configuration
*/}}
{{- define "tenant-operator.validateWebhook" -}}
{{- if and .Values.webhook.enabled (not .Values.certManager.enabled) }}
{{- fail "webhook.enabled requires certManager.enabled to be true" }}
{{- end }}
{{- end }}
