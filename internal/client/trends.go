package client

import (
	"fmt"

	"github.com/yoanbernabeu/bugsnag-cli/internal/models"
)

func (c *Client) GetProjectTrends(projectID, resolution string, bucketsCount int) ([]models.TrendBucket, error) {
	path := fmt.Sprintf("/projects/%s/trend", projectID)
	params := map[string]string{}
	if resolution != "" {
		params["resolution"] = resolution
	}
	if bucketsCount > 0 {
		params["buckets_count"] = fmt.Sprintf("%d", bucketsCount)
	}

	req, err := c.newRequest("GET", path, toURLValues(params))
	if err != nil {
		return nil, err
	}

	var buckets []models.TrendBucket
	_, err = c.do(req, &buckets)
	return buckets, err
}

func (c *Client) GetErrorTrends(projectID, errorID string) ([]models.TrendBucket, error) {
	path := fmt.Sprintf("/projects/%s/errors/%s/trend", projectID, errorID)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var buckets []models.TrendBucket
	_, err = c.do(req, &buckets)
	return buckets, err
}
