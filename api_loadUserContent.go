package notionapi

import "encoding/json"

type LoadUserResponse struct {
	ID    string `json:"id"`
	Table string `json:"table"`
	Role  string `json:"role"`

	Value json.RawMessage `json:"value"`

	Block *Block      `json:"-"`
	Space *Space      `json:"-"`
	User  *NotionUser `json:"-"`

	RawJSON map[string]interface{} `json:"-"`
}

func (c *Client) LoadUserContent() (*LoadUserResponse, error) {
	req := struct{}{}

	var rsp struct {
		RecordMap map[string]map[string]*LoadUserResponse `json:"recordMap"`
	}
	apiURL := "/api/v3/loadUserContent"
	result := LoadUserResponse{}

	err := c.doNotionAPI(apiURL, req, &rsp, &result.RawJSON)
	if err != nil {
		return nil, err
	}

	for table, values := range rsp.RecordMap {
		for _, value := range values {
			var obj interface{}
			if table == TableNotionUser {
				result.User = &NotionUser{}
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
			if err := jsonit.Unmarshal(value.Value, &obj); err != nil {
				return nil, err
			}
		}
	}

	return &result, nil
}
