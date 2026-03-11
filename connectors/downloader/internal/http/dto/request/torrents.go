package request

type AddTorrentPayload struct {
	URL      string   `json:"url"`
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	SavePath string   `json:"save_path,omitempty"`
	Manage   *bool    `json:"manage,omitempty"`
	FindSubs *bool    `json:"find_subs,omitempty"`
	Notify   *bool    `json:"notify,omitempty"`
}

type AddTorrentTagsPayload struct {
	Hash string   `json:"hash"`
	Tags []string `json:"tags"`
}

type DeleteTorrentTagsPayload struct {
	Hash string   `json:"hash"`
	Tags []string `json:"tags"`
}

type SearchTorrentsParams struct {
	Query string `json:"query"`
}
