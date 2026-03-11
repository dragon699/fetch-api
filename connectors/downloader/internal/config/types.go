package config

import (
	"encoding/json"
	"fmt"
)

type ClientError struct {
	Kind       ClientErrorKind
	Message    string
	StatusCode int
	Body       string
	ParsedJSON any
	Err        error
}

type ClientErrorKind string

const (
	ClientErrorConnection ClientErrorKind = "connection"
	ClientErrorUpstream   ClientErrorKind = "upstream_api"
)

func (e *ClientError) Error() string {
	if e == nil {
		return ""
	}

	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

func (e *ClientError) UpstreamResponse() map[string]any {
	if e == nil {
		return nil
	}

	resp := map[string]any{
		"error_type": string(e.Kind),
	}

	if e.StatusCode > 0 {
		resp["status_code"] = e.StatusCode
	}

	if e.Body != "" {
		resp["body"] = e.Body
	}

	if e.ParsedJSON != nil {
		resp["json"] = e.ParsedJSON
	}

	return resp
}

func parseJSONBody(body []byte) any {
	var parsed any

	if len(body) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil
	}

	return parsed
}

func NewConnectionError(message string, err error) error {
	return &ClientError{
		Kind:    ClientErrorConnection,
		Message: message,
		Err:     err,
	}
}

func NewUpstreamError(message string, statusCode int, body []byte, err error) error {
	return &ClientError{
		Kind:       ClientErrorUpstream,
		Message:    message,
		StatusCode: statusCode,
		Body:       string(body),
		ParsedJSON: parseJSONBody(body),
		Err:        err,
	}
}
