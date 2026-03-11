{{- $status := "Firing" -}}
{{- $titleIcon := ":grafana_alert:" -}}
{{- if eq .Status "resolved" -}}
  {{- $status = "Recovered" -}}
  {{- $titleIcon = ":grafana_alert_recovered:" -}}
{{- end -}}
{{- $titleText := or (index .Annotations "summary") (index .Labels "alertname") -}}

{{- $notification := printf "%s > %s" (toUpper $status) $titleText -}}
{{- $title := printf "%s *%s* > %s" $titleIcon $status $titleText -}}
{{- $description := or (index .Annotations "description") "" -}}

{{- $willAttachScreenshot := .ImageURL -}}

{{- $footer := printf "Detected %s" (beautifyTime .StartsAt) -}}
{{- if not (hasPrefix .EndsAt "0001") -}}
  {{- $footer = printf "%s\nRecovered %s" $footer (beautifyTime .EndsAt) -}}
{{- end -}}

{{- $dashboardURL := or .PanelURL .DashboardURL -}}



{
    "text": {{ json $notification }},
    "blocks": [
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": {{ json $title }}
			}
		},
		{{- if $description -}}
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": {{ json $description }}
			}
		},
		{{- end -}}
		{
			"type": "section",
			"fields": [
				{
					"type": "mrkdwn",
					"text": {{ json (printf "*Status*\n%s" $status) }}
				}
				{{- range $key, $value := .Labels }}
				{{- if and (ne $key "alertname") (not (hasPrefix $key "grafana_")) }},
				{
					"type": "mrkdwn",
					"text": {{ json (printf "*%s*\n%s" (capitalize $key) $value) }}
				}
				{{- end }}
				{{- end }}
				{{- range $key, $value := .Annotations }}
				{{- if and (ne $key "summary") (ne $key "description") }},
				{
					"type": "mrkdwn",
					"text": {{ json (printf "*%s*\n%s" (capitalize $key) $value) }}
				}
				{{- end }}
				{{- end }}
			]
		},
		{{- if $willAttachScreenshot -}}
		{
			"type": "section",
			"text": {
				"type": "plain_text",
				"text": "Screenshot attached to this message."
			}
		},
		{{- end -}}
		{
			"type": "context",
			"elements": [
				{
					"type": "mrkdwn",
					"text": {{ json $footer }}
				}
			]
		},
		{
			"type": "actions",
			"elements": [
				{{- if eq $status "Firing" -}}
				{
					"type": "button",
					"text": {
						"type": "plain_text",
						"text": "☉ Investigate"
					},
					"action_id": "grafana_alert_button_investigate",
					"value": "pending"
				},
				{{- end -}}
				{
					"type": "button",
					"text": {
						"type": "plain_text",
						"text": "⇲  Values"
					},
					"action_id": "grafana_alert_button_values"
				}
				{{- if $dashboardURL -}},
				{
					"type": "button",
					"text": {
						"type": "plain_text",
						"text": "⚎  View in dashboard"
					},
					"url": {{ json $dashboardURL }},
					"action_id": "grafana_alert_button_view_in_dashboard"
				}
				{{- end }}
			]
		}
	]
}
