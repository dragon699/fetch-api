package response

type Deployment struct {
	Name       string                `json:"name"`
	Namespace  string                `json:"namespace"`
	CreatedAt  string                `json:"created_at,omitempty"`
	Replicas   int32                 `json:"replicas,omitempty"`
	Status     DeploymentStatus      `json:"status,omitempty"`
	Conditions []DeploymentCondition `json:"conditions,omitempty"`
}

type DeploymentStatus struct {
	Replicas          int32 `json:"replicas,omitempty"`
	AvailableReplicas int32 `json:"available_replicas,omitempty"`
	ReadyReplicas     int32 `json:"ready_replicas,omitempty"`
	UpdatedReplicas   int32 `json:"updated_replicas,omitempty"`
	Generation        int64 `json:"generation,omitempty"`
}

type DeploymentCondition struct {
	Name               string `json:"name"`
	Status             string `json:"status"`
	LastTransitionTime string `json:"last_transition_time,omitempty"`
	Reason             string `json:"reason,omitempty"`
}

type DeploymentListResponse struct {
	TotalItems int          `json:"total_items"`
	Items      []Deployment `json:"items"`
}
