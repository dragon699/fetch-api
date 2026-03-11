{
	"attachments": [
		{
			"color": "#00baed",
			"fallback": "Now downloading - {{ .TorrentName }}",
			"blocks": [
				{
					"type": "section",
					"text": {
						"type": "mrkdwn",
						"text": "*{{ .TorrentName }}*\nNow downloading"
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
					]
				}
			]
		}
	]
}
