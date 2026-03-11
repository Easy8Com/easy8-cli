package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"easy8-cli/internal/config"
)

type Client struct {
	BaseURL string
	APIKey  string
	HTTP    *http.Client
}

func NewClient(cfg config.Config) *Client {
	base := strings.TrimRight(cfg.BaseURL, "/")
	return &Client{
		BaseURL: base,
		APIKey:  cfg.APIKey,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				// Refuse redirect when the HTTP method would change
				// (e.g. 301/302 silently converts PUT/POST to GET,
				// which drops the request body and prevents updates).
				if req.Method != via[0].Method {
					return http.ErrUseLastResponse
				}
				// Carry the API key header to the redirected request.
				if key := via[0].Header.Get("X-Redmine-API-Key"); key != "" {
					req.Header.Set("X-Redmine-API-Key", key)
				}
				return nil
			},
		},
	}
}

type APIError struct {
	StatusCode int
	Body       string
	URL        string
}

func (err APIError) Error() string {
	if err.Body == "" {
		return fmt.Sprintf("api error %d", err.StatusCode)
	}
	return fmt.Sprintf("api error %d: %s", err.StatusCode, err.Body)
}

func (c *Client) doJSON(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	if c.APIKey == "" {
		return fmt.Errorf("missing API key")
	}

	urlValue := c.BaseURL + path
	if query != nil {
		encoded := query.Encode()
		if encoded != "" {
			urlValue = urlValue + "?" + encoded
		}
	}

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, urlValue, bodyReader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Redmine-API-Key", c.APIKey)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(respBody)), URL: urlValue}
	}
	if out == nil {
		return nil
	}
	if len(respBody) == 0 {
		return nil
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return err
	}
	return nil
}
