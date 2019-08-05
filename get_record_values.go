package notionapi

import (
	"fmt"
)

// /api/v3/getRecordValues request
type getRecordValuesRequest struct {
	Requests []getRecordValuesRequestInner `json:"requests"`
}

type getRecordValuesRequestInner struct {
	Table string `json:"table"`
	ID    string `json:"id"`
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
	req := &getRecordValuesRequest{}

	for _, id := range ids {
		dashID := ToDashID(id)
		if !IsValidDashID(dashID) {
			return nil, fmt.Errorf("'%s' is not a valid notion id", id)
		}
		v := getRecordValuesRequestInner{
			Table: TableBlock,
			ID:    dashID,
		}
		req.Requests = append(req.Requests, v)
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
		brJSON := resultsJSON[i].(map[string]interface{})
		b := br.Value
		bJSON := jsonGetMap(brJSON, "value")
		b.RawJSON = bJSON
	}

	return &rsp, nil
}
