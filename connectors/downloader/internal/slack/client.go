package slack

import (
	"fmt"
	"net/http"
	"time"

	"common/utils"
	"connector-downloader/internal/config"
)

var Client *SlackClient

type SlackClient struct {
	WebhookURL            string
	WebhookDefaultHeaders map[string]string
	Client                *http.Client
}

func (instance *SlackClient) Init() {
	instance.WebhookURL = config.Config.SlackNotificationsWebhookUrl
	instance.WebhookDefaultHeaders = map[string]string{
		"Content-Type": "application/json",
	}
	instance.Client = &http.Client{
		Timeout: 10 * time.Second,
	}
}

func (instance *SlackClient) SendMessage(templateName string, templateVars any) error {
	body, err := RenderTemplate(templateName, templateVars)

	if err != nil {
		return fmt.Errorf("Failed to render Slack message template: %w", err)
	}

	req := &utils.Req{Client: instance.Client}

	_, err = req.POST(
		instance.WebhookURL,
		instance.WebhookDefaultHeaders,
		nil,
		body,
	)

	return err
}
