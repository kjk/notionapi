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
	Blocks          map[string]*BlockWithRole      `json:"block"`
	Space           map[string]*SpaceWithRole      `json:"space"`
	Users           map[string]*UserWithRole       `json:"notion_user"`
	Collections     map[string]*CollectionWithRole `json:"collection"`
	CollectionViews map[string]*BlockWithRole      `json:"collection_view"`
	// TDOO: there might be more records types
}

// SpaceWithRole holds a user's role associated with a space and a space.
type SpaceWithRole struct {
	Role  string `json:"role,omitempty"`
	Value *Space `json:"value,omitempty"`
}

// Space is a notion.so workspace.
type Space struct {
	ID               string                  `json:"id"`
	Version          float64                 `json:"version"`
	Name             string                  `json:"name"`
	Domain           string                  `json:"domain"`
	Permissions      []*SpacePermissions     `json:"permissions,omitempty"`
	PermissionGroups []SpacePermissionGroups `json:"permission_groups"`
	EmailDomains     []string                `json:"email_domains"`
	BetaEnabled      bool                    `json:"beta_enabled"`
	Pages            []string                `json:"pages,omitempty"`
	CreatedBy        string                  `json:"created_by"`
	CreatedTime      int64                   `json:"created_time"`
	LastEditedBy     string                  `json:"last_edited_by"`
	LastEditedTime   int64                   `json:"last_edited_time"`

	RawJSON map[string]interface{} `json:"-"`
}

type SpacePermissions struct {
	Role   string `json:"role"`
	Type   string `json:"type"`
	UserID string `json:"user_id"`
}

type SpacePermissionGroups struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	UserIds []string `json:"user_ids,omitempty"`
}

// TODO: ListFormat or TableFormat
// type CollectionViewFormat struct {

// Query describes a query
// TODO: merge with CollectionQuery
type Query struct {
	Aggregate []*AggregateQuery `json:"aggregate"`

	FilterOperator string         `json:"filter_operator"`
	Filter         []*FilterQuery `json:"filter"`
	Sort           []*SortQuery   `json:"sort"`
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
	Direction string `json:"direction"`
	ID        string `json:"id"`
	Property  string `json:"property"`
	Type      string `json:"type"`
}

// CollectionWithRole describes a collection
type CollectionWithRole struct {
	Role  string      `json:"role"`
	Value *Collection `json:"value"`
}

// Collection describes a collection
// TODO: why doesn't it have "type" like a Block?
type Collection struct {
	// form json
	Alive            bool                             `json:"alive"`
	CopiedFrom       string                           `json:"copied_from"`
	FileIDs          []string                         `json:"file_ids"`
	Format           *CollectionFormat                `json:"format"`
	Icon             string                           `json:"icon"`
	ID               string                           `json:"id"`
	NameVal          interface{}                      `json:"name"` // TODO: because created Name() first
	ParentID         string                           `json:"parent_id"`
	ParentTable      string                           `json:"parent_table"`
	CollectionSchema map[string]*CollectionColumnInfo `json:"schema"`
	TemplatePages    []string                         `json:"template_pages"`
	Version          int                              `json:"version"`

	// calculated by us
	name    []*TextSpan
	RawJSON map[string]interface{} `json:"-"`
}

func (c *Collection) Name() string {
	if len(c.name) == 0 {
		if c.NameVal == nil {
			return ""
		}
		c.name, _ = ParseTextSpans(c.NameVal)
	}
	return TextSpansToString(c.name)
}

// CollectionFormat describes format of a collection
type CollectionFormat struct {
	CollectionPageProperties []*CollectionPageProperty `json:"collection_page_properties"`
}

// CollectionPageProperty describes properties of a collection
type CollectionPageProperty struct {
	Property string `json:"property"`
	Visible  bool   `json:"visible"`
}

// CollectionColumnInfo describes a info of a collection column
type CollectionColumnInfo struct {
	Name    string                    `json:"name"`
	Options []*CollectionColumnOption `json:"options"`
	// ColumnTypeTitle etc.
	Type string `json:"type"`

	// For rollup column types
	TargetProperty     string `json:"target_property"`
	RelationProperty   string `json:"relation_property"`
	TargetPropertyType string `json:"target_property_type"`

	RawJSON map[string]interface{} `json:"-"`
}

// CollectionColumnOption describes options for ColumnTypeMultiSelect
// collection column
type CollectionColumnOption struct {
	Color string `json:"color"`
	ID    string `json:"id"`
	Value string `json:"value"`
}

// UserWithRole describes a user and its role
type UserWithRole struct {
	Role  string `json:"role"`
	Value *User  `json:"value"`
}

// User describes a user
type User struct {
	Email                      string `json:"email"`
	FamilyName                 string `json:"family_name"`
	GivenName                  string `json:"given_name"`
	ID                         string `json:"id"`
	Locale                     string `json:"locale"`
	MobileOnboardingCompleted  bool   `json:"mobile_onboarding_completed"`
	OnboardingCompleted        bool   `json:"onboarding_completed"`
	ClipperOnboardingCompleted bool   `json:"clipper_onboarding_completed"`
	ProfilePhoto               string `json:"profile_photo"`
	TimeZone                   string `json:"time_zone"`
	Version                    int    `json:"version"`

	RawJSON map[string]interface{} `json:"-"`
}

// Date describes a date
type Date struct {
	// "MMM DD, YYYY", "MM/DD/YYYY", "DD/MM/YYYY", "YYYY/MM/DD", "relative"
	DateFormat string    `json:"date_format"`
	Reminder   *Reminder `json:"reminder,omitempty"`
	// "2018-07-12"
	StartDate string `json:"start_date"`
	// "09:00"
	StartTime string `json:"start_time,omitempty"`
	// "2018-07-12"
	EndDate string `json:"end_date,omitempty"`
	// "09:00"
	EndTime string `json:"end_time,omitempty"`
	// "America/Los_Angeles"
	TimeZone *string `json:"time_zone,omitempty"`
	// "H:mm" for 24hr, not given for 12hr
	TimeFormat string `json:"time_format,omitempty"`
	// "date", "datetime", "datetimerange", "daterange"
	Type string `json:"type"`
}

// Reminder describes date reminder
type Reminder struct {
	Time  string `json:"time"` // e.g. "09:00"
	Unit  string `json:"unit"` // e.g. "day"
	Value int64  `json:"value"`
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
	rsp.RawJSON, err = doNotionAPI(c, apiURL, req, &rsp)
	if err != nil {
		return nil, err
	}
	setLoadPageChunkResponse(&rsp, rsp.RawJSON)
	return &rsp, nil
}

func setLoadPageChunkResponse(r *LoadPageChunkResponse, json map[string]interface{}) {
	recordMapJSON := jsonGetMap(json, "recordMap")
	{
		blockByID := jsonGetMap(recordMapJSON, "block")
		for id, br := range r.RecordMap.Blocks {
			brJSON := jsonGetMap(blockByID, id)
			b := br.Value
			bJSON := jsonGetMap(brJSON, "value")
			b.RawJSON = bJSON
		}
	}

	{
		spaceByID := jsonGetMap(recordMapJSON, "space")
		for id, sr := range r.RecordMap.Space {
			srJSON := jsonGetMap(spaceByID, id)
			s := sr.Value
			sJSON := jsonGetMap(srJSON, "value")
			s.RawJSON = sJSON
		}
	}
	{
		userByID := jsonGetMap(recordMapJSON, "notion_user")
		for id, ur := range r.RecordMap.Users {
			urJSON := jsonGetMap(userByID, id)
			u := ur.Value
			uJSON := jsonGetMap(urJSON, "value")
			u.RawJSON = uJSON
		}
	}
	{
		collectionByID := jsonGetMap(recordMapJSON, "collection")
		for id, cr := range r.RecordMap.Collections {
			crJSON := jsonGetMap(collectionByID, id)
			c := cr.Value
			cJSON := jsonGetMap(crJSON, "value")
			c.RawJSON = cJSON
		}
	}
	{
		collectionViewByID := jsonGetMap(recordMapJSON, "collection_view")
		for id, cvr := range r.RecordMap.CollectionViews {
			cvrJSON := jsonGetMap(collectionViewByID, id)
			cv := cvr.Value
			cvJSON := jsonGetMap(cvrJSON, "value")
			cv.RawJSON = cvJSON
		}
	}
}
