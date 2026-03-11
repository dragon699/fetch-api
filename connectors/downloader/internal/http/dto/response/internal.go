package response

type HealthResponse struct {
	ConnectorName   string `json:"connector_name"`
	Healthy         *bool  `json:"healthy"`
	HealthLastCheck string `json:"health_last_check,omitempty"`
	HealthNextCheck string `json:"health_next_check,omitempty"`
}
type ReadyResponse struct {
	Ready *bool `json:"ready"`
}
