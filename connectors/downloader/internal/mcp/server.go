package mcp

import (
	"cmp"
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"connector-downloader/internal/config"
	"connector-downloader/internal/http/dto/response"
	"connector-downloader/internal/mapper"
	"connector-downloader/internal/qbittorrent"
	"connector-downloader/internal/service"
	"connector-downloader/internal/torrent_indexer"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type SearchResult struct {
	Name      string  `json:"name"`
	SizeGB    float64 `json:"size_gb"`
	Seeders   int64   `json:"seeders"`
	Leechers  int64   `json:"leechers"`
	IMDB      string  `json:"imdb,omitempty"`
	MagnetURI string  `json:"magnet_uri"`
}

type Download struct {
	Hash        string            `json:"hash"`
	Name        string            `json:"name"`
	Status      string            `json:"status"`
	Progress    float64           `json:"progress_percentage"`
	EtaMinutes  int64             `json:"eta_minutes,omitempty"`
	SizeGB      float64           `json:"size_gb"`
	SpeedMBps   float64           `json:"speed_download_mbps,omitempty"`
	PostActions map[string]string `json:"post_actions,omitempty" jsonschema:"post-download actions and their status, e.g. jellyfin:find_subs=completed"`
}

func toDownload(torrent response.Torrent) Download {
	actions := map[string]string{}
	for _, action := range torrent.Meta.ScheduledActions {
		actions[action.Category+":"+action.Name] = action.Status
	}

	return Download{
		Hash:        torrent.Hash,
		Name:        torrent.Name,
		Status:      torrent.Status,
		Progress:    torrent.ProgressPercentage,
		EtaMinutes:  torrent.EtaMinutes,
		SizeGB:      torrent.SizeTotalGB,
		SpeedMBps:   torrent.SpeedDownloadMBps,
		PostActions: actions,
	}
}

// search_torrents

type SearchInput struct {
	Query string `json:"query" jsonschema:"movie or series name, optionally with year, e.g. 'Dune Part Two 2024'"`
	Limit int    `json:"limit,omitempty" jsonschema:"max results (default 10, max 25)"`
}

type SearchOutput struct {
	Results []SearchResult `json:"results"`
}

func searchTorrents(ctx context.Context, _ *sdk.CallToolRequest, in SearchInput) (*sdk.CallToolResult, SearchOutput, error) {
	if strings.TrimSpace(in.Query) == "" {
		return nil, SearchOutput{}, fmt.Errorf("query is required")
	}

	limit := in.Limit
	if limit <= 0 || limit > 25 {
		limit = 10
	}

	// 200 = All video media
	raw, err := torrent_indexer.Client.SearchTorrents(200, in.Query)
	if err != nil {
		return nil, SearchOutput{}, fmt.Errorf("torrent search failed: %w", err)
	}

	torrents := mapper.TorrentsFromTorrentIndexer(raw)
	slices.SortStableFunc(torrents, func(a, b response.Torrent) int {
		return cmp.Compare(b.Seeders, a.Seeders)
	})

	out := SearchOutput{Results: []SearchResult{}}
	for _, torrent := range torrents {
		if torrent.Seeders == 0 || len(out.Results) == limit {
			break
		}

		out.Results = append(out.Results, SearchResult{
			Name:      torrent.Name,
			SizeGB:    torrent.SizeTotalGB,
			Seeders:   torrent.Seeders,
			Leechers:  torrent.Leechers,
			IMDB:      torrent.IMDB,
			MagnetURI: torrent.MagnetURI,
		})
	}

	return nil, out, nil
}

// download

type DownloadInput struct {
	MagnetURI string `json:"magnet_uri" jsonschema:"magnet URI or .torrent URL, usually from search_torrents"`
	FindSubs  *bool  `json:"find_subs,omitempty" jsonschema:"download subtitles via Jellyfin after the download completes (default true)"`
	Notify    *bool  `json:"notify,omitempty" jsonschema:"send Slack notifications on start and completion (default true)"`
}

type DownloadOutput struct {
	Message string `json:"message"`
}

func download(ctx context.Context, _ *sdk.CallToolRequest, in DownloadInput) (*sdk.CallToolResult, DownloadOutput, error) {
	if strings.TrimSpace(in.MagnetURI) == "" {
		return nil, DownloadOutput{}, fmt.Errorf("magnet_uri is required")
	}

	boolOr := func(value *bool, fallback bool) bool {
		if value == nil {
			return fallback
		}
		return *value
	}

	findSubs := boolOr(in.FindSubs, true)

	err := service.AddTorrent(service.AddTorrentOptions{
		URL:      in.MagnetURI,
		Category: "jellyfin",
		Manage:   true,
		FindSubs: findSubs,
		Notify:   boolOr(in.Notify, true),
	})
	if err != nil {
		return nil, DownloadOutput{}, fmt.Errorf("failed to add torrent: %w", err)
	}

	message := "Added to qBittorrent. When it completes it will be renamed for Jellyfin"
	if findSubs {
		message += fmt.Sprintf(" and %s subtitles will be fetched", config.Config.JellyfinSubtitlesDefaultLanguage)
	}
	message += ". Track it with list_downloads."

	return nil, DownloadOutput{Message: message}, nil
}

// list_downloads

type ListInput struct {
	ActiveOnly bool `json:"active_only,omitempty" jsonschema:"only torrents that are still downloading"`
}

type ListOutput struct {
	Downloads []Download `json:"downloads"`
}

func listDownloads(ctx context.Context, _ *sdk.CallToolRequest, in ListInput) (*sdk.CallToolResult, ListOutput, error) {
	raw, err := qbittorrent.Client.ListTorrents()
	if err != nil {
		return nil, ListOutput{}, fmt.Errorf("failed to list torrents: %w", err)
	}

	out := ListOutput{Downloads: []Download{}}
	for _, torrent := range mapper.TorrentsFromQBittorrent(raw) {
		if in.ActiveOnly && torrent.ProgressPercentage >= 100 {
			continue
		}

		out.Downloads = append(out.Downloads, toDownload(torrent))
	}

	return nil, out, nil
}

// get_download

type GetInput struct {
	Hash string `json:"hash" jsonschema:"torrent hash from list_downloads"`
}

func getDownload(ctx context.Context, _ *sdk.CallToolRequest, in GetInput) (*sdk.CallToolResult, Download, error) {
	if strings.TrimSpace(in.Hash) == "" {
		return nil, Download{}, fmt.Errorf("hash is required")
	}

	raw, err := qbittorrent.Client.ListTorrents()
	if err != nil {
		return nil, Download{}, fmt.Errorf("failed to list torrents: %w", err)
	}

	for _, torrent := range mapper.TorrentsFromQBittorrent(raw) {
		if strings.EqualFold(torrent.Hash, in.Hash) {
			return nil, toDownload(torrent), nil
		}
	}

	return nil, Download{}, fmt.Errorf("no torrent with hash %s", in.Hash)
}

// Server

func NewHandler(version string) http.Handler {
	server := sdk.NewServer(&sdk.Implementation{Name: config.Config.Name, Version: version}, nil)

	sdk.AddTool(server, &sdk.Tool{
		Name:        "search_torrents",
		Description: "Search public torrent indexes for a movie or series. Results are sorted by seeders; prefer high seeders and a sensible size (1080p is usually 2-10 GB).",
	}, searchTorrents)

	sdk.AddTool(server, &sdk.Tool{
		Name:        "download",
		Description: "Start downloading a torrent into the Jellyfin library. After completion the files are renamed for Jellyfin and subtitles are fetched automatically.",
	}, download)

	sdk.AddTool(server, &sdk.Tool{
		Name:        "list_downloads",
		Description: "List torrents in qBittorrent with progress, ETA and the status of post-download actions (rename, subtitles, notify).",
	}, listDownloads)

	sdk.AddTool(server, &sdk.Tool{
		Name:        "get_download",
		Description: "Get the status of a single torrent by hash.",
	}, getDownload)

	// Stateless with plain JSON responses: no sessions and no SSE streaming,
	// which keeps it compatible with Fiber's fasthttp adaptor (it buffers responses).
	return sdk.NewStreamableHTTPHandler(
		func(*http.Request) *sdk.Server { return server },
		&sdk.StreamableHTTPOptions{Stateless: true, JSONResponse: true},
	)
}
