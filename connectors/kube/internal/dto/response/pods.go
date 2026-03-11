package response

type Pod struct {
	Name       string         `json:"name"`
	Namespace  string         `json:"namespace"`
	Owners     []PodOwner     `json:"owners,omitempty"`
	Containers []PodContainer `json:"containers,omitempty"`
	CreatedAt  string         `json:"created_at,omitempty"`
	StartedAt  string         `json:"started_at,omitempty"`
	Status     string         `json:"status,omitempty"`
	Conditions []PodCondition `json:"conditions,omitempty"`
}

type PodStatus struct {
	Status string `json:"status"`
}

type PodOwner struct {
	Name string `json:"name,omitempty"`
	Kind string `json:"kind,omitempty"`
}

type PodCondition struct {
	Name               string `json:"name"`
	Status             string `json:"status"`
	LastTransitionTime string `json:"last_transition_time,omitempty"`
	Reason             string `json:"reason,omitempty"`
}

type PodContainer struct {
	Name   string             `json:"name"`
	Image  string             `json:"image"`
	Status PodContainerStatus `json:"status,omitempty"`
}

type PodContainerStatus struct {
	Started              bool                     `json:"started,omitempty"`
	Ready                bool                     `json:"ready,omitempty"`
	RestartCount         int                      `json:"restart_count,omitempty"`
	LastTerminationState *PodContainerStatusState `json:"last_termination_state,omitempty"`
	State                PodContainerStatusState  `json:"state,omitempty"`
}

type PodContainerStatusState struct {
	StartedAt  string `json:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
	Reason     string `json:"reason,omitempty"`
	ExitCode   int    `json:"exit_code,omitempty"`
}

type PodListResponse struct {
	TotalItems int   `json:"total_items"`
	Items      []Pod `json:"items"`
}
