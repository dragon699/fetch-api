package handlers

import (
	"errors"
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

	if err := ctx.BodyParser(&reqPayload); err != nil {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "Invalid request payload",
			},
		)
	}

	if reqPayload.ChannelID == "" {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "channel_id is required",
			},
		)
	}

	if len(reqPayload.Blocks) == 0 && len(reqPayload.Attachments) == 0 {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "At least one block or attachment is required in blocks or attachments",
			},
		)
	}

	options := []slackapi.MsgOption{
		slackapi.MsgOptionUsername(reqPayload.Options.User),
		slackapi.MsgOptionIconURL(reqPayload.Options.UserIcon),
		slackapi.MsgOptionText(reqPayload.Options.ExtraText, false),
	}

	result, err := slack.Client.SendMsg(reqPayload.ChannelID, reqPayload.Blocks, reqPayload.Attachments, options...)

	if err != nil {
		var clientErr *config.ClientError
		if errors.As(err, &clientErr) {
			return ctx.Status(502).JSON(
				response.ErrorResponse{
					Error:            err.Error(),
					UpstreamResponse: clientErr.UpstreamResponse(),
				},
			)
		}

		return ctx.Status(500).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
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

	if err := ctx.BodyParser(&reqPayload); err != nil {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "Invalid request payload",
			},
		)
	}

	if len(reqPayload.Alerts) == 0 {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "Got 0 alerts in payload, at least 1 is required",
			},
		)
	}

	for _, alert := range reqPayload.Alerts {
		templatePath := fmt.Sprintf("%s/notifications/grafana/%s.tpl", config.Config.TemplatesBasePath, "alert")
		result, err := slack.Client.SendMsgFromTemplate(config.Config.SlackGrafanaAlertsChannelID, "grafana", templatePath, alert)

		if err != nil {
			var clientErr *config.ClientError
			if errors.As(err, &clientErr) {
				return ctx.Status(502).JSON(
					response.ErrorResponse{
						Error:            err.Error(),
						UpstreamResponse: clientErr.UpstreamResponse(),
					},
				)
			}

			return ctx.Status(500).JSON(
				response.ErrorResponse{
					Error: err.Error(),
				},
			)
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

			_, err := slack.Client.SendMsg(result.Channel, blocks, nil, options...)

			if err != nil {
				var clientErr *config.ClientError
				if errors.As(err, &clientErr) {
					return ctx.Status(502).JSON(
						response.ErrorResponse{
							Error:            err.Error(),
							UpstreamResponse: clientErr.UpstreamResponse(),
						},
					)
				}

				return ctx.Status(500).JSON(
					response.ErrorResponse{
						Error: err.Error(),
					},
				)
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

	if err := ctx.BodyParser(&reqPayload); err != nil {
		return ctx.Status(400).JSON(
			response.ErrorResponse{
				Error: "Invalid request payload",
			},
		)
	}

	templatePath := fmt.Sprintf("%s/notifications/connector-downloader/%s.tpl", config.Config.TemplatesBasePath, "torrent")
	result, err := slack.Client.SendMsgFromTemplate(config.Config.SlackConnectorDownloaderChannelID, "connector-downloader", templatePath, reqPayload)

	if err != nil {
		var clientErr *config.ClientError
		if errors.As(err, &clientErr) {
			return ctx.Status(502).JSON(
				response.ErrorResponse{
					Error:            err.Error(),
					UpstreamResponse: clientErr.UpstreamResponse(),
				},
			)
		}

		return ctx.Status(500).JSON(
			response.ErrorResponse{
				Error: err.Error(),
			},
		)
	}

	return ctx.JSON(
		response.NotificationStatus{
			Success:   true,
			Channel:   result.Channel,
			Timestamp: result.Timestamp,
		},
	)
}
