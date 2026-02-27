package client

import (
	"fmt"

	"github.com/yoanbernabeu/bugsnag-cli/internal/models"
)

func (c *Client) ListReleases(projectID string, allPages bool) ([]models.Release, bool, error) {
	path := fmt.Sprintf("/projects/%s/releases", projectID)
	if allPages {
		items, err := CollectAllPages[models.Release](c, path, nil)
		return items, false, err
	}
	return FetchSinglePage[models.Release](c, path, nil)
}
