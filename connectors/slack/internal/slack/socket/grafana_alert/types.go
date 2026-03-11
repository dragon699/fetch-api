package grafana_alert

type ConnectorMLResponse struct {
	TotalItems int               `json:"total_items,omitempty"`
	Items      []ConnectorMLItem `json:"items,omitempty"`
}

type ConnectorMLItem struct {
	Answer          string  `json:"answer,omitempty"`
	DurationSeconds float64 `json:"duration_seconds,omitempty"`
	Provider        string  `json:"provider,omitempty"`
	Model           string  `json:"model,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
}
