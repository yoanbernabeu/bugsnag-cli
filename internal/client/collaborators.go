package client

import (
	"fmt"

	"github.com/yoanbernabeu/bugsnag-cli/internal/models"
)

func (c *Client) ListCollaborators(orgID string, allPages bool) ([]models.Collaborator, bool, error) {
	path := fmt.Sprintf("/organizations/%s/collaborators", orgID)
	if allPages {
		items, err := CollectAllPages[models.Collaborator](c, path, nil)
		return items, false, err
	}
	return FetchSinglePage[models.Collaborator](c, path, nil)
}
