package notionapi

type navigableBlockID struct {
	ID string `json:"id"`
}

// /api/v3/getActivityLog request
type getActivityLogRequest struct {
	SpaceID         string           `json:"spaceId"`
	StartingAfterID string           `json:"startingAfterId,omitempty"`
	NavigableBlock  navigableBlockID `json:"navigableBlock,omitempty"`
	Limit           int              `json:"limit"`
}

// GetActivityLogResponse is a response to /api/v3/getActivityLog api
type GetActivityLogResponse struct {
	ActivityIDs []string   `json:"activityIds"`
	RecordMap   *RecordMap `json:"recordMap"`
	NextID      string     `json:"-"`

	RawJSON map[string]interface{} `json:"-"`
}

// GetActivityLog executes a raw API call /api/v3/getActivityLog.
// If startingAfterId is "", starts at the most recent log entry.
// navBlockID is the ID of a navigable block (like a page in a database)
func (c *Client) GetActivityLog(spaceID string, startingAfterID string, navBlockID string, limit int) (*GetActivityLogResponse, error) {
	req := &getActivityLogRequest{
		SpaceID:         spaceID,
		StartingAfterID: startingAfterID,
		Limit:           limit,
		NavigableBlock:  navigableBlockID{ID: navBlockID},
	}
	var rsp GetActivityLogResponse
	var err error
	apiURL := "/api/v3/getActivityLog"
	if err = c.doNotionAPI(apiURL, req, &rsp, &rsp.RawJSON); err != nil {
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
