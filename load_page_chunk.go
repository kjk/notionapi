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
	// TDOO: parses those types as well
	Comments    map[string]*Record `josn:"comment"`
	Discussions map[string]*Record `jsoon:"discussion"`
}

// Space describes Notion workspace.
type Space struct {
	ID                  string                  `json:"id"`
	Version             float64                 `json:"version"`
	Name                string                  `json:"name"`
	Domain              string                  `json:"domain"`
	Permissions         []*SpacePermissions     `json:"permissions,omitempty"`
	PermissionGroups    []SpacePermissionGroups `json:"permission_groups"`
	Icon                string                  `json:"icon"`
	EmailDomains        []string                `json:"email_domains"`
	BetaEnabled         bool                    `json:"beta_enabled"`
	Pages               []string                `json:"pages,omitempty"`
	DisablePublicAccess bool                    `json:"disable_public_access"`
	DisableGuests       bool                    `json:"disable_guests"`
	DisableMoveToSpace  bool                    `json:"disable_move_to_space"`
	DisableExport       bool                    `json:"disable_export"`

	CreatedBy      string `json:"created_by"`
	CreatedTime    int64  `json:"created_time"`
	LastEditedBy   string `json:"last_edited_by"`
	LastEditedTime int64  `json:"last_edited_time"`

	RawJSON map[string]interface{} `json:"-"`
}

type SpacePermissions struct {
	Role   string `json:"role"`
	Type   string `json:"type"` // e.g. "user_permission"
	UserID string `json:"user_id"`
}

type SpacePermissionGroups struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	UserIds []string `json:"user_ids,omitempty"`
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

// Collection describes a collection
type Collection struct {
	ID          string                           `json:"id"`
	Version     int                              `json:"version"`
	Name        interface{}                      `json:"name"`
	Schema      map[string]*CollectionColumnInfo `json:"schema"`
	Format      *CollectionFormat                `json:"format"`
	ParentID    string                           `json:"parent_id"`
	ParentTable string                           `json:"parent_table"`
	Alive       bool                             `json:"alive"`
	CopiedFrom  string                           `json:"copied_from"`

	// TODO: are those ever present?
	Type          string   `json:"type"`
	FileIDs       []string `json:"file_ids"`
	Icon          string   `json:"icon"`
	TemplatePages []string `json:"template_pages"`

	// calculated by us
	name    []*TextSpan
	RawJSON map[string]interface{} `json:"-"`
}

// GetName parses Name and returns as a string
func (c *Collection) GetName() string {
	if len(c.name) == 0 {
		if c.Name == nil {
			return ""
		}
		c.name, _ = ParseTextSpans(c.Name)
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
