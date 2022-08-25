package notionapi

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
func (c *Client) SyncRecordValues(req syncRecordRequest) (*SyncRecordValuesResponse, error) {
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

// GetBlockRecords emulates deprecated /api/v3/getRecordValues with /api/v3/syncRecordValues
// Gets Block records with given ids
// Used to retrieve version information for each block so that we can skip re-downloading pages
// that didn't change
func (c *Client) GetBlockRecords(ids []string) ([]*Block, error) {
	var req syncRecordRequest
	for _, id := range ids {
		id = ToDashID(id)
		p := Pointer{
			ID:    id,
			Table: TableBlock,
		}
		pver := PointerWithVersion{
			Pointer: p,
			Version: -1,
		}
		req.Requests = append(req.Requests, pver)
	}

	rsp, err := c.SyncRecordValues(req)
	if err != nil {
		return nil, err
	}
	var res []*Block
	rm := rsp.RecordMap
	for _, id := range ids {
		id = ToDashID(id)

		// sometimes notion does not return the block ask by the API
		var b *Block
		if rm.Blocks[id] != nil {
			b = rm.Blocks[id].Block
		}

		res = append(res, b)
	}
	return res, nil
}
