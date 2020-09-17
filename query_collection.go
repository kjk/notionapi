package notionapi

import (
	"fmt"
)

// AggregateQuery describes an aggregate query
type AggregateQuery struct {
	// e.g. "count"
	AggregationType string `json:"aggregation_type"`
	ID              string `json:"id"`
	Property        string `json:"property"`
	// "title" is the special field that references a page
	Type string `json:"type"`
	// "table", "list"
	ViewType string `json:"view_type"`
}

// QueryFilter describes the filtering of a query
type QueryFilter struct {
	Comparator string `json:"comparator"`
	ID         string `json:"id"`
	Property   string `json:"property"`
	Type       string `json:"type"`
	Value      string `json:"value"`
}

// QuerySort describes sorting of a query
type QuerySort struct {
	ID        string `json:"id"`
	Direction string `json:"direction"`
	Property  string `json:"property"`
	Type      string `json:"type"`
}

// Aggregator describes part of the quer
type Aggregator struct {
	Aggregator string `json:"aggregator"` // e.g. "count"
	Property   string `json:"property"`   // e.g. "title"
}

// Query describes a query
type Query struct {
	Aggregate  []*AggregateQuery `json:"aggregate"`
	GroupBy    interface{}       `json:"group_by"`
	CalendarBy interface{}       `json:"calendar_by"`

	FilterOperator string         `json:"filter_operator"`
	Filter         []*QueryFilter `json:"filter"`
	Sort           []*QuerySort   `json:"sort"`
}

type Query2FilterValue struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"` // can be string or number
}

type Query2FilterFilter struct {
	Value    Query2FilterValue `json:"value"`
	Operator string            `json:"operator"`
}

type Query2FilterElement struct {
	Filter   Query2FilterFilter `json:"filter"`
	Property string             `json:"property"`
}

type Query2Filter struct {
	Filters  []Query2FilterElement `json:"filters"`
	Operator string                `json:"operator"`
}

// Query2 describes a query
type Query2 struct {
	Filter       *Query2Filter     `json:"filter"`
	Sort         []*QuerySort      `json:"sort"`
	Aggregate    []*AggregateQuery `json:"aggregate"`
	Aggregations []*Aggregator     `json:"aggregations"`
}

type loader struct {
	Type  string `json:"type"`  // e.g. "table"
	Limit int    `json:"limit"` // Notion uses 70 by default
	// from User.TimeZone
	UserTimeZone string `json:"userTimeZone"`
	// from User.Locale
	//UserLocale       string `json:"userLocale"`
	LoadContentCover bool `json:"loadContentCover"`
}

// /api/v3/queryCollection request
type queryCollectionRequest struct {
	CollectionID     string  `json:"collectionId"`
	CollectionViewID string  `json:"collectionViewId"`
	Query            *Query  `json:"query"`
	Loader           *loader `json:"loader"`
}

// AggregationResult represents result of aggregation
type AggregationResult struct {
	ID string `json:"id"`
	// TODO: maybe json.Number? Shouldn't float64 cover both?
	Value float64 `json:"value"`
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
func (c *Client) QueryCollection(collectionID, collectionViewID string, q *Query, user *User) (*QueryCollectionResponse, error) {

	// Notion has this as 70 and re-does the query if user scrolls to see more
	// of the table. We start with a bigger number because we want all the data
	// // and there seems to be no downside
	const startLimit = 256

	req := &queryCollectionRequest{
		CollectionID:     collectionID,
		CollectionViewID: collectionViewID,
		Query:            q,
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
