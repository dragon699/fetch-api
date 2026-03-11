package response

type BaseResponse[Item any] struct {
	TotalItems int    `json:"total_items"`
	Items      []Item `json:"items"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Error            string         `json:"error"`
	UpstreamResponse map[string]any `json:"upstream_response,omitempty"`
}
