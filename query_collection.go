package notionapi

import "fmt"

// /api/v3/queryCollection request
type queryCollectionRequest struct {
	CollectionID     string           `json:"collectionId"`
	CollectionViewID string           `json:"collectionViewId"`
	Query            *CollectionQuery `json:"query"`
	Loader           *loader          `json:"loader"`
}

type loader struct {
	Type  string `json:"type"`  // e.g. "table"
	Limit int    `json:"limit"` // Notion uses 70 by default
	// from User.TimeZone
	UserTimeZone string `json:"userTimeZone"`
	// from User.Locale
	UserLocale       string `json:"userLocale"`
	LoadContentCover bool   `json:"loadContentCover"`
}

// Query describes a query
// TODO: merge with CollectionQuery
type Query struct {
	Aggregate []*AggregateQuery `json:"aggregate"`

	FilterOperator string         `json:"filter_operator"`
	Filter         []*FilterQuery `json:"filter"`
	Sort           []*SortQuery   `json:"sort"`
}

// CollectionQuery describes a collection query
// TODO: merge with Query
type CollectionQuery struct {
	// copy from CollectionView.Query
	Aggregate  []*AggregateQuery `json:"aggregate"`
	GroupBy    interface{}       `json:"group_by"`
	CalendarBy interface{}       `json:"calendar_by"`

	// "and"
	FilterOperator string         `json:"filter_operator"` // e.g. "and"
	Filter         []*FilterQuery `json:"filter"`
	Sort           []*SortQuery   `json:"sort"`
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

// AggregationResult represents result of aggregation
type AggregationResult struct {
	ID string `json:"id"`
	// TODO: maybe json.Number? Shouldn't float64 cover both?
	Value float64 `json:"value"`
}

// QueryCollection executes a raw API call /api/v3/queryCollection
func (c *Client) QueryCollection(collectionID, collectionViewID string, aggregateQuery []*AggregateQuery, user *User) (*QueryCollectionResponse, error) {

	// Notion has this as 70 and re-does the query if user scrolls to see more
	// of the table. We start with a bigger number because we want all the data
	// // and there seems to be no downside
	const startLimit = 256

	req := &queryCollectionRequest{
		CollectionID:     collectionID,
		CollectionViewID: collectionViewID,
	}
	if aggregateQuery != nil {
		req.Query = &CollectionQuery{
			Aggregate:      aggregateQuery,
			FilterOperator: "and",
		}
	}
	req.Loader = &loader{
		Type:         "table",
		Limit:        startLimit,
		UserLocale:   user.Locale,
		UserTimeZone: user.TimeZone,
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
	if err := parseRecordMap(rsp.RecordMap); err != nil {
		return nil, err
	}
	return &rsp, nil
}
