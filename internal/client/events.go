package client

import (
	"fmt"

	"github.com/yoanbernabeu/bugsnag-cli/internal/models"
)

func (c *Client) ListEvents(projectID, errorID string, allPages bool) ([]models.Event, bool, error) {
	var path string
	if errorID != "" {
		path = fmt.Sprintf("/projects/%s/errors/%s/events", projectID, errorID)
	} else {
		path = fmt.Sprintf("/projects/%s/events", projectID)
	}

	if allPages {
		items, err := CollectAllPages[models.Event](c, path, nil)
		return items, false, err
	}
	return FetchSinglePage[models.Event](c, path, nil)
}

func (c *Client) GetEvent(projectID, eventID string) (*models.Event, error) {
	path := fmt.Sprintf("/projects/%s/events/%s", projectID, eventID)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var event models.Event
	_, err = c.do(req, &event)
	if err != nil {
		return nil, err
	}
	return &event, nil
}
