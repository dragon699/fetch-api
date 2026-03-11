package qbittorrent

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"common/utils"
	"connector-downloader/internal/config"
)

var Client *QBittorrentClient

type QBittorrentClient struct {
	APIBaseURL        string
	APIDefaultHeaders map[string]string
	Client            *http.Client
}

func (instance *QBittorrentClient) Init() {
	instance.APIBaseURL = fmt.Sprintf("%s/api/v2", config.Config.QBittorrentUrl)
	instance.APIDefaultHeaders = map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	instance.Client = &http.Client{
		Timeout: 10 * time.Second,
	}
}

func (instance *QBittorrentClient) Ping() (string, int, error) {
	req := &utils.Req{Client: instance.Client}

	resp, err := req.GET(
		fmt.Sprintf("%s/app/defaultSavePath", instance.APIBaseURL),
		instance.APIDefaultHeaders,
		nil,
	)

	if err != nil {
		return "not_ok", 0, err
	}

	return "ok", resp.StatusCode, nil
}

func (instance *QBittorrentClient) ListTorrents() ([]Torrent, error) {
	req := &utils.Req{Client: instance.Client}

	resp, err := req.GET(
		fmt.Sprintf("%s/torrents/info", instance.APIBaseURL),
		instance.APIDefaultHeaders,
		nil,
	)

	if err != nil {
		return nil, err
	}

	torrents, ok := resp.Body.([]map[string]any)

	if !ok {
		return nil, config.NewUpstreamError("Invalid qBittorrent torrents response", resp.StatusCode, nil, nil)
	}

	rawTorrents, err := json.Marshal(torrents)

	if err != nil {
		return nil, err
	}

	result := []Torrent{}

	if err := json.Unmarshal(rawTorrents, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (instance *QBittorrentClient) ListTorrentContents(torrentHash string) ([]TorrentContentFile, error) {
	req := &utils.Req{Client: instance.Client}

	resp, err := req.GET(
		fmt.Sprintf("%s/torrents/files", instance.APIBaseURL),
		instance.APIDefaultHeaders,
		map[string]any{
			"hash": torrentHash,
		},
	)

	if err != nil {
		return nil, err
	}

	torrentContentFiles, ok := resp.Body.([]map[string]any)

	if !ok {
		return nil, config.NewUpstreamError("Invalid qBittorrent torrent content response", resp.StatusCode, nil, nil)
	}

	rawTorrentContentFiles, err := json.Marshal(torrentContentFiles)

	if err != nil {
		return nil, err
	}

	result := []TorrentContentFile{}

	if err := json.Unmarshal(rawTorrentContentFiles, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (instance *QBittorrentClient) StopTorrent(torrentHash string) error {
	req := &utils.Req{Client: instance.Client}

	_, err := req.POST(
		fmt.Sprintf("%s/torrents/stop", instance.APIBaseURL),
		instance.APIDefaultHeaders,
		nil,
		map[string]any{
			"hashes": torrentHash,
		},
	)

	return err
}

func (instance *QBittorrentClient) AddTorrent(torrentURL string, category string, tags []string, savePath string) error {
	req := &utils.Req{Client: instance.Client}

	resp, err := req.POST(
		fmt.Sprintf("%s/torrents/add", instance.APIBaseURL),
		instance.APIDefaultHeaders,
		nil,
		map[string]any{
			"urls":     torrentURL,
			"savepath": savePath,
			"category": category,
			"tags":     strings.Join(tags, ","),
		},
	)

	if err != nil {
		return err
	}

	if strings.Contains(strings.ToLower(fmt.Sprintf("%v", resp.Body)), "fails") {
		return config.NewUpstreamError("Invalid torrent parameters", 502, []byte(fmt.Sprintf("%v", resp.Body)), nil)
	}

	return nil
}

func (instance *QBittorrentClient) RemoveTorrent(torrentHash string, deleteFiles bool) error {
	req := &utils.Req{Client: instance.Client}

	_, err := req.POST(
		fmt.Sprintf("%s/torrents/delete", instance.APIBaseURL),
		instance.APIDefaultHeaders,
		nil,
		map[string]any{
			"hashes":      torrentHash,
			"deleteFiles": deleteFiles,
		},
	)

	return err
}

func (instance *QBittorrentClient) AddTorrentTags(torrentHash string, tags []string) error {
	req := &utils.Req{Client: instance.Client}

	_, err := req.POST(
		fmt.Sprintf("%s/torrents/addTags", instance.APIBaseURL),
		instance.APIDefaultHeaders,
		nil,
		map[string]any{
			"hashes": torrentHash,
			"tags":   strings.Join(tags, ","),
		},
	)

	return err
}

func (instance *QBittorrentClient) DeleteTorrentTags(torrentHash string, tags []string) error {
	req := &utils.Req{Client: instance.Client}

	_, err := req.POST(
		fmt.Sprintf("%s/torrents/removeTags", instance.APIBaseURL),
		instance.APIDefaultHeaders,
		nil,
		map[string]any{
			"hashes": torrentHash,
			"tags":   strings.Join(tags, ","),
		},
	)

	return err
}
