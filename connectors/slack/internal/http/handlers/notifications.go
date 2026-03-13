package handlers

import (
	"fmt"

	"connector-slack/internal/config"
	"connector-slack/internal/http/dto/request"
	"connector-slack/internal/http/dto/response"
	"connector-slack/internal/slack"

	"github.com/gofiber/fiber/v2"
	slackapi "github.com/slack-go/slack"
)

func SendNotification(ctx *fiber.Ctx) error {
	var reqPayload request.NotificationPayload
	if err := parseBody(ctx, &reqPayload); err != nil {
		return err
	}

	if reqPayload.ChannelID == "" {
		return badRequestError(ctx, "channel_id: missing required field")
	}
	if len(reqPayload.Blocks) == 0 && len(reqPayload.Attachments) == 0 {
		return badRequestError(ctx, "Either `blocks` or `attachments` field with at least 1 item is required")
	}

	options := []slackapi.MsgOption{
		slackapi.MsgOptionUsername(reqPayload.Options.User),
		slackapi.MsgOptionIconURL(reqPayload.Options.UserIcon),
		slackapi.MsgOptionText(reqPayload.Options.ExtraText, false),
	}

	result, err := slack.Client.SendMsg(reqPayload.ChannelID, reqPayload.Blocks, reqPayload.Attachments, options...)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(
		response.NotificationStatus{
			Success:   true,
			Channel:   result.Channel,
			Timestamp: result.Timestamp,
		},
	)
}

func SendGrafanaAlertNotification(ctx *fiber.Ctx) error {
	var reqPayload request.GrafanaAlertNotificationPayload
	if err := parseBody(ctx, &reqPayload); err != nil {
		return err
	}

	if len(reqPayload.Alerts) == 0 {
		return badRequestError(ctx, "`alerts` field is required with at least 1 item")
	}

	for _, alert := range reqPayload.Alerts {
		templatePath := fmt.Sprintf("%s/notifications/grafana/alert.tpl", config.Config.TemplatesBasePath)

		result, err := slack.Client.SendMsgFromTemplate(config.Config.SlackGrafanaAlertsChannelID, "grafana", templatePath, alert)
		if err != nil {
			return serviceError(ctx, err)
		}

		if alert.ImageURL != "" {
			options := []slackapi.MsgOption{
				slackapi.MsgOptionText(fmt.Sprintf("Screenshot attached -> %s", alert.Labels["alertname"]), false),
				slackapi.MsgOptionUsername(config.Config.SlackGrafanaUsername),
				slackapi.MsgOptionIconURL(config.Config.SlackGrafanaIconURL),
				slackapi.MsgOptionTS(result.Timestamp),
			}
			blocks := []map[string]any{
				{
					"type": "image",
					"title": map[string]any{
						"type":  "plain_text",
						"text":  "Screenshot from dashboard",
						"emoji": true,
					},
					"image_url": alert.ImageURL,
					"alt_text":  "Dashboard preview",
				},
			}
			if _, err := slack.Client.SendMsg(result.Channel, blocks, nil, options...); err != nil {
				return serviceError(ctx, err)
			}
		}
	}

	return ctx.JSON(
		response.NotificationStatus{
			Success: true,
		},
	)
}

func SendTorrentNotification(ctx *fiber.Ctx) error {
	var reqPayload request.TorrentNotificationPayload
	if err := parseBody(ctx, &reqPayload); err != nil {
		return err
	}

	templatePath := fmt.Sprintf("%s/notifications/connector-downloader/torrent.tpl", config.Config.TemplatesBasePath)

	result, err := slack.Client.SendMsgFromTemplate(config.Config.SlackConnectorDownloaderChannelID, "connector-downloader", templatePath, reqPayload)
	if err != nil {
		return serviceError(ctx, err)
	}

	return ctx.JSON(
		response.NotificationStatus{
			Success:   true,
			Channel:   result.Channel,
			Timestamp: result.Timestamp,
		},
	)
}
