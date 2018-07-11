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
	Alive bool `json:"alive"`
	// List of block ids for that make up content of this block
	// Use Content to get corresponding block (they are in the same order)
	ContentIDs []string `json:"content"`
	// ID of the user who created this block
	CreatedBy   string `json:"created_by"`
	CreatedTime int64  `json:"created_time"`
	// only available when different than default?
	// TODO: this is different for different types
	Format *format `json:"format,omitempty"`
	// a unique ID of the block
	ID string `json:"id"`
	// ID of the user who last edited this block
	LastEditedBy   string `json:"last_edited_by"`
	LastEditedTime int64  `json:"last_edited_time"`
	// ID of parent Block
	ParentID    string `json:"parent_id"`
	ParentTable string `json:"parent_table"`
	// not always available
	Permissions *[]permission          `json:"permissions,omitempty"`
	Properties  map[string]interface{} `json:"properties"`
	// type of the block e.g. TypeText, TypePage etc.
	Type string `json:"type"`
	// blocks are versioned
	Version int64 `json:"version"`

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

	// for TypeGist it's the url for the gist
	// for TypeBookmark it's the url of the page
	// fot TypeImage it's url of the image, but use ImageURL instead
	// because Source is sometimes not accessible
	Source string `json:"source"`

	// for TypeImage it's
	ImageURL string `json:"image_url"`

	// for TypeCode
	Code         string `json:"code"`
	CodeLanguage string `json:"code_language"`
}

// IsLinkToPage returns true if block element is a link to existing page
// (as opposed to )
func (b *Block) IsLinkToPage() bool {
	if b.Type != TypePage {
		return false
	}
	return b.ParentTable == TableSpace
}

// IsPage returns true if block represents a page (either a sub-page or
// a link to a page)
func (b *Block) IsPage() bool {
	return b.Type == TypePage
}

// IsImage returns true if block represents an image
func (b *Block) IsImage() bool {
	return b.Type == TypeImage
}

// IsCode returns true if block represents a code block
func (b *Block) IsCode() bool {
	return b.Type == TypeCode
}

/* TODO: not sure if I need this
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
*/

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
