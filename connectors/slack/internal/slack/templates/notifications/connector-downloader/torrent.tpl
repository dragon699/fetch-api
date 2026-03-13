{{- $color := "#4cc9f0" -}}
{{- $titleText := "Now downloading" -}}
{{- $notification := "Download started" -}}

{{- if eq .Stage "completed" -}}
  {{- $color = "#72c18f" -}}
  {{- $titleText = "Download completed" -}}
  {{- $notification = "Download completed" -}}
{{- end -}}


{
	"text": {{ json $notification }},
	"attachments": [
		{
			"color": {{ json $color }},
			"fallback": {{ json $notification }},
			"blocks": [
				{
					"type": "section",
					"text": {
						"type": "mrkdwn",
						"text": "*{{ .Name }}*\n{{ $titleText }}"
					}
				},
				{
					"type": "actions",
					"elements": [
						{
							"type": "button",
             			    "action_id": "torrent_button_qbittorrent",
							"text": {
								"type": "plain_text",
								"text": "qBittorrent ↠",
								"emoji": true
							},
							"url": {{ json .QBittorrentURL }}
						}
						{{ if eq .Category "jellyfin" }}
						,{
							"type": "button",
              				"action_id": "torrent_button_jellyfin",
							"text": {
								"type": "plain_text",
								"text": "Jellyfin ↠",
								"emoji": true
							},
							"url": {{ json .JellyfinURL }}
						}
						{{ end }}
					]
				}
			]
		}
	]
}
