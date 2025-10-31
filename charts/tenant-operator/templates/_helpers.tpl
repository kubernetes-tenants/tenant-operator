{{/*
Expand the name of the chart.
*/}}
{{- define "tenant-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a fully qualified app name.
*/}}
{{- define "tenant-operator.fullname" -}}
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
Create chart label.
*/}}
{{- define "tenant-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "tenant-operator.labels" -}}
helm.sh/chart: {{ include "tenant-operator.chart" . }}
{{ include "tenant-operator.selectorLabels" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "tenant-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tenant-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{/*
Returns the name of the service account to use.
*/}}
{{- define "tenant-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (printf "%s-controller-manager" (include "tenant-operator.fullname" .)) .Values.serviceAccount.name -}}
{{- else -}}
{{- required "serviceAccount.name is required when serviceAccount.create is false" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}

{{/*
Base name for controller manager derived resources.
*/}}
{{- define "tenant-operator.managerName" -}}
{{- printf "%s-controller-manager" (include "tenant-operator.fullname" .) -}}
{{- end -}}

{{/*
Webhook service name.
*/}}
{{- define "tenant-operator.webhookServiceName" -}}
{{- printf "%s-webhook" (include "tenant-operator.fullname" .) -}}
{{- end -}}

{{/*
Webhook TLS secret name.
*/}}
{{- define "tenant-operator.webhookSecretName" -}}
{{- if .Values.webhook.existingSecretName -}}
{{- .Values.webhook.existingSecretName -}}
{{- else -}}
{{- printf "%s-webhook-server-cert" (include "tenant-operator.fullname" .) -}}
{{- end -}}
{{- end -}}

{{/*
Webhook certificate resource name.
*/}}
{{- define "tenant-operator.webhookCertificateName" -}}
{{- printf "%s-serving-cert" (include "tenant-operator.fullname" .) -}}
{{- end -}}

{{/*
Webhook issuer name (handles auto-generated issuer).
*/}}
{{- define "tenant-operator.webhookIssuerName" -}}
{{- if and .Values.webhook.certManager.enabled .Values.webhook.certManager.issuer.create -}}
{{- printf "%s-%s" (include "tenant-operator.fullname" .) .Values.webhook.certManager.issuer.name -}}
{{- else -}}
{{- .Values.webhook.certManager.issuer.name -}}
{{- end -}}
{{- end -}}
