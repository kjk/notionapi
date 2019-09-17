package notionapi

// /api/v3/queryCollection request
type queryCollectionRequest struct {
	CollectionID     string           `json:"collectionId"`
	CollectionViewID string           `json:"collectionViewId"`
	Query            *CollectionQuery `json:"query"`
	Loader           *Loader          `json:"loader"`
}

// Loader describes a loader
/*
  "loader": {
    "type": "table",
    "limit": 70,
    "userTimeZone": "America/Los_Angeles",
    "userLocale": "en"
  }
*/
type Loader struct {
	Type  string `json:"type"`
	Limit int    `json:"limit"`
	// from User.TimeZone
	UserTimeZone string `json:"userTimeZone"`
	// from User.Locale
	UserLocale string `json:"userLocale"`
}

// CollectionQuery describes a collection query
type CollectionQuery struct {
	// copy from CollectionView.Query
	Aggregate  []*AggregateQuery `json:"aggregate"`
	GroupBy    interface{}       `json:"group_by"`
	CalendarBy interface{}       `json:"calendar_by"`
	// "and"
	FilterOperator string                 `json:"filter_operator"`
	Filter         []interface{}          `json:"filter"`
	Sort           []*CollectionQuerySort `json:"sort"`
}

type CollectionQuerySort struct {
	// "ascending"
	Direction string `json:"direction"`
	ID        string `json:"id"`
	Property  string `json:"property"`
	Type      string `json:"type"`
}

// QueryCollectionResponse is json response for /api/v3/queryCollection
type QueryCollectionResponse struct {
	RecordMap *RecordMap             `json:"recordMap"`
	Result    *QueryCollectionResult `json:"result"`
	RawJSON   map[string]interface{} `json:"-"`
}

// QueryCollectionResult is part of response for /api/v3/queryCollection
type QueryCollectionResult struct {
	Type               string               `json:"type"`
	BlockIDS           []string             `json:"blockIds"`
	AggregationResults []*AggregationResult `json:"aggregationResults"`
	Total              int                  `json:"total"`
}

// AggregationResult represents result of aggregation
type AggregationResult struct {
	ID    string `json:"id"`
	Value int64  `json:"value"`
}

// QueryCollection executes a raw API call /api/v3/queryCollection
func (c *Client) QueryCollection(collectionID, collectionViewID string, aggregateQuery []*AggregateQuery, user *User) (*QueryCollectionResponse, error) {
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
	req.Loader = &Loader{
		Type:         "table",
		Limit:        70,
		UserLocale:   user.Locale,
		UserTimeZone: user.TimeZone,
	}

	apiURL := "/api/v3/queryCollection"
	var rsp QueryCollectionResponse
	var err error
	rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}
