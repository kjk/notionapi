package notionapi

import (
	"errors"
)

// CreateEmailUser invites a new user through his email address
func (c *Client) CreateEmailUser(email string) (*User, error) {
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
	_, err := doNotionAPI(c, apiURL, req, &rsp)

	recordMap := rsp.RecordMap
	ParseRecordMap(recordMap)
	users, ok := recordMap.Users[rsp.UserID]
	if !ok {
		return nil, errors.New("error inviting user")
	}

	return users.User, err
}
