{
	"type": "modal",
	"title": {
		"type": "plain_text",
		"text": "Alert values"
	},
	"close": {
		"type": "plain_text",
		"text": "Close"
	},
	"blocks": [
		{{- $hasValues := .Values -}}
		{{- $hasLabels := .Labels -}}
		{{- $hasAnnotations := .Annotations -}}
		{{- if $hasValues }}
		{
			"type": "header",
			"text": {
				"type": "plain_text",
				"text": "Values"
			}
		},
		{
			"type": "section",
			"fields": [
				{{- $first := true }}
				{{- range $key, $value := .Values }}
				{{- if not $first }},{{ end }}
				{
					"type": "mrkdwn",
					"text": {{ json (printf "*%s*\n%v" $key $value) }}
				}
				{{- $first = false }}
				{{- end }}
			]
		}
		{{- end }}
		{{- if and $hasValues $hasLabels }},{{ end }}
		{{- if $hasLabels }}
		{
			"type": "header",
			"text": {
				"type": "plain_text",
				"text": "Labels"
			}
		},
		{
			"type": "section",
			"fields": [
				{{- $first := true }}
				{{- range $key, $value := .Labels }}
				{{- if not $first }},{{ end }}
				{
					"type": "mrkdwn",
					"text": {{ json (printf "*%s*\n%v" $key $value) }}
				}
				{{- $first = false }}
				{{- end }}
			]
		}
		{{- end }}
		{{- if and (or $hasValues $hasLabels) $hasAnnotations }},{{ end }}
		{{- if $hasAnnotations }}
		{
			"type": "header",
			"text": {
				"type": "plain_text",
				"text": "Annotations"
			}
		},
		{
			"type": "section",
			"fields": [
				{{- $first := true }}
				{{- range $key, $value := .Annotations }}
				{{- if not $first }},{{ end }}
				{
					"type": "mrkdwn",
					"text": {{ json (printf "*%s*\n%v" $key $value) }}
				}
				{{- $first = false }}
				{{- end }}
			]
		}
		{{- end }}
	]
}
