package notionapi

import "fmt"

// /api/v3/syncRecordValues request
type syncRecordValuesRequest struct {
	Requests []SyncRecordRequest `json:"requests"`
}

type SyncRecordRequest struct {
	Pointer struct {
		Table string `json:"table"`
		ID    string `json:"id"`
	} `json:"pointer"`
	Version int `json:"version"`
}

type RecordsByID map[string]*Record

// SyncRecordValuesResponse represents response to /api/v3/getRecordValues api
// Note: it depends on Table type in request
type SyncRecordValuesResponse struct {
	RecordMap map[string]RecordsByID `json:"recordMap"`

	// TODO: maybe temporary, we synthesize this for compat with
	// GetBlockRecords()
	Results []*Record              `json:"-"`
	RawJSON map[string]interface{} `json:"-"`
}

// SyncRecordValues executes a raw API call /api/v3/syncRecordValues
func (c *Client) SyncRecordValues(records []SyncRecordRequest) (*SyncRecordValuesResponse, error) {
	req := &syncRecordValuesRequest{
		Requests: records,
	}

	var rsp SyncRecordValuesResponse
	var err error
	apiURL := "/api/v3/syncRecordValues"
	if rsp.RawJSON, err = c.doNotionAPI(apiURL, req, &rsp); err != nil {
		return nil, err
	}

	for table, records := range rsp.RecordMap {
		for _, r := range records {
			err = parseRecord(table, r)
			rsp.Results = append(rsp.Results, r)
			if err != nil {
				return nil, err
			}
		}
	}
	return &rsp, nil
}

// SyncBlockRecords executes a raw API call /api/v3/getRecordValues
// to get records for blocks with given ids
func (c *Client) SyncBlockRecords(ids []string) (*SyncRecordValuesResponse, error) {
	records := make([]SyncRecordRequest, len(ids))
	for pos, id := range ids {
		dashID := ToDashID(id)
		if !IsValidDashID(dashID) {
			return nil, fmt.Errorf("'%s' is not a valid notion id", id)
		}
		r := &records[pos]
		r.Version = -1
		r.Pointer.Table = TableBlock
		r.Pointer.ID = dashID
	}
	return c.SyncRecordValues(records)
}
