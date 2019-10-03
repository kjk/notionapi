package notionapi

// /api/v3/loadPageChunk request
type loadPageChunkRequest struct {
	PageID          string `json:"pageId"`
	ChunkNumber     int    `json:"chunkNumber"`
	Limit           int    `json:"limit"`
	Cursor          cursor `json:"cursor"`
	VerticalColumns bool   `json:"verticalColumns"`
}

type cursor struct {
	Stack [][]stack `json:"stack"`
}

type stack struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
	Table string `json:"table"`
}

// LoadPageChunkResponse is a response to /api/v3/loadPageChunk api
type LoadPageChunkResponse struct {
	RecordMap *RecordMap `json:"recordMap"`
	Cursor    cursor     `json:"cursor"`

	RawJSON map[string]interface{} `json:"-"`
}

// RecordMap contains a collections of blocks, a space, users, and collections.
type RecordMap struct {
	Blocks          map[string]*Record `json:"block"`
	Spaces          map[string]*Record `json:"space"`
	Users           map[string]*Record `json:"notion_user"`
	Collections     map[string]*Record `json:"collection"`
	CollectionViews map[string]*Record `json:"collection_view"`
	Comments        map[string]*Record `josn:"comment"`
	Discussions     map[string]*Record `jsoon:"discussion"`
}

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

// FilterQuery describes the filtering of a query
// TODO: rename QueryFilter
type FilterQuery struct {
	Comparator string `json:"comparator"`
	ID         string `json:"id"`
	Property   string `json:"property"`
	Type       string `json:"type"`
	Value      string `json:"value"`
}

// SortQuery describes the sorting of a query
// TODO: rename QuerySort
type SortQuery struct {
	ID        string `json:"id"`
	Direction string `json:"direction"`
	Property  string `json:"property"`
	Type      string `json:"type"`
}

// CollectionFormat describes format of a collection
type CollectionFormat struct {
	CoverPosition  float64                   `json:"collection_cover_position"`
	PageProperties []*CollectionPageProperty `json:"collection_page_properties"`
}

// CollectionPageProperty describes properties of a collection
type CollectionPageProperty struct {
	Property string `json:"property"`
	Visible  bool   `json:"visible"`
}

// CollectionColumnInfo describes a info of a collection column
type CollectionColumnInfo struct {
	Name string `json:"name"`
	// ColumnTypeTitle etc.
	Type string `json:"type"`

	// for Type == ColumnTypeNumber, e.g. "dollar", "number"
	NumberFormat string `json:"number_format"`

	// For Type == ColumnTypeRollup
	TargetProperty     string `json:"target_property"`
	RelationProperty   string `json:"relation_property"`
	TargetPropertyType string `json:"target_property_type"`

	// for Type == ColumnTypeRelation
	CollectionID string `json:"collection_id"`
	Property     string `json:"property"`

	Options []*CollectionColumnOption `json:"options"`

	// TODO: would have to set it up from Collection.RawJSON
	//RawJSON map[string]interface{} `json:"-"`
}

// CollectionColumnOption describes options for ColumnTypeMultiSelect
// collection column
type CollectionColumnOption struct {
	Color string `json:"color"`
	ID    string `json:"id"`
	Value string `json:"value"`
}

// LoadPageChunk executes a raw API call /api/v3/loadPageChunk
func (c *Client) LoadPageChunk(pageID string, chunkNo int, cur *cursor) (*LoadPageChunkResponse, error) { // emulating notion's website api usage: 50 items on first request,
	// 30 on subsequent requests
	limit := 30
	apiURL := "/api/v3/loadPageChunk"
	if cur == nil {
		cur = &cursor{
			// to mimic browser api which sends empty array for this argment
			Stack: make([][]stack, 0),
		}
		limit = 50
	}
	req := &loadPageChunkRequest{
		PageID:          pageID,
		ChunkNumber:     chunkNo,
		Limit:           limit,
		Cursor:          *cur,
		VerticalColumns: false,
	}
	var rsp LoadPageChunkResponse
	var err error
	if rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp); err != nil {
		return nil, err
	}
	if err = parseRecordMap(rsp.RecordMap); err != nil {
		return nil, err
	}
	return &rsp, nil
}

func parseRecordMap(recordMap *RecordMap) error {
	for _, r := range recordMap.Blocks {
		if err := parseRecord(TableBlock, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.Spaces {
		if err := parseRecord(TableSpace, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.Users {
		if err := parseRecord(TableUser, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.CollectionViews {
		if err := parseRecord(TableCollectionView, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.Collections {
		if err := parseRecord(TableCollection, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.Discussions {
		if err := parseRecord(TableDiscussion, r); err != nil {
			return err
		}
	}

	for _, r := range recordMap.Comments {
		if err := parseRecord(TableComment, r); err != nil {
			return err
		}
	}

	return nil
}
