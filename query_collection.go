package notionapi

const (
	// key in LoaderReducer.Reducers map
	ReducerCollectionGroupResultsName = "collection_group_results"
)

type ReducerCollectionGroupResults struct {
	Type  string `json:"type"`
	Limit int    `json:"limit"`
}

type LoaderReducer struct {
	Type         string                 `json:"type"` //"reducer"
	Reducers     map[string]interface{} `json:"reducers"`
	Sort         []*QuerySort           `json:"sort"`
	SearchQuery  string                 `json:"searchQuery"`
	UserTimeZone string                 `json:"userTimeZone"` // e.g. "America/Los_Angeles" from User.Locale
}

func MakeLoaderReducer(sort *QuerySort) *LoaderReducer {
	res := &LoaderReducer{
		Type:     "reducer",
		Reducers: map[string]interface{}{},
		Sort:     []*QuerySort{sort},
	}
	res.Reducers[ReducerCollectionGroupResultsName] = &ReducerCollectionGroupResults{
		Type:  "results",
		Limit: 50,
	}
	// set some default value, should over-ride with User.TimeZone
	res.UserTimeZone = "America/Los_Angeles"
	return res
}

// /api/v3/queryCollection request
type QueryCollectionRequest struct {
	Collection struct {
		ID      string `json:"id"`
		SpaceID string `json:"spaceId"`
	} `json:"collection"`
	CollectionView struct {
		ID      string `json:"id"`
		SpaceID string `json:"spaceId"`
	} `json:"collectionView"`
	Loader interface{} `json:"loader"` // e.g. LoaderReducer
}

type CollectionGroupResults struct {
	Type     string   `json:"type"`
	BlockIds []string `json:"blockIds"`
	Total    int      `json:"total"`
}
type ReducerResults struct {
	// TODO: probably more types
	CollectionGroupResults *CollectionGroupResults `json:"collection_group_results"`
}

// QueryCollectionResponse is json response for /api/v3/queryCollection
type QueryCollectionResponse struct {
	RecordMap *RecordMap `json:"recordMap"`
	Result    struct {
		Type string `json:"type"`
		// TODO: there's probably more
		ReducerResults *ReducerResults `json:"reducerResults"`
	} `json:"result"`
	RawJSON map[string]interface{} `json:"-"`
}

// QueryCollection executes a raw API call /api/v3/queryCollection
func (c *Client) QueryCollection(req QueryCollectionRequest, sort *QuerySort) (*QueryCollectionResponse, error) {
	req.Loader = MakeLoaderReducer(sort)
	var rsp QueryCollectionResponse
	var err error
	apiURL := "/api/v3/queryCollection"
	rsp.RawJSON, err = c.doNotionAPI(apiURL, req, &rsp)
	if err != nil {
		return nil, err
	}
	// TODO: fetch more if exceeded limit
	if err := ParseRecordMap(rsp.RecordMap); err != nil {
		return nil, err
	}
	return &rsp, nil
}
