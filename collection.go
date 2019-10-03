package notionapi

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
