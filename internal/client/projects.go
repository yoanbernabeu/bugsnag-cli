package client

import (
	"fmt"

	"github.com/yoanbernabeu/bugsnag-cli/internal/models"
)

func (c *Client) ListProjects(orgID string, allPages bool) ([]models.Project, bool, error) {
	path := fmt.Sprintf("/organizations/%s/projects", orgID)
	if allPages {
		items, err := CollectAllPages[models.Project](c, path, nil)
		return items, false, err
	}
	return FetchSinglePage[models.Project](c, path, nil)
}

func (c *Client) GetProject(projectID string) (*models.Project, error) {
	path := fmt.Sprintf("/projects/%s", projectID)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var project models.Project
	_, err = c.do(req, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}
