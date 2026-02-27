package client

import "github.com/yoanbernabeu/bugsnag-cli/internal/models"

func (c *Client) ListOrganizations(allPages bool) ([]models.Organization, bool, error) {
	if allPages {
		items, err := CollectAllPages[models.Organization](c, "/user/organizations", nil)
		return items, false, err
	}
	return FetchSinglePage[models.Organization](c, "/user/organizations", nil)
}
