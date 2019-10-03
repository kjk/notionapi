package notionapi

// Discussion represents a discussion
type Discussion struct {
	ID          string   `json:"id"`
	Version     int64    `json:"version"`
	ParentID    string   `json:"parent_id"`
	ParentTable string   `json:"parent_table"`
	Resolved    bool     `json:"resolved"`
	Comments    []string `json:"comments"`
	// set by us
	RawJSON map[string]interface{} `json:"-"`
}
