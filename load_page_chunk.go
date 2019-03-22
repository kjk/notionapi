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
	RecordMap RecordMap `json:"recordMap"`
	Cursor    cursor    `json:"cursor"`
	RawJSON   []byte    `json:"-"`
}

// RecordMap contains a collections of blocks, a space, users, and collections.
type RecordMap struct {
	Blocks          map[string]*BlockWithRole          `json:"block"`
	Space           map[string]*SpaceWithRole          `json:"space"`
	Users           map[string]*UserWithRole           `json:"notion_user"`
	Collections     map[string]*CollectionWithRole     `json:"collection"`
	CollectionViews map[string]*CollectionViewWithRole `json:"collection_view"`
	// TDOO: there might be more records types
}

// SpaceWithRole holds a user's role associated with a space and a space.
type SpaceWithRole struct {
	Role  string `json:"role,omitempty"`
	Value *Space `json:"value,omitempty"`
}

// Space is a notion.so workspace.
type Space struct {
	ID          string        `json:"id"`
	Version     float64       `json:"version"`
	Name        string        `json:"name"`
	BetaEnabled bool          `json:"beta_enabled"`
	Permissions *[]Permission `json:"permissions,omitempty"`
	Pages       []string      `json:"pages,omitempty"`
}

// CollectionViewWithRole describes a role and a collection view
type CollectionViewWithRole struct {
	Role  string          `json:"role"`
	Value *CollectionView `json:"value"`
}

// CollectionView describes a collection
type CollectionView struct {
	ID          string                `json:"id"`
	Alive       bool                  `json:"alive"`
	Format      *CollectionViewFormat `json:"format"`
	Name        string                `json:"name"`
	PageSort    []string              `json:"page_sort"`
	ParentID    string                `json:"parent_id"`
	ParentTable string                `json:"parent_table"`
	Query       *CollectionViewQuery  `json:"query"`
	Type        string                `json:"type"`
	Version     int                   `json:"version"`
}

// CollectionViewFormat describes a fomrat of a collection view
type CollectionViewFormat struct {
	TableProperties []*TableProperty `json:"table_properties"`
	TableWrap       bool             `json:"table_wrap"`
}

// CollectionViewQuery describes a query
type CollectionViewQuery struct {
	Aggregate []*AggregateQuery `json:"aggregate"`
}

// AggregateQuery describes an aggregate query
type AggregateQuery struct {
	AggregationType string `json:"aggregation_type"`
	ID              string `json:"id"`
	Property        string `json:"property"`
	Type            string `json:"type"`
	ViewType        string `json:"view_type"`
}

// CollectionWithRole describes a collection
type CollectionWithRole struct {
	Role  string      `json:"role"`
	Value *Collection `json:"value"`
}

// Collection describes a collection
type Collection struct {
	Alive            bool                             `json:"alive"`
	Format           *CollectionFormat                `json:"format"`
	ID               string                           `json:"id"`
	Name             [][]string                       `json:"name"`
	ParentID         string                           `json:"parent_id"`
	ParentTable      string                           `json:"parent_table"`
	CollectionSchema map[string]*CollectionColumnInfo `json:"schema"`
	Version          int                              `json:"version"`
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
	Type    string                    `json:"type"`
}

// CollectionColumnOption describes options for a collection column
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
	Email                     string `json:"email"`
	FamilyName                string `json:"family_name"`
	GivenName                 string `json:"given_name"`
	ID                        string `json:"id"`
	Locale                    string `json:"locale"`
	MobileOnboardingCompleted bool   `json:"mobile_onboarding_completed"`
	OnboardingCompleted       bool   `json:"onboarding_completed"`
	ProfilePhoto              string `json:"profile_photo"`
	TimeZone                  string `json:"time_zone"`
	Version                   int    `json:"version"`
}

// Date describes a date
type Date struct {
	// "MMM DD, YYYY", "MM/DD/YYYY", "DD/MM/YYYY", "YYYY/MM/DD", "relative"
	DateFormat string    `json:"date_format"`
	Reminder   *Reminder `json:"reminder,omitempty"`
	// "2018-07-12"
	StartDate string `json:"start_date"`
	// "09:00"
	StartTime *string `json:"start_time,omitempty"`
	// "America/Los_Angeles"
	TimeZone *string `json:"time_zone,omitempty"`
	// "H:mm" for 24hr, not given for 12hr
	TimeFormat *string `json:"time_format,omitempty"`
	// "date", "datetime"
	Type string `json:"type"`
}

// Reminder describes date reminder
type Reminder struct {
	Time  string `json:"time"` // e.g. "09:00"
	Unit  string `json:"unit"` // e.g. "day"
	Value int64  `json:"value"`
}

// LoadPageChunk executes a raw API call /api/v3/loadPageChunk
func (c *Client) LoadPageChunk(pageID string, chunkNo int, cur *cursor) (*LoadPageChunkResponse, error) {
	// emulating notion's website api usage: 50 items on first request,
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
	return &rsp, nil
}
