package torrents

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"common/utils"
	"connector-downloader/internal/config"
	"connector-downloader/internal/http/dto/response"
	"connector-downloader/internal/jellyfin"
	"connector-downloader/internal/qbittorrent"
	"connector-downloader/internal/slack"
	t "connector-downloader/internal/telemetry"
)

func TorrentPostDownload(torrent response.Torrent) ([]qbittorrent.TorrentContentFile, []string, error) {
	torrentContent, err := qbittorrent.Client.ListTorrentContents(torrent.Hash)

	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get torrent content: %w", err)
	}

	var torrentContentFiles = []qbittorrent.TorrentContentFile{}
	var torrentContentNewFileNames = []string{}
	var allowedExtensions = []string{
		".mkv", ".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".m4v",
		".mpg", ".mpeg", ".ts", ".m2ts", ".mts", ".3gp", ".3g2", ".ogv",
		".vob", ".asf", ".rm", ".rmvb", ".divx", ".f4v", ".mxf", ".mpv",
		".qt", ".dat", ".amv", ".y4m",
	}

	for _, file := range torrentContent {
		if file.Progress < 1 {
			os.Remove(path.Join(torrent.SavePath, file.Name))
		} else {
			if !slices.Contains(allowedExtensions, path.Ext(file.Name)) {
				continue
			}

			fileName := path.Base(file.Name)
			fileExt := path.Ext(fileName)

			torrentContentFiles = append(torrentContentFiles, file)
			torrentContentNewFileNames = append(
				torrentContentNewFileNames,
				fmt.Sprintf(
					"%s%s",
					utils.BeautifyMovieName(strings.TrimSuffix(fileName, fileExt)),
					fileExt,
				),
			)
		}
	}

	return torrentContentFiles, torrentContentNewFileNames, nil
}

func SlackNotify(stage string, torrent response.Torrent) error {
	var lastTag, newTag string

	switch stage {
	case "initial":
		lastTag = "slack:notify=pending"
		newTag = "slack:notify=initial"
	case "completed":
		lastTag = "slack:notify=initial"
		newTag = "slack:notify=completed"
	}

	templateVars := TorrentSlackNotificationVars{
		TorrentName:    torrent.Name,
		Category:       torrent.Category,
		QBittorrentURL: config.Config.QBittorrentPublicUrl,
		JellyfinURL:    config.Config.JellyfinUrl,
	}

	err := slack.Client.SendMessage(
		fmt.Sprintf("torrents_%s", stage),
		templateVars,
	)

	if err != nil {
		t.Log.Error("Failed to send slack notification for a torrent!", "error", err.Error())

		tagErr := SwitchTorrentTags(torrent.Hash, []string{lastTag}, []string{"slack:notify=failed"})

		if tagErr != nil {
			t.Log.Error(fmt.Sprintf("Failed to update tags for slack:notify action status for torrent %s", torrent.Hash), "error", tagErr.Error(), "torrent_hash", torrent.Hash)

			return fmt.Errorf("Failed to send slack notification and update tags for torrent %s: notify error: %w; tag error: %v", torrent.Hash, err, tagErr)
		}

		return fmt.Errorf("Failed to send slack notification for a torrent: %w", err)
	}

	err = SwitchTorrentTags(torrent.Hash, []string{lastTag}, []string{newTag})

	if err != nil {
		t.Log.Error(fmt.Sprintf("Failed to update tags for slack:notify action status for torrent %s", torrent.Hash), "error", err.Error(), "torrent_hash", torrent.Hash)

		return fmt.Errorf("Failed to update slack:notify tags for torrent %s after sending notification: %w", torrent.Hash, err)
	}

	return nil
}

func JellyfinRename(torrent response.Torrent, torrentContentFiles []qbittorrent.TorrentContentFile, torrentContentNewFileNames []string) error {
	var renameFailed bool = false

	for _, file := range torrentContentFiles {
		filePath := path.Dir(file.Name)
		fileName := path.Base(file.Name)
		fileExt := path.Ext(fileName)
		fileNameBase := strings.TrimSuffix(fileName, fileExt)

		fileNameNew := fmt.Sprintf(
			"%s%s",
			utils.BeautifyMovieName(fileNameBase),
			fileExt,
		)

		srcFile := path.Join(torrent.SavePath, file.Name)
		destPath := path.Join(torrent.SavePath, filePath)
		destFile := path.Join(destPath, fileNameNew)

		err := os.Rename(srcFile, destFile)
		if err != nil {
			renameFailed = true
		}
	}

	var dirNameNew string
	dirPath := filepath.Dir(torrent.FilesPath)
	dirName := filepath.Base(torrent.FilesPath)

	if len(torrentContentNewFileNames) == 1 {
		dirNameNew = strings.TrimSuffix(torrentContentNewFileNames[0], path.Ext(torrentContentNewFileNames[0]))
	} else {
		dirNameNew = utils.BeautifyMovieName(dirName)
	}

	dirPathNew := path.Join(
		dirPath,
		dirNameNew,
	)

	err := os.Rename(torrent.FilesPath, dirPathNew)
	if err != nil {
		renameFailed = true
	}

	jellyfin.Client.RefreshLibrary()
	time.Sleep(2 * time.Second)

	qbittorrent.Client.DeleteTorrentTags(torrent.Hash, []string{"jellyfin:rename=pending"})

	if renameFailed {
		qbittorrent.Client.AddTorrentTags(torrent.Hash, []string{"jellyfin:rename=failed"})

		t.Log.Error("Failed to rename torrent files", "action", "jellyfin:rename", "torrent_hash", torrent.Hash)
		return fmt.Errorf("Failed to rename torrent files")
	}

	qbittorrent.Client.AddTorrentTags(torrent.Hash, []string{"jellyfin:rename=completed"})
	return nil
}

func JellyfinFindSubs(torrent response.Torrent, torrentContentNewFileNames []string) error {
	var subsDownloadedCount int = 0
	var subsAlreadyPresentCount int = 0

	jellyfin.Client.RefreshLibrary()
	time.Sleep(2 * time.Second)

	jellyfinItems, err := jellyfin.Client.GetItems()

	if err != nil {
		t.Log.Error("Failed to get Jellyfin items", "error", err.Error())

		err := SwitchTorrentTags(torrent.Hash, []string{"jellyfin:find_subs=pending"}, []string{"jellyfin:find_subs=failed"})

		if err != nil {
			t.Log.Error(fmt.Sprintf("Failed to update tags for jellyfin:find_subs action status for torrent %s", torrent.Hash), "error", err.Error(), "torrent_hash", torrent.Hash)
		}

		return fmt.Errorf("Failed to get Jellyfin items")
	}

	for _, item := range jellyfinItems {
		fileName := filepath.Base(item.Path)
		
		if !slices.Contains(torrentContentNewFileNames, fileName) {
			continue
		}

		if item.HasSubtitles {
			var subtitlesFound bool = false

			for _, stream := range item.MediaStreams {
				if stream.Type == "Subtitle" && stream.Language == config.Config.JellyfinSubtitlesDefaultLanguage[:3] {
					subtitlesFound = true

					break
				}
			}

			if subtitlesFound {
				subsAlreadyPresentCount += 1

				continue
			}
		}

		err = jellyfin.Client.DownloadSubtitles(item.ID, config.Config.JellyfinSubtitlesDefaultLanguage)

		if err != nil {
			t.Log.Error("Failed to download subtitles in Jellyfin", "error", err.Error())

			continue
		}

		subsDownloadedCount += 1
	}

	qbittorrent.Client.DeleteTorrentTags(torrent.Hash, []string{"jellyfin:find_subs=pending"})

	if subsAlreadyPresentCount == len(torrentContentNewFileNames) {
		qbittorrent.Client.AddTorrentTags(torrent.Hash, []string{"jellyfin:find_subs=already_present"})
	} else if (subsDownloadedCount + subsAlreadyPresentCount) == len(torrentContentNewFileNames) {
		qbittorrent.Client.AddTorrentTags(torrent.Hash, []string{"jellyfin:find_subs=completed"})
	} else {
		qbittorrent.Client.AddTorrentTags(torrent.Hash, []string{"jellyfin:find_subs=partially_completed"})
	}

	return nil
}
