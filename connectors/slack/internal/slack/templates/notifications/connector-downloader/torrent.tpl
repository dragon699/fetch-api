{
  "channel": "C0AJ52LCFNZ",
  "username": "connector-downloader",
  "icon_url": "https://i.imgur.com/wWCcoW6.png",
  "text": "{{ if .Title }}{{ .Title }}{{ else }}Downloader event: {{ .Event }}{{ end }}",
  "blocks": [
    {
      "type": "header",
      "text": {
        "type": "plain_text",
        "text": "{{ if .Title }}{{ .Title }}{{ else }}Downloader event: {{ .Event }}{{ end }}"
      }
    },
    {
      "type": "section",
      "fields": [
        {
          "type": "mrkdwn",
          "text": "*Event*\n{{ .Event }}"
        }{{ if .Severity }},
        {
          "type": "mrkdwn",
          "text": "*Severity*\n{{ .Severity }}"
        }{{ end }}{{ if .Category }},
        {
          "type": "mrkdwn",
          "text": "*Category*\n{{ .Category }}"
        }{{ end }}{{ if .Hash }},
        {
          "type": "mrkdwn",
          "text": "*Hash*\n`{{ .Hash }}`"
        }{{ end }}
      ]
    }{{ if .Message }},
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "{{ .Message }}"
      }
    }{{ end }}{{ if .Fields }},
    {
      "type": "section",
      "fields": [
        {{- range $index, $field := .Fields }}
        {{- if $index }},{{ end }}
        {
          "type": "mrkdwn",
          "text": "*{{ $field.Title }}*\n{{ $field.Value }}"
        }
        {{- end }}
      ]
    }{{ end }}{{ if .Actions }},
    {
      "type": "actions",
      "elements": [
        {{- range $index, $action := .Actions }}
        {{- if $index }},{{ end }}
        {
          "type": "button",
          "action_id": "{{ if $action.ID }}{{ $action.ID }}{{ else }}{{ $action.Label }}{{ end }}",
          "text": {
            "type": "plain_text",
            "text": "{{ $action.Label }}"
          }{{ if $action.Style }},
          "style": "{{ $action.Style }}"{{ end }}{{ if $action.Value }},
          "value": "{{ $action.Value }}"{{ end }}{{ if $action.URL }},
          "url": "{{ $action.URL }}"{{ end }}
        }
        {{- end }}
      ]
    }{{ end }}
  ]
}
