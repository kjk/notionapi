package notionapi

import (
	"encoding/json"
	"fmt"
)

// /api/v3/getRecordValues request
type getRecordValuesRequest struct {
	Requests []RecordValueRequest `json:"requests"`
}

type RecordValueRequest struct {
	Table string `json:"table"`
	ID    string `json:"id"`
}

type ValueResponse struct {
	ID    string `json:"id"`
	Table string `json:"table"`
	Role  string `json:"role"`

	Value json.RawMessage `json:"value"`

	Block *Block `json:"-"`
	Space *Space `json:"-"`
	User  *User  `json:"-"`
}

// BlockWithRole describes a block info
type BlockWithRole struct {
	Role  string `json:"role"`
	Value *Block `json:"value"`
}

// GetRecordValuesResponse represents response to /api/v3/getRecordValues api
// Note: it depends on Table type in request
type GetRecordValuesResponse struct {
	Results []*BlockWithRole       `json:"results"`
	RawJSON map[string]interface{} `json:"-"`
}

// GetRecordValues executes a raw API call /api/v3/getRecordValues
func (c *Client) GetRecordValues(ids []string) (*GetRecordValuesResponse, error) {
	requests := make([]RecordValueRequest, len(ids))

	for pos, id := range ids {
		dashID := ToDashID(id)
		if !IsValidDashID(dashID) {
			return nil, fmt.Errorf("'%s' is not a valid notion id", id)
		}
		requests[pos].Table = TableBlock
		requests[pos].ID = dashID
	}

	req := &getRecordValuesRequest{
		Requests: requests,
	}

	apiURL := "/api/v3/getRecordValues"
	var rsp GetRecordValuesResponse
	var err error
	rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp)
	if err != nil {
		return nil, err
	}
	resultsJSON := rsp.RawJSON["results"].([]interface{})
	for i, br := range rsp.Results {
		b := br.Value
		if b == nil {
			continue
		}
		brJSON := resultsJSON[i].(map[string]interface{})
		bJSON := jsonGetMap(brJSON, "value")
		b.RawJSON = bJSON
	}

	return &rsp, nil
}

func (c *Client) RequestRecordValues(requests []RecordValueRequest) ([]ValueResponse, error) {
	req := &getRecordValuesRequest{
		Requests: requests,
	}

	apiURL := "/api/v3/getRecordValues"
	var rsp struct {
		Results []ValueResponse `json:"results"`
	}
	var err error
	if _, err = doNotionAPI(c, apiURL, req, &rsp); err != nil {
		return nil, err
	}

	for pos := range rsp.Results {
		rsp.Results[pos].Table = requests[pos].Table
		var obj interface{}
		if requests[pos].Table == TableUser {
			rsp.Results[pos].User = &User{}
			obj = rsp.Results[pos].User
		}
		if requests[pos].Table == TableBlock {
			rsp.Results[pos].Block = &Block{}
			obj = rsp.Results[pos].Block
		}
		if requests[pos].Table == TableSpace {
			rsp.Results[pos].Space = &Space{}
			obj = rsp.Results[pos].Space
		}
		if obj != nil {
			if err := json.Unmarshal(rsp.Results[pos].Value, &obj); err != nil {
				return nil, err
			}
		}
	}

	return rsp.Results, nil
}
