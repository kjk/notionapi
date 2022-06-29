package notionapi

type permissionRecord struct {
	ID      string `json:"id"`
	Table   string `json:"table"`
	SpaceID string `json:"spaceId"`
}

type signedURLRequest struct {
	URL              string            `json:"url"`
	PermissionRecord *permissionRecord `json:"permissionRecord"`
}

// /api/v3/getSignedFileUrls request
type getSignedFileURLsRequest struct {
	URLs []signedURLRequest `json:"urls"`
}

// GetSignedURLsResponse represents response to /api/v3/getSignedFileUrls api
// Note: it depends on Table type in request
type GetSignedURLsResponse struct {
	SignedURLS []string               `json:"signedUrls"`
	RawJSON    map[string]interface{} `json:"-"`
}

// GetSignedURLs executes a raw API call /api/v3/getSignedFileUrls
func (c *Client) GetSignedURLs(urls []string, block *Block) (*GetSignedURLsResponse, error) {
	permRec := &permissionRecord{
		ID:      block.ID,
		Table:   block.ParentTable,
		SpaceID: block.SpaceID,
	}
	var recs []signedURLRequest
	for _, url := range urls {
		srec := signedURLRequest{
			URL:              url,
			PermissionRecord: permRec,
		}
		recs = append(recs, srec)
	}
	req := &getSignedFileURLsRequest{
		URLs: recs,
	}

	var rsp GetSignedURLsResponse
	var err error
	apiURL := "/api/v3/getSignedFileUrls"
	if rsp.RawJSON, err = c.doNotionAPI(apiURL, req, &rsp); err != nil {
		return nil, err
	}
	return &rsp, nil
}
