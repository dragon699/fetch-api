{
	"attachments": [
		{
			"color": "#00b389",
			"fallback": "Download completed - {{ .TorrentName }}",
			"blocks": [
				{
					"type": "section",
					"text": {
						"type": "mrkdwn",
						"text": "*{{ .TorrentName }}*\nDownload completed!"
					}
				},
				{
					"type": "actions",
					"elements": [
						{
							"type": "button",
							"value": "open_qbittorrent",
							"text": {
								"type": "plain_text",
								"text": "qBittorrent 🡕",
								"emoji": true
							},
							"url": "{{ .QBittorrentURL }}"
						}
						{{ if eq .Category "jellyfin" }}
						,{
							"type": "button",
							"value": "open_jellyfin",
							"text": {
								"type": "plain_text",
								"text": "Jellyfin 🡕",
								"emoji": true
							},
							"url": "{{ .JellyfinURL }}"
						}
						{{ end }}
					]
				}
			]
		}
	]
}
