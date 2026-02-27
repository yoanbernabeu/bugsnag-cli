package client

import (
	"fmt"

	"github.com/yoanbernabeu/bugsnag-cli/internal/models"
)

func (c *Client) GetStabilityTrend(projectID, releaseStage string) (*models.StabilityTrend, error) {
	path := fmt.Sprintf("/projects/%s/stability_trend", projectID)
	params := map[string]string{}
	if releaseStage != "" {
		params["release_stage"] = releaseStage
	}

	req, err := c.newRequest("GET", path, toURLValues(params))
	if err != nil {
		return nil, err
	}

	var trend models.StabilityTrend
	_, err = c.do(req, &trend)
	if err != nil {
		return nil, err
	}
	return &trend, nil
}
