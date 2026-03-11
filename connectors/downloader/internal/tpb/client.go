package tpb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"common/utils"
	"connector-downloader/internal/config"
)

var Client *TPBClient

type TPBClient struct {
	APIBaseURL        string
	APIDefaultHeaders map[string]string
	Client            *http.Client
}

func (instance *TPBClient) Init() {
	instance.APIBaseURL = config.Config.TPBAPIUrl
	instance.APIDefaultHeaders = map[string]string{}
	instance.Client = &http.Client{
		Timeout: 10 * time.Second,
	}
}

func (instance *TPBClient) SearchTorrents(category int64, query string) ([]Torrent, error) {
	req := utils.Req{Client: instance.Client}

	resp, err := req.GET(
		fmt.Sprintf("%s/q.php", instance.APIBaseURL),
		instance.APIDefaultHeaders,
		map[string]any{
			"q":   query,
			"cat": category,
		},
	)

	if err != nil {
		return nil, err
	}

	torrents, ok := resp.Body.([]map[string]any)

	if !ok {
		return nil, config.NewUpstreamError("Invalid TPB search response", resp.StatusCode, nil, nil)
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
