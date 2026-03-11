package jellyfin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"common/utils"
	"connector-downloader/internal/config"
)

var Client *JellyfinClient

type JellyfinClient struct {
	APIBaseURL        string
	APIDefaultHeaders map[string]string
	Client            *http.Client
}

func (instance *JellyfinClient) Init() {
	instance.APIBaseURL = config.Config.JellyfinUrl
	instance.APIDefaultHeaders = map[string]string{
		"Content-Type": "application/json",
		"X-Emby-Token": config.Config.JellyfinAPIKey,
	}
	instance.Client = &http.Client{
		Timeout: 10 * time.Second,
	}
}

func (instance *JellyfinClient) RefreshLibrary() error {
	req := &utils.Req{Client: instance.Client}

	_, err := req.POST(
		fmt.Sprintf("%s/Library/Refresh", instance.APIBaseURL),
		instance.APIDefaultHeaders,
		nil,
		nil,
	)

	return err
}

func (instance *JellyfinClient) GetItems() ([]Item, error) {
	req := &utils.Req{Client: instance.Client}

	resp, err := req.GET(
		fmt.Sprintf("%s/Items", instance.APIBaseURL),
		instance.APIDefaultHeaders,
		map[string]any{
			"Recursive":        true,
			"Fields":           "Path,DateCreated,MediaStreams",
			"IncludeItemTypes": "Episode,Movie,Series,Trailer,Video",
			"sortBy":           "DateCreated",
			"sortOrder":        "Descending",
		},
	)

	if err != nil {
		return nil, err
	}

	body, ok := resp.Body.(map[string]any)

	if !ok {
		return nil, config.NewUpstreamError("Invalid Jellyfin items response", resp.StatusCode, nil, nil)
	}

	itemsRaw, ok := body["Items"]
	if !ok {
		return nil, config.NewUpstreamError("Invalid Jellyfin items response", resp.StatusCode, nil, nil)
	}

	rawItems, err := json.Marshal(itemsRaw)

	if err != nil {
		return nil, err
	}

	result := []Item{}

	if err := json.Unmarshal(rawItems, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (instance *JellyfinClient) DownloadSubtitles(itemID string, language string) error {
	req := &utils.Req{Client: instance.Client}

	searchResp, err := req.GET(
		fmt.Sprintf("%s/Items/%s/RemoteSearch/Subtitles/%s", instance.APIBaseURL, itemID, language),
		instance.APIDefaultHeaders,
		nil,
	)

	if err != nil {
		return err
	}

	subtitlesRaw, ok := searchResp.Body.([]map[string]any)
	if !ok {
		return config.NewUpstreamError("Invalid Jellyfin subtitles response", searchResp.StatusCode, nil, nil)
	}

	rawSubtitles, err := json.Marshal(subtitlesRaw)
	if err != nil {
		return err
	}

	subs := []RemoteSubtitle{}
	if err := json.Unmarshal(rawSubtitles, &subs); err != nil {
		return err
	}

	if len(subs) == 0 {
		return fmt.Errorf("No %s subtitles found for item ID %s in Jellyfin", language, itemID)
	}

	subsID := subs[0].ID

	_, err = req.POST(
		fmt.Sprintf("%s/Items/%s/RemoteSearch/Subtitles/%s", instance.APIBaseURL, itemID, subsID),
		instance.APIDefaultHeaders,
		nil,
		nil,
	)

	return err
}
