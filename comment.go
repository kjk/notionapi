package notionapi

// Comment describes a single comment in a discussion
type Comment struct {
	ID             string      `json:"id"`
	Version        int64       `json:"version"`
	Alive          bool        `json:"alive"`
	ParentID       string      `json:"parent_id"`
	ParentTable    string      `json:"parent_table"`
	CreatedBy      string      `json:"created_by"`
	CreatedTime    int64       `json:"created_time"`
	Text           interface{} `json:"text"`
	LastEditedTime int64       `json:"last_edited_time"`

	// set by us
	RawJSON map[string]interface{} `json:"-"`
}
