package notionapi

// Author represents the author of an Edit
type Author struct {
	ID    string `json:"id"`
	Table string `json:"table"`
}

// Edit represents a Notion edit (ie. a change made during an Activity)
type Edit struct {
	SpaceID   string   `json:"space_id"`
	Authors   []Author `json:"authors"`
	Timestamp int64    `json:"timestamp"`
	Type      string   `json:"type"`
	Version   int      `json:"version"`

	CommentData  Comment `json:"comment_data"`
	CommentID    string  `json:"comment_id"`
	DiscussionID string  `json:"discussion_id"`

	BlockID   string `json:"block_id"`
	BlockData struct {
		BlockValue Block `json:"block_value"`
	} `json:"block_data"`
	NavigableBlockID string `json:"navigable_block_id"`

	CollectionID    string `json:"collection_id"`
	CollectionRowID string `json:"collection_row_id"`
}

// Activity represents a Notion activity (ie. event)
type Activity struct {
	Role string `json:"role"`

	ID        string `json:"id"`
	SpaceID   string `json:"space_id"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Type      string `json:"type"`
	Version   int    `json:"version"`

	ParentID    string `json:"parent_id"`
	ParentTable string `json:"parent_table"`

	// If the edit was to a block inside a regular page
	NavigableBlockID string `json:"navigable_block_id"`

	// If the edit was to a block inside a collection or collection row
	CollectionID    string `json:"collection_id"`
	CollectionRowID string `json:"collection_row_id"`

	Edits []Edit `json:"edits"`

	Index   int  `json:"index"`
	Invalid bool `json:"invalid"`

	RawJSON map[string]interface{} `json:"-"`
}
