package notionapi

import "encoding/json"

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
	FilterOperator string        `json:"filter_operator"`
	Filter         []interface{} `json:"filter"`
	Sort           []interface{} `json:"sort"`
}

// /api/v3/queryCollection response

type queryCollectionResponse struct {
	RecordMap recordMap              `json:"recordMap"`
	Result    *queryCollectionResult `json:"result"`
}

type queryCollectionResult struct {
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

func parseQueryCollectionRequest(d []byte) (*queryCollectionResponse, error) {
	var rsp queryCollectionResponse
	err := json.Unmarshal(d, &rsp)
	if err != nil {
		dbg("parseQueryCollectionRequest: json.Unmarshal() failed with '%s'\n", err)
		return nil, err
	}
	return &rsp, nil
}

func apiQueryCollection(collectionID, collectionViewID string, aggregateQuery []*AggregateQuery, user *User) (*queryCollectionResponse, error) {
	req := &queryCollectionRequest{
		CollectionID:     collectionID,
		CollectionViewID: collectionViewID,
	}
	req.Query = &CollectionQuery{
		Aggregate:      aggregateQuery,
		FilterOperator: "and",
	}
	req.Loader = &Loader{
		Type:         "table",
		Limit:        70,
		UserLocale:   user.Locale,
		UserTimeZone: user.TimeZone,
	}

	apiURL := "/api/v3/queryCollection"
	var rsp *queryCollectionResponse
	parse1 := func(d []byte) error {
		var err error
		rsp, err = parseQueryCollectionRequest(d)
		return err
	}
	err := doNotionAPI(apiURL, req, parse1)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}
