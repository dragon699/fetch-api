package response

type NotificationStatus struct {
	Success   bool   `json:"success"`
	Channel   string `json:"channel,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}
