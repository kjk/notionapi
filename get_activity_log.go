package notionapi

// /api/v3/getActivityLog request
type getActivityLogRequest struct {
	SpaceID         string `json:"spaceId"`
	StartingAfterID string `json:"startingAfterId,omitempty"`
	Limit           int    `json:"limit"`
}

// LoadPageChunkResponse is a response to /api/v3/loadPageChunk api
type GetActivityLogResponse struct {
	ActivityIDs []string   `json:"activityIds"`
	RecordMap   *RecordMap `json:"recordMap"`
	NextID      string     `json:"-"`

	RawJSON map[string]interface{} `json:"-"`
}

// GetActivityLog executes a raw API call /api/v3/getActivityLog.
// If startingAfterId is "", starts at the most recent log entry.
func (c *Client) GetActivityLog(spaceID string, startingAfterID string, limit int) (*GetActivityLogResponse, error) {
	apiURL := "/api/v3/getActivityLog"
	req := &getActivityLogRequest{
		SpaceID:         spaceID,
		StartingAfterID: startingAfterID,
		Limit:           limit,
	}
	var rsp GetActivityLogResponse
	var err error
	if rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp); err != nil {
		return nil, err
	}
	if err = ParseRecordMap(rsp.RecordMap); err != nil {
		return nil, err
	}
	if len(rsp.ActivityIDs) > 0 {
		rsp.NextID = rsp.ActivityIDs[len(rsp.ActivityIDs)-1]
	} else {
		rsp.NextID = ""
	}
	return &rsp, nil
}
