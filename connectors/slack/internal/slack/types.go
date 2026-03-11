package slack

import (
	"encoding/json"

	slackapi "github.com/slack-go/slack"
)

type Message struct {
	Channel     string           `json:"channel"`
	Username    string           `json:"username,omitempty"`
	IconURL     string           `json:"icon_url,omitempty"`
	Text        string           `json:"text,omitempty"`
	Blocks      []map[string]any `json:"blocks,omitempty"`
	Attachments []map[string]any `json:"attachments,omitempty"`
}

type MessageResponse struct {
	OK          bool
	Channel     string
	Timestamp   string
	Blocks      []map[string]any
	Attachments []map[string]any
	Meta        map[string]string
}

type rawBlock struct {
	data map[string]any
}

func (b rawBlock) BlockType() slackapi.MessageBlockType {
	if t, ok := b.data["type"].(string); ok {
		return slackapi.MessageBlockType(t)
	}
	return ""
}

func (b rawBlock) ID() string {
	if id, ok := b.data["block_id"].(string); ok {
		return id
	}
	return ""
}

func (b rawBlock) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.data)
}

func toBlockSet(blocks []map[string]any) []slackapi.Block {
	blockSet := make([]slackapi.Block, len(blocks))
	for i, b := range blocks {
		blockSet[i] = rawBlock{data: b}
	}
	return blockSet
}

func toAttachmentSet(attachments []map[string]any) []slackapi.Attachment {
	result := make([]slackapi.Attachment, 0, len(attachments))
	for _, a := range attachments {
		b, _ := json.Marshal(a)
		var att slackapi.Attachment
		if err := json.Unmarshal(b, &att); err == nil {
			result = append(result, att)
		}
	}
	return result
}
