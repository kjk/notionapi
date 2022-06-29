package notionapi

import (
	"errors"
)

// CreateEmailUser invites a new user through his email address
func (c *Client) CreateEmailUser(email string) (*NotionUser, error) {
	req := struct {
		Email string `json:"email"`
	}{
		Email: email,
	}

	var rsp struct {
		UserID    string     `json:"userId"`
		RecordMap *RecordMap `json:"recordMap"`
	}

	apiURL := "/api/v3/createEmailUser"
	err := c.doNotionAPI(apiURL, req, &rsp, nil)

	recordMap := rsp.RecordMap
	ParseRecordMap(recordMap)
	users, ok := recordMap.NotionUsers[rsp.UserID]
	if !ok {
		return nil, errors.New("error inviting user")
	}

	return users.Value.NotionUser, err
}
