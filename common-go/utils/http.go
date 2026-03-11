package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	netUrl "net/url"
	"strings"
	"time"
)

type Req struct {
	Client *http.Client
}

type Response struct {
	StatusCode int
	Body       any
	Bytes      []byte
}

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

func newConnectionError(message string, err error) error {
	return &ClientError{
		Kind:    ClientErrorConnection,
		Message: message,
		Err:     err,
	}
}

func newUpstreamError(message string, statusCode int, body []byte, err error) error {
	return &ClientError{
		Kind:       ClientErrorUpstream,
		Message:    message,
		StatusCode: statusCode,
		Body:       string(body),
		ParsedJSON: parseJSONBody(body),
		Err:        err,
	}
}

func (e *ClientError) Error() string {
	if e == nil {
		return ""
	}

	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}

	return e.Message
}

func (e *ClientError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
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

func normalizeJSONBody(parsed any) any {
	if arr, ok := parsed.([]any); ok {
		maps := make([]map[string]any, 0, len(arr))
		for _, item := range arr {
			m, ok := item.(map[string]any)
			if !ok {
				return parsed
			}
			maps = append(maps, m)
		}
		return maps
	}
	return parsed
}

func (r *Req) GET(url string, headers map[string]string, params map[string]any) (*Response, error) {
	if r.Client == nil {
		r.Client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	if len(headers) == 0 {
		headers = map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
		}
	}

	if len(params) > 0 {
		reqParams := netUrl.Values{}

		for key, value := range params {
			reqParams.Set(key, fmt.Sprintf("%v", value))
		}

		url = fmt.Sprintf("%s?%s", url, reqParams.Encode())
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, newConnectionError(fmt.Sprintf("[GET] [%s] Request failed", url), err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := r.Client.Do(req)

	if err != nil {
		return nil, newConnectionError(fmt.Sprintf("[GET] [%s] Request failed", url), err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, newConnectionError(fmt.Sprintf("[GET] [%s] Failed to read response body", url), err)
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil, newUpstreamError(fmt.Sprintf("[GET] [%s] Returned non-2xx status code", url), resp.StatusCode, respBody, nil)
	}

	respContentType := strings.ToLower(resp.Header.Get("Content-Type"))
	isJSONResponse := strings.Contains(respContentType, "application/json")

	var parsedBody any

	if isJSONResponse && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &parsedBody); err != nil {
			return nil, newUpstreamError(fmt.Sprintf("[GET] [%s] Returned invalid JSON", url), resp.StatusCode, respBody, err)
		}

		parsedBody = normalizeJSONBody(parsedBody)
	} else {
		parsedBody = string(respBody)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       parsedBody,
		Bytes:      respBody,
	}, nil
}

func (r *Req) POST(url string, headers map[string]string, params map[string]any, body map[string]any) (*Response, error) {
	if r.Client == nil {
		r.Client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	if len(headers) == 0 {
		headers = map[string]string{
			"Content-Type": "application/json",
		}
	}

	if len(params) > 0 {
		reqParams := netUrl.Values{}

		for key, value := range params {
			reqParams.Set(key, fmt.Sprintf("%v", value))
		}

		url = fmt.Sprintf("%s?%s", url, reqParams.Encode())
	}

	var bodyReader io.Reader

	if len(body) > 0 {
		contentType := ""
		for k, v := range headers {
			if strings.ToLower(k) == "content-type" {
				contentType = strings.ToLower(v)
				break
			}
		}

		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			formData := netUrl.Values{}

			for key, value := range body {
				formData.Set(key, fmt.Sprintf("%v", value))
			}

			bodyReader = strings.NewReader(formData.Encode())
		} else {
			bodyBytes, err := json.Marshal(body)

			if err != nil {
				return nil, newConnectionError(fmt.Sprintf("[POST] [%s] Failed to serialize request body", url), err)
			}

			bodyReader = strings.NewReader(string(bodyBytes))
		}
	}

	req, err := http.NewRequest(http.MethodPost, url, bodyReader)

	if err != nil {
		return nil, newConnectionError(fmt.Sprintf("[POST] [%s] Request failed", url), err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := r.Client.Do(req)

	if err != nil {
		return nil, newConnectionError(fmt.Sprintf("[POST] [%s] Request failed", url), err)
	}

	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, newConnectionError(fmt.Sprintf("[POST] [%s] Failed to read response body", url), err)
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return nil, newUpstreamError(fmt.Sprintf("[POST] [%s] Returned non-2xx status code", url), resp.StatusCode, respBody, nil)
	}

	respContentType := strings.ToLower(resp.Header.Get("Content-Type"))
	isJSONResponse := strings.Contains(respContentType, "application/json")

	var parsedBody any

	if isJSONResponse && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &parsedBody); err != nil {
			return nil, newUpstreamError(fmt.Sprintf("[POST] [%s] Returned invalid JSON", url), resp.StatusCode, respBody, err)
		}

		parsedBody = normalizeJSONBody(parsedBody)
	} else {
		parsedBody = string(respBody)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       parsedBody,
	}, nil
}
