package notionapi

import (
	"fmt"
)

// /api/v3/getRecordValues request
type getRecordValuesRequest struct {
	Requests []RecordRequest `json:"requests"`
}

// RecordRequest represents argument to GetRecordValues
type RecordRequest struct {
	Table string `json:"table"`
	ID    string `json:"id"`
}

// GetRecordValuesResponse represents response to /api/v3/getRecordValues api
// Note: it depends on Table type in request
type GetRecordValuesResponse struct {
	Results []*Record              `json:"results"`
	RawJSON map[string]interface{} `json:"-"`
}

// GetBlockRecords executes a raw API call /api/v3/getRecordValues
// to get records for blocks with given ids
func (c *Client) GetBlockRecords(ids []string) (*GetRecordValuesResponse, error) {
	records := make([]RecordRequest, len(ids))
	for pos, id := range ids {
		dashID := ToDashID(id)
		if !IsValidDashID(dashID) {
			return nil, fmt.Errorf("'%s' is not a valid notion id", id)
		}
		records[pos].Table = TableBlock
		records[pos].ID = dashID
	}
	return c.GetRecordValues(records)
}

// GetRecordValues executes a raw API call /api/v3/getRecordValues
func (c *Client) GetRecordValues(records []RecordRequest) (*GetRecordValuesResponse, error) {
	req := &getRecordValuesRequest{
		Requests: records,
	}

	var rsp GetRecordValuesResponse
	var err error
	apiURL := "/api/v3/getRecordValues"
	if rsp.RawJSON, err = c.doNotionAPI(apiURL, req, &rsp); err != nil {
		return nil, err
	}

	for idx, r := range rsp.Results {
		table := records[idx].Table
		err = parseRecord(table, r)
		if err != nil {
			return nil, err
		}
	}

	return &rsp, nil
}
