package request

import (
	"connector-slack/internal/notifications"
)

type NotificationPayload struct {
	ChannelID   string                     `json:"channel_id"`
	Blocks      []map[string]any           `json:"blocks,omitempty"`
	Attachments []map[string]any           `json:"attachments,omitempty"`
	Options     NotificationOptionsPayload `json:"options"`
}

type NotificationOptionsPayload struct {
	User      string `json:"user,omitempty"`
	UserIcon  string `json:"user_icon,omitempty"`
	ExtraText string `json:"extra_text,omitempty"`
}

type GrafanaAlertNotificationPayload = notifications.GrafanaAlert
type TorrentNotificationPayload = notifications.Torrent
