package service

import (
	"errors"
	"slices"

	"connector-downloader/internal/config"
	"connector-downloader/internal/qbittorrent"
)

type AddTorrentOptions struct {
	URL      string
	Category string
	Tags     []string
	SavePath string
	Manage   bool
	FindSubs bool
	Notify   bool
}

var ErrFindSubsNotJellyfin = errors.New("find_subs can only be true when `category` is jellyfin")

func AddTorrent(opts AddTorrentOptions) error {
	if opts.Category == "" {
		opts.Category = "jellyfin"
	}

	if opts.SavePath == "" {
		opts.SavePath = config.Config.QBittorrentDefaultSavePath
	}

	tags := slices.Clone(opts.Tags)
	if tags == nil {
		tags = []string{}
	}

	addTag := func(tag string) {
		if !slices.Contains(tags, tag) {
			tags = append(tags, tag)
		}
	}

	if opts.Category == "jellyfin" {
		addTag("jellyfin:rename=pending")
	}

	if opts.Manage {
		addTag("fetch-api")
	}

	if opts.FindSubs {
		if opts.Category != "jellyfin" {
			return ErrFindSubsNotJellyfin
		}

		addTag("jellyfin:find_subs=pending")
	}

	if opts.Notify {
		addTag("slack:notify=pending")
	}

	return qbittorrent.Client.AddTorrent(opts.URL, opts.Category, tags, opts.SavePath)
}
