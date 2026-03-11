package qbittorrent

type Torrent struct {
	Name      string  `json:"name"`
	Hash      string  `json:"hash,omitempty"`
	Category  string  `json:"category,omitempty"`
	Tags      string  `json:"tags,omitempty"`
	State     string  `json:"state,omitempty"`
	Progress  float64 `json:"progress,omitempty"`
	ETA       int64   `json:"eta,omitempty"`
	MagnetURI string  `json:"magnet_uri,omitempty"`
	Leechers  int64   `json:"num_leechs,omitempty"`
	Seeders   int64   `json:"num_seeds,omitempty"`

	AddedOn      int64 `json:"added_on,omitempty"`
	LastActivity int64 `json:"last_activity,omitempty"`
	CompletionOn int64 `json:"completion_on,omitempty"`

	TotalSize  int64 `json:"total_size,omitempty"`
	Downloaded int64 `json:"downloaded,omitempty"`
	Uploaded   int64 `json:"uploaded,omitempty"`
	AmountLeft int64 `json:"amount_left,omitempty"`
	Size       int64 `json:"size,omitempty"`

	Availability float64 `json:"availability,omitempty"`
	ContentPath  string  `json:"content_path,omitempty"`
	SavePath     string  `json:"save_path,omitempty"`

	SpeedDownload int64 `json:"dlspeed,omitempty"`
	SpeedUpload   int64 `json:"upspeed,omitempty"`
}

type TorrentContentFile struct {
	Name     string  `json:"name"`
	Size     int64   `json:"size,omitempty"`
	Progress float64 `json:"progress,omitempty"`
	Priority int64   `json:"priority,omitempty"`
}
