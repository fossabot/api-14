{{/* vim: set filetype=mustache: */}}

{{/* Expand the name of the chart. */}}
{{- define "api.name" -}}
	{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this
(by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "api.fullname" -}}
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

{{/* Create chart name and version as used by the chart label. */}}
{{- define "api.chart" -}}
	{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Create the name for the Varnish cache proxy. */}}
{{- define "varnish.name" -}}
	{{- $name := include "api.name" . -}}
	{{- printf "%s-varnish" $name -}}
{{- end -}}
{{- define "varnish.fullname" -}}
	{{- $fullName := include "api.fullname" . -}}
	{{- printf "%s-varnish" $fullName -}}
{{- end -}}

{{/* Create the name for the frontend components. */}}
{{- define "frontend.name" -}}
	{{- $name := include "api.name" . -}}
	{{- printf "%s-frontend" $name -}}
{{- end -}}
{{- define "frontend.fullname" -}}
	{{- $fullName := include "api.fullname" . -}}
	{{- printf "%s-frontend" $fullName -}}
{{- end -}}

{{/* Name jobserver components. */}}
{{- define "jobs.name" -}}
  {{- $name := include "api.name" . -}}
  {{- printf "%s-jobs" $name -}}
{{- end -}}
{{- define "jobs.fullname" -}}
  {{- $fullName := include "api.fullname" . -}}
  {{- printf "%s-jobs" $fullName -}}
{{- end -}}
{{- define "jobs-ui.name" -}}
  {{- $name := include "api.name" . -}}
  {{- printf "%s-jobs-ui" $name -}}
{{- end -}}
{{- define "jobs-ui.fullname" -}}
  {{- $fullName := include "api.fullname" . -}}
  {{- printf "%s-jobs-ui" $fullName -}}
{{- end -}}