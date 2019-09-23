package notionapi

import (
	"errors"
)

// CreateEmailUser invites a new user through his email address
func (c *Client) CreateEmailUser(email string) (*UserWithRole, error) {
	req := struct {
		Email string `json:"email"`
	}{
		Email: email,
	}

	var rsp struct {
		UserID    string `json:"userId"`
		RecordMap struct {
			NotionUser map[string]UserWithRole `json:"notion_user"`
		} `json:"recordMap"`
	}

	apiURL := "/api/v3/createEmailUser"
	_, err := doNotionAPI(c, apiURL, req, &rsp)

	users, ok := rsp.RecordMap.NotionUser[rsp.UserID]
	if !ok {
		return nil, errors.New("error inviting user")
	}

	return &users, err
}
