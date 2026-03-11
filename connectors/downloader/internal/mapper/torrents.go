package mapper

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"common/utils"
	"connector-downloader/internal/http/dto/response"
	"connector-downloader/internal/qbittorrent"
	"connector-downloader/internal/tpb"
)

func TorrentsFromQBittorrent(torrents []qbittorrent.Torrent) []response.Torrent {
	result := make([]response.Torrent, 0, len(torrents))

	for _, torrentSrc := range torrents {
		tags := []string{}

		if torrentSrc.Tags != "" {
			tags = strings.Split(torrentSrc.Tags, ", ")
		}

		torrentData := response.Torrent{
			Name:                        torrentSrc.Name,
			Hash:                        torrentSrc.Hash,
			Category:                    torrentSrc.Category,
			Tags:                        tags,
			Status:                      torrentSrc.State,
			ProgressPercentage:          utils.ProgressToPercentage(torrentSrc.Progress),
			EtaMinutes:                  utils.SecondsToMinutes(torrentSrc.ETA),
			MagnetURI:                   torrentSrc.MagnetURI,
			Leechers:                    torrentSrc.Leechers,
			Seeders:                     torrentSrc.Seeders,
			DateAdded:                   utils.TimeFromUnix(torrentSrc.AddedOn),
			DateLastActivity:            utils.TimeFromUnix(torrentSrc.LastActivity),
			DateCompleted:               utils.TimeFromUnix(torrentSrc.CompletionOn),
			SizeTotalMB:                 utils.BytesToMegabytes(torrentSrc.TotalSize),
			SizeDownloadedMB:            utils.BytesToMegabytes(torrentSrc.Downloaded),
			SizeUploadedMB:              utils.BytesToMegabytes(torrentSrc.Uploaded),
			SizeLeftMB:                  utils.BytesToMegabytes(torrentSrc.AmountLeft),
			SizeMB:                      utils.BytesToMegabytes(torrentSrc.Size),
			FilesAvailabilityPercentage: utils.RoundToTwoDecimals(torrentSrc.Availability * 100),
			FilesPath:                   torrentSrc.ContentPath,
			SavePath:                    torrentSrc.SavePath,
			SpeedDownloadMBps:           utils.BytesPerSecondToMBps(torrentSrc.SpeedDownload),
			SpeedUploadMBps:             utils.BytesPerSecondToMBps(torrentSrc.SpeedUpload),
		}

		torrentMeta := response.TorrentMeta{}

		if slices.Contains(torrentData.Tags, "fetch-api") {
			torrentMeta.ManagedBy = "connector-downloader"
		} else {
			torrentMeta.ManagedBy = "qBittorrent"
		}

		statesDownloading := []string{
			"allocating", "downloading", "metaDL", "queuedDL", "stalledDL", "checkingDL", "forcedDL",
		}
		statesPaused := []string{
			"pausedUP", "pausedDL", "stoppedDL",
		}
		statesError := []string{
			"error", "missingFiles",
		}

		if slices.Contains(statesDownloading, torrentData.Status) {
			torrentData.Status = "downloading"
		} else if slices.Contains(statesPaused, torrentData.Status) {
			torrentData.Status = "paused"
		} else if slices.Contains(statesError, torrentData.Status) {
			torrentData.Status = "error"
		} else {
			torrentData.Status = "unknown"
		}

		if torrentData.ProgressPercentage == 100 {
			torrentData.Status = "completed"
			torrentData.EtaMinutes = 0
		} else {
			torrentData.DateCompleted = ""
		}

		for _, tag := range torrentData.Tags {
			tagParts := strings.Split(tag, ":")

			if !(len(tagParts) > 1) {
				continue
			}

			tagOpParts := strings.Split(tagParts[1], "=")

			if !(len(tagOpParts) > 1) {
				continue
			}

			tagAction := response.TorrentMetaScheduledAction{}

			tagCategory := tagParts[0]
			tagOpName := tagOpParts[0]
			tagOpStatus := tagOpParts[1]

			if tagCategory == "jellyfin" {
				if tagOpName == "rename" {
					switch tagOpStatus {
					case "pending":
						tagAction.Description = "Torrent dir/files will be renamed to match Jellyfin library structure once completed."
					case "completed":
						tagAction.Description = "Torrent dir/files renamed to match Jellyfin library structure."
					case "failed":
						tagAction.Description = "[!] Something went wrong while renaming Torrent's content."
					}
				}

				if tagOpName == "find_subs" {
					switch tagOpStatus {
					case "pending":
						tagAction.Description = "Subtitles will be fetched from OpenSubtitles in Jellyfin for this torrent media."
					case "completed":
						tagAction.Description = "Subtitles in Jellyfin fully fetched."
					case "partially_completed":
						tagAction.Description = "Subtitles fetched in Jellyfin for some of the media."
					case "already_present":
						tagAction.Description = "Subtitles and preferred language already present in Jellyfin for this media."
					case "failed":
						tagAction.Description = "[!] Something went wrong while fetching subtitles from OpenSubtitles in Jellyfin for this torrent media."
					}
				}
			}

			if tagCategory == "slack" {
				if tagOpName == "notify" {
					switch tagOpStatus {
					case "pending":
						tagAction.Description = "Slack notification still not sent."
					case "initial":
						tagAction.Description = "Initial Slack notification already sent, awaiting for torrent completion."
					case "completed":
						tagAction.Description = "Slack notifications sent."
					case "failed":
						tagAction.Description = "[!] Something went wrong while sending notifications to Slack."
					}
				}
			}

			tagAction.Name = tagOpName
			tagAction.Status = tagOpStatus
			tagAction.Category = tagCategory

			if (tagAction.Name != "") && (tagAction.Status != "") {
				torrentMeta.ScheduledActions = append(torrentMeta.ScheduledActions, tagAction)
			}
		}

		torrentData.Meta = torrentMeta
		result = append(result, torrentData)
	}

	return result
}

func TorrentsFromTPB(torrents []tpb.Torrent) []response.Torrent {
	result := make([]response.Torrent, 0, len(torrents))

	for _, torrentSrc := range torrents {
		id, _ := utils.ToInt(torrentSrc.ID)
		magnetURI := fmt.Sprintf(
			"magnet:?xt=urn:btih:%s&dn=%s",
			torrentSrc.InfoHash,
			url.QueryEscape(torrentSrc.Name),
		)
		leechers, _ := utils.ToInt(torrentSrc.Leechers)
		seeders, _ := utils.ToInt(torrentSrc.Seeders)
		sizeTotalB, _ := utils.ToInt(torrentSrc.Size)
		filesCount, _ := utils.ToInt(torrentSrc.NumFiles)
		dateAdded, _ := utils.ToInt(torrentSrc.Added)

		torrentData := response.Torrent{
			ID:          id,
			Name:        torrentSrc.Name,
			Hash:        torrentSrc.InfoHash,
			MagnetURI:   magnetURI,
			Leechers:    leechers,
			Seeders:     seeders,
			SizeTotalMB: utils.BytesToMegabytes(sizeTotalB),
			SizeTotalGB: utils.BytesToGigabytes(sizeTotalB),
			FilesCount:  filesCount,
			DateAdded:   utils.TimeFromUnix(dateAdded),
			IMDB:        torrentSrc.IMDB,
		}
		result = append(result, torrentData)
	}

	if len(result) > 0 && result[0].ID == 0 {
		result = []response.Torrent{}
	}

	return result
}
