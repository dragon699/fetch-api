package response

type Namespace struct {
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type NamespaceListResponse struct {
	TotalItems int         `json:"total_items"`
	Items      []Namespace `json:"items"`
}
