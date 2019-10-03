package notionapi

const (
	// TODO: those are probably CollectionViewType
	// CollectionViewTypeTable is a table block
	CollectionViewTypeTable = "table"
	// CollectionViewTypeTable is a lists block
	CollectionViewTypeList = "list"
)

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

// TableProperty describes property of a table
type TableProperty struct {
	Width    int    `json:"width"`
	Visible  bool   `json:"visible"`
	Property string `json:"property"`
}

// FormatTable describes format for BlockTable
type FormatTable struct {
	PageSort        []string         `json:"page_sort"`
	TableWrap       bool             `json:"table_wrap"`
	TableProperties []*TableProperty `json:"table_properties"`
}

// CollectionView represents a collection view
type CollectionView struct {
	ID          string       `json:"id"`
	Version     int64        `json:"version"`
	Type        string       `json:"type"` // "table"
	Format      *FormatTable `json:"format"`
	Name        string       `json:"name"`
	ParentID    string       `json:"parent_id"`
	ParentTable string       `json:"parent_table"`
	Query       *Query       `json:"query"`
	Alive       bool         `json:"alive"`
	PageSort    []string     `json:"page_sort"`

	// set by us
	RawJSON map[string]interface{} `json:"-"`
}

// CollectionViewInfo describes a particular view of the collection
// TODO: same as table?
type CollectionViewInfo struct {
	OriginatingBlock *Block
	CollectionView   *CollectionView
	Collection       *Collection
	CollectionRows   []*Block
}
