package response

type Torrent struct {
	ID                 int64    `json:"id,omitempty"`
	Name               string   `json:"name"`
	Hash               string   `json:"hash,omitempty"`
	Category           string   `json:"category,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	Status             string   `json:"status,omitempty"`
	ProgressPercentage float64  `json:"progress_percentage,omitempty"`
	EtaMinutes         int64    `json:"eta_minutes,omitempty"`
	MagnetURI          string   `json:"magnet_uri,omitempty"`
	Leechers           int64    `json:"leechers,omitempty"`
	Seeders            int64    `json:"seeders,omitempty"`

	DateAdded        string `json:"date_added,omitempty"`
	DateLastActivity string `json:"date_last_activity,omitempty"`
	DateCompleted    string `json:"date_completed,omitempty"`

	SizeTotalMB      int64   `json:"size_total_mb,omitempty"`
	SizeTotalGB      float64 `json:"size_total_gb,omitempty"`
	SizeDownloadedMB int64   `json:"size_downloaded_mb,omitempty"`
	SizeUploadedMB   int64   `json:"size_uploaded_mb,omitempty"`
	SizeLeftMB       int64   `json:"size_left_mb,omitempty"`
	SizeMB           int64   `json:"size_mb,omitempty"`

	FilesCount                  int64   `json:"files_count,omitempty"`
	FilesAvailabilityPercentage float64 `json:"files_availability_percentage,omitempty"`
	FilesPath                   string  `json:"files_path,omitempty"`
	SavePath                    string  `json:"save_path,omitempty"`

	SpeedDownloadMBps float64 `json:"speed_download_mbps,omitempty"`
	SpeedUploadMBps   float64 `json:"speed_upload_mbps,omitempty"`

	IMDB string      `json:"imdb,omitempty"`
	Meta TorrentMeta `json:"meta"`
}

type TorrentMeta struct {
	ManagedBy        string                       `json:"managed_by,omitempty"`
	ScheduledActions []TorrentMetaScheduledAction `json:"scheduled_actions,omitempty"`
}

type TorrentMetaScheduledAction struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	Category    string `json:"category,omitempty"`
}
