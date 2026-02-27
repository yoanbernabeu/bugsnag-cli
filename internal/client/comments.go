package client

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/yoanbernabeu/bugsnag-cli/internal/models"
)

func (c *Client) ListComments(projectID, errorID string, allPages bool) ([]models.Comment, bool, error) {
	path := fmt.Sprintf("/projects/%s/errors/%s/comments", projectID, errorID)
	if allPages {
		items, err := CollectAllPages[models.Comment](c, path, nil)
		return items, false, err
	}
	return FetchSinglePage[models.Comment](c, path, nil)
}

func (c *Client) CreateComment(projectID, errorID, message string) (*models.Comment, error) {
	path := fmt.Sprintf("/projects/%s/errors/%s/comments", projectID, errorID)
	body := map[string]string{"message": message}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := c.newRequestWithBody("POST", path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	var comment models.Comment
	_, err = c.do(req, &comment)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}
