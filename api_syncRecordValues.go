package notionapi

import "fmt"

// /api/v3/syncRecordValues request
type syncRecordRequest struct {
	Requests []PointerWithVersion `json:"requests"`
}

type Pointer struct {
	Table string `json:"table"`
	ID    string `json:"id"`
}

type PointerWithVersion struct {
	Pointer Pointer `json:"pointer"`
	Version int     `json:"version"`
}

// SyncRecordValuesResponse represents response to /api/v3/syncRecordValues api
// Note: it depends on Table type in request
type SyncRecordValuesResponse struct {
	RecordMap *RecordMap `json:"recordMap"`

	RawJSON map[string]interface{} `json:"-"`
}

// SyncRecordValues executes a raw API call /api/v3/syncRecordValues
func (c *Client) SyncRecordValuesAPI(req syncRecordRequest) (*SyncRecordValuesResponse, error) {
	var rsp SyncRecordValuesResponse
	var err error
	apiURL := "/api/v3/syncRecordValues"
	if err = c.doNotionAPI(apiURL, req, &rsp, &rsp.RawJSON); err != nil {
		return nil, err
	}
	if err = ParseRecordMap(rsp.RecordMap); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// SyncBlockValues executes a raw API call /api/v3/syncRecordValues
// to get records for blocks with given ids
func (c *Client) SyncBlockValues(ids []string) (*SyncRecordValuesResponse, error) {
	var req syncRecordRequest
	for _, id := range ids {
		dashID := ToDashID(id)
		if !IsValidDashID(dashID) {
			return nil, fmt.Errorf("'%s' is not a valid notion id", id)
		}
		p := Pointer{
			Table: TableBlock,
			ID:    dashID,
		}
		pver := PointerWithVersion{
			Pointer: p,
			Version: -1,
		}
		req.Requests = append(req.Requests, pver)
	}
	return c.SyncRecordValuesAPI(req)
}

// GetRecordValuesResponse represents response to /api/v3/getRecordValues api
// Note: it depends on Table type in request
type GetRecordValuesResponse struct {
	Results []*Record              `json:"results"`
	RawJSON map[string]interface{} `json:"-"`
}

// GetRecordValues emulates deprecated /api/v3/getRecordValues with  /api/v3/syncRecordValues
func (c *Client) GetRecordValues(records []Pointer) (*GetRecordValuesResponse, error) {

	var req syncRecordRequest
	for _, p := range records {
		pver := PointerWithVersion{
			Pointer: p,
			Version: -1,
		}
		req.Requests = append(req.Requests, pver)
	}

	srsp, err := c.SyncRecordValuesAPI(req)
	if err != nil {
		return nil, err
	}

	var rsp GetRecordValuesResponse
	rm := srsp.RecordMap
	for _, rv := range rm.Blocks {
		rsp.Results = append(rsp.Results, rv.Value)
	}

	return &rsp, nil
}

// GetBlockRecords executes a raw API call /api/v3/getRecordValues
// to get records for blocks with given ids
func (c *Client) GetBlockRecords(ids []string) (*GetRecordValuesResponse, error) {
	records := make([]Pointer, len(ids))
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
