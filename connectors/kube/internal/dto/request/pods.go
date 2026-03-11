package request

type ListPodsParams struct {
	Namespace             string `query:"namespace"`
	ContainerNameContains string `query:"container_name_contains"`
	Status                string `query:"status"`
	StatusNot             string `query:"status_not"`
	JobsOnly              bool   `query:"jobs_only"`
}
