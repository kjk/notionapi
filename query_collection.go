package notionapi

import (
	"encoding/json"
	"fmt"
)

type loader struct {
	Type  string `json:"type"`  // e.g. "table"
	Limit int    `json:"limit"` // Notion uses 70 by default
	// from User.TimeZone
	UserTimeZone string `json:"userTimeZone"`
	// from User.Locale
	//UserLocale       string `json:"userLocale"`
	LoadContentCover bool `json:"loadContentCover"`
	// TODO: searchQuery
}

// /api/v3/queryCollection request
type queryCollectionRequest struct {
	CollectionID     string          `json:"collectionId"`
	CollectionViewID string          `json:"collectionViewId"`
	Query2           json.RawMessage `json:"query"`
	Loader           *loader         `json:"loader"`
}

// AggregationResult represents result of aggregation
type AggregationResult struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	// TODO: maybe json.Number? Shouldn't float64 cover both?
	// When type is equal to date, value is an object.
	Value interface{} `json:"value"`
}

// QueryCollectionResult is part of response for /api/v3/queryCollection
type QueryCollectionResult struct {
	Type               string               `json:"type"`
	BlockIDS           []string             `json:"blockIds"`
	AggregationResults []*AggregationResult `json:"aggregationResults"`
	Total              int                  `json:"total"`
}

// QueryCollectionResponse is json response for /api/v3/queryCollection
type QueryCollectionResponse struct {
	RecordMap *RecordMap             `json:"recordMap"`
	Result    *QueryCollectionResult `json:"result"`
	RawJSON   map[string]interface{} `json:"-"`
}

// QueryCollection executes a raw API call /api/v3/queryCollection
func (c *Client) QueryCollection(collectionID, collectionViewID string, q json.RawMessage, user *User) (*QueryCollectionResponse, error) {

	// Notion has this as 70 and re-does the query if user scrolls to see more
	// of the table. We start with a bigger number because we want all the data
	// // and there seems to be no downside
	const startLimit = 256

	req := &queryCollectionRequest{
		CollectionID:     collectionID,
		CollectionViewID: collectionViewID,
		Query2:           q,
	}
	timeZone := "America/Los_Angeles"
	if user != nil {
		timeZone = user.TimeZone
	}
	req.Loader = &loader{
		Type:         "table",
		Limit:        startLimit,
		UserTimeZone: timeZone,
		// don't know what this is, Notion sets it to true
		LoadContentCover: true,
	}

	apiURL := "/api/v3/queryCollection"
	var rsp QueryCollectionResponse
	var err error
	rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp)
	if err != nil {
		return nil, err
	}

	// fetch everything if a collection has more rows
	// than we originally asked for
	actualTotal := rsp.Result.Total
	if actualTotal > startLimit {
		rsp = QueryCollectionResponse{}
		req.Loader.Limit = actualTotal
		rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp)
		if err != nil {
			return nil, fmt.Errorf("Client.QueryCollection() 2nd fetch failed: %s", err)
		}
	}
	if err := ParseRecordMap(rsp.RecordMap); err != nil {
		return nil, err
	}
	return &rsp, nil
}
