package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type PageResult[T any] struct {
	Items   []T
	NextURL string
	HasMore bool
}

var linkNextRe = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

func parseLinkHeader(header string) string {
	matches := linkNextRe.FindStringSubmatch(header)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func fetchPage[T any](c *Client, req *http.Request) (*PageResult[T], error) {
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
		return nil, apiErr
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var items []T
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	nextURL := parseLinkHeader(resp.Header.Get("Link"))

	return &PageResult[T]{
		Items:   items,
		NextURL: nextURL,
		HasMore: nextURL != "",
	}, nil
}

func FetchSinglePage[T any](c *Client, path string, params map[string]string) ([]T, bool, error) {
	req, err := c.newRequest("GET", path, toURLValues(params))
	if err != nil {
		return nil, false, err
	}

	result, err := fetchPage[T](c, req)
	if err != nil {
		return nil, false, err
	}

	return result.Items, result.HasMore, nil
}

func CollectAllPages[T any](c *Client, path string, params map[string]string) ([]T, error) {
	var all []T

	req, err := c.newRequest("GET", path, toURLValues(params))
	if err != nil {
		return nil, err
	}

	for {
		result, err := fetchPage[T](c, req)
		if err != nil {
			return nil, err
		}

		all = append(all, result.Items...)

		if !result.HasMore {
			break
		}

		req, err = http.NewRequest("GET", result.NextURL, nil)
		if err != nil {
			return nil, fmt.Errorf("building next page request: %w", err)
		}
		req.Header.Set("Authorization", "token "+c.Token)
		req.Header.Set("Accept", "application/json")
	}

	return all, nil
}

func toURLValues(params map[string]string) map[string][]string {
	if params == nil {
		return nil
	}
	values := make(map[string][]string)
	for k, v := range params {
		if v != "" {
			values[k] = []string{v}
		}
	}
	return values
}
