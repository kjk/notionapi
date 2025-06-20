package notionapi

import (
	"net/url"
)

const (
	// key in LoaderReducer.Reducers map
	ReducerCollectionGroupResultsName = "collection_group_results"
)

type ReducerCollectionGroupResults struct {
	Type  string `json:"type"`
	Limit int    `json:"limit"`
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
	HasMore  bool     `json:"hasMore"`
}
type ReducerResults struct {
	// TODO: probably more types
	CollectionGroupResults *CollectionGroupResults `json:"collection_group_results"`
}

// QueryCollectionResponse is json response for /api/v3/queryCollection
type QueryCollectionResponse struct {
	RecordMap *RecordMap `json:"recordMap"`
	Result    struct {
		Type     string `json:"type"`
		SizeHint int    `json:"sizeHint"` // e.g. 50
		// TODO: there's probably more
		ReducerResults *ReducerResults `json:"reducerResults"`
	} `json:"result"`
	RawJSON map[string]interface{} `json:"-"`
}

type BlockPointer struct {
	Table string `json:"table"`
	ID    string `json:"id"`
}

type QueryCollectionBlockRequest struct {
	Pointer BlockPointer `json:"pointer"`
	Version int          `json:"version"`
}

// /api/v3/syncRecordValuesSpaceInitial request
type QueryCollectionBlocksRequest struct {
	Requests []QueryCollectionBlockRequest `json:"requests"`
}

// QueryCollectionBlocksResponse is json response for /api/v3/syncRecordValuesSpaceInitial
type QueryCollectionBlocksResponse struct {
	RecordMap *RecordMap             `json:"recordMap"`
	RawJSON   map[string]interface{} `json:"-"`
}

type QueryPageShortIdRequest struct {
	BlockID                   string `json:"blockId"`                   // e.g. "20209ec4-150f-80f7-9f2f-d3a8965849ef"
	Name                      string `json:"name"`                      // e.g. "page"
	Type                      string `json:"type"`                      // e.g. "block-space"
	RequestedOnPublicDomain   bool   `json:"requestedOnPublicDomain"`   // e.g. false
	CollectionViewID          string `json:"collectionViewId"`          // e.g. "20209ec4-150f-81be-9839-000c1487315f"
	ShowMoveTo                bool   `json:"showMoveTo"`                // e.g. false
	SaveParent                bool   `json:"saveParent"`                // e.g. false
	ShouldDuplicate           bool   `json:"shouldDuplicate"`           // e.g. false
	ProjectManagementLaunch   bool   `json:"projectManagementLaunch"`   // e.g. false
	ConfigureOpenInDesktopApp bool   `json:"configureOpenInDesktopApp"` // e.g. false
	MobileData                struct {
		IsPush bool `json:"isPush"` // e.g. false
	} `json:"mobileData"`
	DemoWorkspaceMode bool `json:"demoWorkspaceMode"` // e.g. false
}

type QueryPageShortIdResponse struct {
	PageId       string                 `json:"pageId"`       // e.g. "20209ec4-150f-80f7-9f2f-d3a8965849ef"
	SpaceName    string                 `json:"spaceName"`    // e.g. "Demo"
	SpaceId      string                 `json:"spaceId"`      // e.g. "adcf8e8f-7e37-4d4b-97aa-5f2e26797a28"
	BetaEnabled  bool                   `json:"betaEnabled"`  // e.g. false
	SpaceDomain  string                 `json:"spaceDomain"`  // e.g. "wild-join-91f"
	SpaceShortId string                 `json:"spaceShortId"` // e.g. "2663650575"
	RawJSON      map[string]interface{} `json:"-"`
}

type LoaderReducer struct {
	Type         string                 `json:"type"` //"reducer"
	Reducers     map[string]interface{} `json:"reducers"`
	Sort         []QuerySort            `json:"sort,omitempty"`
	Filter       map[string]interface{} `json:"filter,omitempty"`
	SearchQuery  string                 `json:"searchQuery"`
	UserTimeZone string                 `json:"userTimeZone"` // e.g. "America/Los_Angeles" from User.Locale
}

func MakeLoaderReducer(query *Query, limits ...int) *LoaderReducer {
	res := &LoaderReducer{
		Type:     "reducer",
		Reducers: map[string]interface{}{},
	}
	if query != nil {
		res.Sort = query.Sort
		res.Filter = query.Filter
	}
	limit := 50
	if len(limits) > 0 {
		limit = limits[0]
	}
	res.Reducers[ReducerCollectionGroupResultsName] = &ReducerCollectionGroupResults{
		Type:  "results",
		Limit: limit,
	}
	// set some default value, should over-ride with User.TimeZone
	res.UserTimeZone = "America/Los_Angeles"
	return res
}

// QueryCollection executes a raw API call /api/v3/queryCollection
func (c *Client) QueryCollection(req QueryCollectionRequest, query *Query, params ...map[string]string) (*QueryCollectionResponse, error) {
	if req.Loader == nil {
		req.Loader = MakeLoaderReducer(query)
	}
	var rsp QueryCollectionResponse
	var err error
	values := url.Values{}
	for _, p := range params {
		for k, v := range p {
			values.Add(k, v)
		}
	}
	apiURL := "/api/v3/queryCollection"
	if len(values) > 0 {
		apiURL += "?" + values.Encode()
	}
	err = c.doNotionAPI(apiURL, req, &rsp, &rsp.RawJSON)
	if err != nil {
		return nil, err
	}
	// TODO: fetch more if exceeded limit
	if err := ParseRecordMap(rsp.RecordMap); err != nil {
		return nil, err
	}
	return &rsp, nil
}

// QuerySpaceShortId executes a raw API call /api/v3/getPublicPageData
func (c *Client) QuerySpaceShortId(pageId string, collectionViewID string) (*QueryPageShortIdResponse, error) {
	req := QueryPageShortIdRequest{
		BlockID:                   pageId,
		Name:                      "page",
		Type:                      "block-space",
		RequestedOnPublicDomain:   false,
		CollectionViewID:          collectionViewID,
		ShowMoveTo:                false,
		SaveParent:                false,
		ShouldDuplicate:           false,
		ProjectManagementLaunch:   false,
		ConfigureOpenInDesktopApp: false,
		MobileData: struct {
			IsPush bool `json:"isPush"` // e.g. false
		}{
			IsPush: false,
		},
		DemoWorkspaceMode: false,
	}

	var rsp QueryPageShortIdResponse
	var err error
	apiURL := "/api/v3/getPublicPageData"
	err = c.doNotionAPI(apiURL, req, &rsp, &rsp.RawJSON)
	if err != nil {
		return nil, err
	}
	return &rsp, nil
}
