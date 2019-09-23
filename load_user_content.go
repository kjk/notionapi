package notionapi

import (
	"encoding/json"
)

type LoadUserResponse struct {
	ID    string `json:"id"`
	Table string `json:"table"`
	Role  string `json:"role"`

	Value json.RawMessage `json:"value"`

	Block *Block `json:"-"`
	Space *Space `json:"-"`
	User  *User  `json:"-"`

	RawJSON map[string]interface{} `json:"-"`
}

func (c *Client) LoadUserContent() (*LoadUserResponse, error) {
	req := struct{}{}

	apiURL := "/api/v3/loadUserContent"
	var rsp struct {
		RecordMap map[string]map[string]*LoadUserResponse `json:"recordMap"`
	}
	rawJSON, err := doNotionAPI(c, apiURL, req, &rsp)
	if err != nil {
		return nil, err
	}

	result := &LoadUserResponse{
		RawJSON: rawJSON,
	}

	for table, values := range rsp.RecordMap {
		for _, value := range values {
			var obj interface{}
			if table == TableUser {
				result.User = &User{}
				obj = result.User
			}
			if table == TableBlock {
				result.Block = &Block{}
				obj = result.Block
			}
			if table == TableSpace {
				result.Space = &Space{}
				obj = result.Space
			}
			if obj == nil {
				continue
			}
			if err := json.Unmarshal(value.Value, &obj); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}
