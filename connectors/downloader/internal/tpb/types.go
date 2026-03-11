package tpb

type Torrent struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	InfoHash string `json:"info_hash"`
	Leechers string `json:"leechers"`
	Seeders  string `json:"seeders"`
	Size     string `json:"size"`
	NumFiles string `json:"num_files"`
	Added    string `json:"added"`
	IMDB     string `json:"imdb,omitempty"`
}
