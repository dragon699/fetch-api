package jellyfin

type Item struct {
	ID           string `json:"Id,omitempty"`
	Path         string `json:"Path,omitempty"`
	HasSubtitles bool   `json:"HasSubtitles,omitempty"`
	MediaStreams []struct {
		Type     string `json:"Type,omitempty"`
		Language string `json:"Language,omitempty"`
	} `json:"MediaStreams,omitempty"`
}

type RemoteSubtitle struct {
	ID       string `json:"Id,omitempty"`
	Language string `json:"Language,omitempty"`
}
