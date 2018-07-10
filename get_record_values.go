package notion

import "encoding/json"

// /api/v3/getRecordValues request
type getRecordValuesRequest struct {
	Requests []getRecordValuesRequestInner `json:"requests"`
}

type getRecordValuesRequestInner struct {
	Table string `json:"table"`
	ID    string `json:"id"`
}

// /api/v3/getRecordValues response
// Note: it depends on Table type in request
type getRecordValuesResponse struct {
	Results []*BlockWithRole `json:"results"`
}

// BlockWithRole describes a block info
type BlockWithRole struct {
	Role  string `json:"role"`
	Value *Block `json:"value"`
}

// Block describes a block
type Block struct {
	// values that come from JSON
	Alive       bool     `json:"alive"`
	ContentIDs  []string `json:"content"`
	CreatedBy   string   `json:"created_by"`
	CreatedTime int64    `json:"created_time"`
	// only available when different than default?
	// TODO: this is different for different types
	Format         *format `json:"format,omitempty"`
	ID             string  `json:"id"`
	LastEditedBy   string  `json:"last_edited_by"`
	LastEditedTime int64   `json:"last_edited_time"`
	ParentID       string  `json:"parent_id"`
	ParentTable    string  `json:"parent_table"`
	// not always available
	Permissions *[]permission          `json:"permissions,omitempty"`
	Properties  map[string]interface{} `json:"properties"`
	Type        string                 `json:"type"`
	Version     int64                  `json:"version"`

	// Values calculated by us
	// maps ContentIDs array
	Content []*Block `json:"content_resolved"`
	// this is for some types like TypePage, TypeText, TypeHeader etc.
	InlineContent []*InlineBlock `json:"inline_content"`

	// for TypePage
	Title string `json:"title"`
	// For TypeTodo, a checked state
	IsChecked bool `json:"is_checked"`

	// for TypeBookmark
	Description string `json:"description"`
	Link        string `json:"link"`

	// for TypeGist
	Source string `json:"source"`

	// we resolve blocks, possilby multiple times, so we mark them
	// as resolved to avoid duplicate work
	isResolved bool
}

// IsTypeWithBlocks returns true if BlockValue contains Blocks value
// extracted from Properties["title"]
func IsTypeWithBlocks(blockType string) bool {
	switch blockType {
	case TypePage, TypeText, TypeBulletedList, TypeNumberedList,
		TypeToggle, TypeTodo, TypeHeader, TypeSubHeader, TypeQuote:
		return true
	}
	return false
}

type format struct {
	PageFullWidth bool `json:"page_full_width"`
	PageSmallText bool `json:"page_small_text"`
}

type permission struct {
	Role   string  `json:"role"`
	Type   string  `json:"type"`
	UserID *string `json:"user_id,omitempty"`
}

func parseGetRecordValues(d []byte) (*getRecordValuesResponse, error) {
	var rec getRecordValuesResponse
	err := json.Unmarshal(d, &rec)
	if err != nil {
		return nil, err
	}
	return &rec, nil
}
