package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	PerPage    int
	HTTPClient *http.Client
}

type APIError struct {
	StatusCode int
	Message    string
	Errors     []map[string]string `json:"errors,omitempty"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("API error (%d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("API error (%d)", e.StatusCode)
}

func New(baseURL, token string, perPage int) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		PerPage: perPage,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) newRequest(method, path string, params url.Values) (*http.Request, error) {
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if params == nil {
		params = url.Values{}
	}
	if _, ok := params["per_page"]; !ok && c.PerPage > 0 {
		params.Set("per_page", fmt.Sprintf("%d", c.PerPage))
	}
	u.RawQuery = params.Encode()

	req, err := http.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) newRequestWithBody(method, path string, body io.Reader) (*http.Request, error) {
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) do(req *http.Request, v any) (*http.Response, error) {
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		apiErr := &APIError{StatusCode: resp.StatusCode}

		var errResp struct {
			Errors []map[string]string `json:"errors"`
		}
		if json.Unmarshal(body, &errResp) == nil && len(errResp.Errors) > 0 {
			if msg, ok := errResp.Errors[0]["message"]; ok {
				apiErr.Message = msg
			}
			apiErr.Errors = errResp.Errors
		} else {
			apiErr.Message = string(body)
		}

		return resp, apiErr
	}

	if v != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, fmt.Errorf("reading response: %w", err)
		}
		if err := json.Unmarshal(body, v); err != nil {
			return resp, fmt.Errorf("decoding response: %w", err)
		}
	}

	return resp, nil
}
