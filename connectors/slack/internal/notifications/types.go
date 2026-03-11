package notifications

type GrafanaAlert struct {
	Receiver          string             `json:"receiver,omitempty"`
	Status            string             `json:"status,omitempty"`
	OrgID             int64              `json:"orgId,omitempty"`
	Version           string             `json:"version,omitempty"`
	GroupKey          string             `json:"groupKey,omitempty"`
	TruncatedAlerts   int                `json:"truncatedAlerts,omitempty"`
	ExternalURL       string             `json:"externalURL,omitempty"`
	GroupLabels       map[string]string  `json:"groupLabels,omitempty"`
	CommonLabels      map[string]string  `json:"commonLabels,omitempty"`
	CommonAnnotations map[string]string  `json:"commonAnnotations,omitempty"`
	Alerts            []GrafanaAlertItem `json:"alerts,omitempty"`
}

type GrafanaAlertItem struct {
	Status        string            `json:"status,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
	StartsAt      string            `json:"startsAt,omitempty"`
	EndsAt        string            `json:"endsAt,omitempty"`
	Values        map[string]any    `json:"values,omitempty"`
	GeneratorURL  string            `json:"generatorURL,omitempty"`
	Fingerprint   string            `json:"fingerprint,omitempty"`
	SilenceURL    string            `json:"silenceURL,omitempty"`
	DashboardURL  string            `json:"dashboardURL,omitempty"`
	PanelURL      string            `json:"panelURL,omitempty"`
	ImageURL      string            `json:"imageURL,omitempty"`
}

type Torrent struct {
	Name           string `json:"name,omitempty"`
	Category       string `json:"category,omitempty"`
	JellyfinURL    string `json:"jellyfin_url,omitempty"`
	QBittorrentURL string `json:"qbittorrent_url,omitempty"`
}
