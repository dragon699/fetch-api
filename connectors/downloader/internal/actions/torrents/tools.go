package torrents

import (
	"connector-downloader/internal/qbittorrent"
)

func SwitchTorrentTags(torrentHash string, deleteTags []string, addTags []string) error {
	if err := qbittorrent.Client.DeleteTorrentTags(torrentHash, deleteTags); err != nil {
		return err
	}

	return qbittorrent.Client.AddTorrentTags(torrentHash, addTags)
}
