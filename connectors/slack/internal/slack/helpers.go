package slack

import (
	"encoding/json"

	slackapi "github.com/slack-go/slack"
)

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

func asBlocks(blocks []map[string]any) []slackapi.Block {
	blockSet := make([]slackapi.Block, len(blocks))
	for i, b := range blocks {
		blockSet[i] = rawBlock{data: b}
	}
	return blockSet
}

func asAttachments(attachments []map[string]any) []slackapi.Attachment {
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
