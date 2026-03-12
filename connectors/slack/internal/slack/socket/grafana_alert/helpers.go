package grafana_alert

import (
	"encoding/json"

	"connector-slack/internal/notifications"

	slackapi "github.com/slack-go/slack"
)

func parseAlert(msg slackapi.Message) notifications.GrafanaAlertItem {
	var alert notifications.GrafanaAlertItem

	jsonData, _ := json.Marshal(msg.Metadata.EventPayload)
	_ = json.Unmarshal(jsonData, &alert)

	return alert
}
