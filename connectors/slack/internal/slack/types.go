package slack

import (
	slackapi "github.com/slack-go/slack"
)

type View struct {
	Type   string           `json:"type"`
	Title  *TextObject      `json:"title,omitempty"`
	Blocks []map[string]any `json:"blocks,omitempty"`
	Close  *TextObject      `json:"close,omitempty"`
	Submit *TextObject      `json:"submit,omitempty"`
}

type ViewResponse = slackapi.ViewResponse

type Message struct {
	Channel     string           `json:"channel"`
	Username    string           `json:"username,omitempty"`
	IconURL     string           `json:"icon_url,omitempty"`
	Text        string           `json:"text,omitempty"`
	Blocks      []map[string]any `json:"blocks,omitempty"`
	Attachments []map[string]any `json:"attachments,omitempty"`
}

type MessageResponse struct {
	OK          bool
	Channel     string
	Timestamp   string
	Blocks      []map[string]any
	Attachments []map[string]any
	Meta        map[string]string
}

type TextObject struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
