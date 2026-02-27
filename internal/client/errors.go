package client

import (
	"fmt"

	"github.com/yoanbernabeu/bugsnag-cli/internal/models"
)

type ListErrorsOptions struct {
	ProjectID string
	Status    string
	Severity  string
	Sort      string
	Direction string
	AllPages  bool
}

func (c *Client) ListErrors(opts ListErrorsOptions) ([]models.BugsnagError, bool, error) {
	path := fmt.Sprintf("/projects/%s/errors", opts.ProjectID)
	params := map[string]string{}
	if opts.Status != "" {
		params["status"] = opts.Status
	}
	if opts.Severity != "" {
		params["severity"] = opts.Severity
	}
	if opts.Sort != "" {
		params["sort"] = opts.Sort
	}
	if opts.Direction != "" {
		params["direction"] = opts.Direction
	}

	if opts.AllPages {
		items, err := CollectAllPages[models.BugsnagError](c, path, params)
		return items, false, err
	}
	return FetchSinglePage[models.BugsnagError](c, path, params)
}

func (c *Client) GetError(projectID, errorID string) (*models.BugsnagError, error) {
	path := fmt.Sprintf("/projects/%s/errors/%s", projectID, errorID)
	req, err := c.newRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	var bugsnagErr models.BugsnagError
	_, err = c.do(req, &bugsnagErr)
	if err != nil {
		return nil, err
	}
	return &bugsnagErr, nil
}
