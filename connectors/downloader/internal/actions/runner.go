package actions

import (
	"slices"
	"time"

	run "connector-downloader/internal/actions/torrents"
	"connector-downloader/internal/config"
	"connector-downloader/internal/http/dto/response"
	"connector-downloader/internal/jellyfin"
	"connector-downloader/internal/mapper"
	"connector-downloader/internal/qbittorrent"
	t "connector-downloader/internal/telemetry"
)

func (instance *ActionsRunner) run() {
	lastCheckTime := time.Now().Format("2006-01-02T15:04:05")
	config.Config.TorrentActionsLastCheck = &lastCheckTime

	rawTorrents, err := qbittorrent.Client.ListTorrents()

	if err != nil {
		t.Log.Error("Failed to fetch torrents list from qBittorrent", "error", err.Error())
		nextCheckTime := instance.getNextCheckTime()
		config.Config.TorrentActionsNextCheck = &nextCheckTime

		return
	}

	torrents := mapper.TorrentsFromQBittorrent(rawTorrents)
	torrentActionsOrder := map[string]int{
		"rename":    0,
		"find_subs": 1,
	}

	for _, torrent := range torrents {
		if torrent.ProgressPercentage < 100 {
			for _, action := range torrent.Meta.ScheduledActions {
				shouldNotify := (action.Category == "slack") && (action.Name == "notify") && (action.Status == "pending")

				if shouldNotify {
					err := run.SlackNotify(
						"initial",
						torrent,
					)

					if err != nil {
						t.Log.Error("Failed to send Slack notification for downloading torrent", "error", err.Error())
					}

					break
				}
			}

			continue
		}

		for _, action := range torrent.Meta.ScheduledActions {
			shouldNotify := (action.Category == "slack") && (action.Name == "notify") && (action.Status == "initial")

			if shouldNotify {
				err = run.SlackNotify(
					"completed",
					torrent,
				)

				if err != nil {
					t.Log.Error("Failed to send Slack notification for completed torrent", "error", err.Error())
				}

				break
			}
		}

		if torrent.Meta.ManagedBy != "connector-downloader" {
			continue
		}

		hasPendingActions := false

		slices.SortStableFunc(torrent.Meta.ScheduledActions, func(a, b response.TorrentMetaScheduledAction) int {
			orderA, okA := torrentActionsOrder[a.Name]
			orderB, okB := torrentActionsOrder[b.Name]

			if !okA {
				orderA = 999
			}

			if !okB {
				orderB = 999
			}

			return orderA - orderB
		})

		for _, action := range torrent.Meta.ScheduledActions {
			if action.Status != "pending" {
				continue
			}

			hasPendingActions = true
			qbittorrent.Client.StopTorrent(torrent.Hash)

			if action.Category != "jellyfin" {
				continue
			}

			torrentContentFiles, torrentContentNewFileNames, err := run.TorrentPostDownload(torrent)

			if err != nil {
				t.Log.Error("Failed to execute torrent post-download actions", "error", err.Error())
				continue
			}

			switch action.Name {
			case "rename":
				err := run.JellyfinRename(torrent, torrentContentFiles, torrentContentNewFileNames)

				if err != nil {
					t.Log.Error("Failed to rename torrent files in Jellyfin", "error", err.Error())
					continue
				}

			case "find_subs":
				err := run.JellyfinFindSubs(torrent, torrentContentNewFileNames)

				if err != nil {
					t.Log.Error("Failed to find subtitles for torrent in Jellyfin", "error", err.Error())
					continue
				}
			}
		}

		if !hasPendingActions {
			if torrent.Category == "jellyfin" {
				jellyfin.Client.RefreshLibrary()
			}
		}
	}
}
