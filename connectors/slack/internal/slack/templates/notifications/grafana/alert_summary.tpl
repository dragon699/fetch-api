{
    "text": "Investigation summary is ready",
	"blocks": [
		{
			"type": "context",
			"elements": [
				{
					"type": "mrkdwn",
					"text": {{ json (printf "Thinked *%.1f* seconds." .DurationSeconds) }}
				}
			]
		},
		{
			"type": "section",
			"text": {
				"type": "mrkdwn",
				"text": {{ json .Answer }}
			}
		}
	]
}
