package slack

import (
	"encoding/json"
	"fmt"

	"common/utils"
	"connector-slack/internal/config"
)

func wrapReqError(message string, err error) error {
	var clientErr *utils.ClientError
	if ok := AsClientError(err, &clientErr); ok {
		return &config.ClientError{
			Kind:       config.ClientErrorKind(clientErr.Kind),
			Message:    message,
			StatusCode: clientErr.StatusCode,
			Body:       clientErr.Body,
			ParsedJSON: clientErr.ParsedJSON,
			Err:        clientErr.Err,
		}
	}

	return fmt.Errorf("%s: %w", message, err)
}

func AsClientError(err error, target **utils.ClientError) bool {
	if err == nil {
		return false
	}

	clientErr, ok := err.(*utils.ClientError)
	if !ok {
		return false
	}

	*target = clientErr

	return true
}

func asString(value any) string {
	if value == nil {
		return ""
	}

	str, ok := value.(string)
	if !ok {
		return ""
	}

	return str
}

func mustJSON(value any) []byte {
	body, err := json.Marshal(value)
	if err != nil {
		return []byte("{}")
	}

	return body
}
