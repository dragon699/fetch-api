package slack

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"common/utils"
	"connector-slack/internal/config"

	slackapi "github.com/slack-go/slack"
)

//go:embed templates/notifications/*/*.tpl
var notificationTemplates embed.FS
var Client *SlackClient

type SlackClient struct {
	BotToken          string
	AppToken          string
	DefaultChannel    string
	APIBaseURL        string
	APIDefaultHeaders map[string]string
	Client            *slackapi.Client
}

func (instance *SlackClient) Init() {
	instance.BotToken = config.Config.SlackBotToken
	instance.AppToken = config.Config.SlackAppToken
	instance.DefaultChannel = config.Config.SlackDefaultChannelID

	instance.APIBaseURL = "https://slack.com/api"
	instance.APIDefaultHeaders = map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", instance.BotToken),
		"Content-Type":  "application/json; charset=utf-8",
	}

	instance.Client = slackapi.New(
		instance.BotToken,
		slackapi.OptionAppLevelToken(instance.AppToken),
		slackapi.OptionDebug(config.Config.SlackSocketDebug),
	)
}

func (instance *SlackClient) Ping() error {
	_, err := instance.Client.AuthTestContext(context.Background())

	if err != nil {
		return config.NewConnectionError("Slack auth test failed", err)
	}

	return nil
}

func (instance *SlackClient) SendEphemeralMsg(channelID string, userID string, blocks []map[string]any, attachments []map[string]any, options ...slackapi.MsgOption) error {
	opts := append(
		[]slackapi.MsgOption{
			slackapi.MsgOptionBlocks(toBlockSet(blocks)...),
			slackapi.MsgOptionAttachments(toAttachmentSet(attachments)...),
		},
		options...,
	)

	_, err := instance.Client.PostEphemeral(channelID, userID, opts...)

	if err != nil {
		return config.NewUpstreamError(
			fmt.Sprintf("chat.postEphemeral failed for channel %q and user %q", channelID, userID),
			0,
			nil,
			err,
		)
	}

	return nil
}

func (instance *SlackClient) SendMsg(channelID string, blocks []map[string]any, attachments []map[string]any, options ...slackapi.MsgOption) (*MessageResponse, error) {
	opts := append(
		[]slackapi.MsgOption{
			slackapi.MsgOptionBlocks(toBlockSet(blocks)...),
			slackapi.MsgOptionAttachments(toAttachmentSet(attachments)...),
		},
		options...,
	)
	ch, ts, err := instance.Client.PostMessage(channelID, opts...)

	if err != nil {
		return nil, config.NewUpstreamError(
			fmt.Sprintf("chat.postMessage failed for channel %q", channelID),
			0,
			nil,
			err,
		)
	}

	return &MessageResponse{
		Channel:     ch,
		Timestamp:   ts,
		Blocks:      blocks,
		Attachments: attachments,
	}, nil
}

func (instance *SlackClient) SendMsgFromTemplate(channel string, app string, templatePath string, templateVars any, options ...slackapi.MsgOption) (*MessageResponse, error) {
	var msg Message
	var msgPayload map[string]any
	var appName, appIcon string

	msgPayloadName := strings.Split(path.Base(templatePath), ".")[0]

	raw, err := utils.RenderTemplate(templatePath, templateVars, notificationTemplates)
	if err != nil {
		return nil, fmt.Errorf("failed to render notification template %q: %w", templatePath, err)
	}

	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal notification template %q: %w", templatePath, err)
	}

	if metaJSON, err := json.Marshal(templateVars); err == nil {
		_ = json.Unmarshal(metaJSON, &msgPayload)
	}

	switch app {
	case "grafana":
		appName = config.Config.SlackGrafanaUsername
		appIcon = config.Config.SlackGrafanaIconURL
	case "connector-downloader":
		appName = config.Config.SlackConnectorDownloaderUsername
		appIcon = config.Config.SlackConnectorDownloaderIconURL
	}

	opts := []slackapi.MsgOption{
		slackapi.MsgOptionMetadata(slackapi.SlackMetadata{
			EventType:    msgPayloadName,
			EventPayload: msgPayload,
		}),
		slackapi.MsgOptionBlocks(toBlockSet(msg.Blocks)...),
		slackapi.MsgOptionAttachments(toAttachmentSet(msg.Attachments)...),
	}

	opts = append(opts, options...)

	if appName != "" {
		opts = append(opts, slackapi.MsgOptionUsername(appName))
	}

	if appIcon != "" {
		opts = append(opts, slackapi.MsgOptionIconURL(appIcon))
	}

	if msg.Text != "" {
		opts = append(opts, slackapi.MsgOptionText(msg.Text, false))
	}

	ch, ts, err := instance.Client.PostMessage(channel, opts...)
	if err != nil {
		return nil, config.NewUpstreamError(
			fmt.Sprintf("chat.postMessage failed for %q", templatePath),
			0,
			nil,
			err,
		)
	}

	return &MessageResponse{
		Channel:     ch,
		Timestamp:   ts,
		Blocks:      msg.Blocks,
		Attachments: msg.Attachments,
		Meta: map[string]string{
			"username": appName,
			"icon_url": appIcon,
		},
	}, nil
}

func (instance *SlackClient) UpdateMsg(channelID string, ts string, blocks []map[string]any, attachments []map[string]any, options ...slackapi.MsgOption) (*MessageResponse, error) {
	opts := append(
		[]slackapi.MsgOption{
			slackapi.MsgOptionBlocks(toBlockSet(blocks)...),
			slackapi.MsgOptionAttachments(toAttachmentSet(attachments)...),
		},
		options...,
	)
	// _, _, _, err := instance.Client.UpdateMessage(channelID, ts, opts...)
	ch, newTs, _, err := instance.Client.UpdateMessage(channelID, ts, opts...)

	if err != nil {
		return nil, config.NewUpstreamError(
			fmt.Sprintf("chat.update failed for channel %q and timestamp %q", channelID, ts),
			0,
			nil,
			err,
		)
	}

	return &MessageResponse{
		Channel:     ch,
		Timestamp:   newTs,
		Blocks:      blocks,
		Attachments: attachments,
	}, nil
}
