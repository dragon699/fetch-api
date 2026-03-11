package response

type BaseResponse[Item any] struct {
	TotalItems int    `json:"total_items"`
	Items      []Item `json:"items"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
