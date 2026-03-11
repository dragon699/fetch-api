package grafana_alert

import (
	"encoding/json"
	"fmt"

	"connector-slack/internal/config"
	"connector-slack/internal/notifications"
	"connector-slack/internal/slack"
	t "connector-slack/internal/telemetry"

	slackapi "github.com/slack-go/slack"
)

func ButtonInvestigate(value string, message slackapi.Message, user string) {
	var alert notifications.GrafanaAlertItem
	askMsg := []map[string]any{
		{
			"type": "section",
			"text": map[string]any{
				"type": "mrkdwn",
				"text": fmt.Sprintf("<@%s>, check it please :point_up::skin-tone-3:", config.Config.SlackOpenClawUserID),
			},
		},
	}

	alertJSON, err := json.Marshal(message.Metadata.EventPayload)

	if err == nil {
		_ = json.Unmarshal(alertJSON, &alert)
	}

	_, err = slack.Client.SendMsg(
		config.Config.SlackGrafanaAlertsChannelID,
		askMsg,
		nil,
		slackapi.MsgOptionText(fmt.Sprintf("%s > Investigation requested from Stitch", alert.Labels["alert_name"]), false),
		slackapi.MsgOptionTS(message.Timestamp),
	)

	if err != nil {
		t.Log.Error("Failed to send message", "error", err)
	}

	for i := len(message.Blocks.BlockSet) - 1; i >= 0; i-- {
		actionsBlock, ok := message.Blocks.BlockSet[i].(*slackapi.ActionBlock)

		if !ok || len(actionsBlock.Elements.ElementSet) == 0 {
			continue
		}

		filtered := actionsBlock.Elements.ElementSet[:0]
		for _, element := range actionsBlock.Elements.ElementSet {
			button, ok := element.(*slackapi.ButtonBlockElement)
			if ok && button.ActionID == "grafana_alert_button_investigate" {
				continue
			}
			filtered = append(filtered, element)
		}

		if len(filtered) == len(actionsBlock.Elements.ElementSet) {
			continue
		}

		actionsBlock.Elements.ElementSet = filtered

		var blocks []map[string]any
		if raw, err := json.Marshal(message.Blocks.BlockSet); err == nil {
			_ = json.Unmarshal(raw, &blocks)
		}

		if _, err := slack.Client.UpdateMsg(config.Config.SlackGrafanaAlertsChannelID, message.Timestamp, blocks, nil); err != nil {
			t.Log.Error("Failed to update message status", "error", err)
			return
		}

		break
	}
}
