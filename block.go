package notionapi

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	// BlockPage is a notion Page
	BlockPage = "page"
	// BlockText is a text block
	BlockText = "text"
	// BlockBookmark is a bookmark block
	BlockBookmark = "bookmark"
	// BlockBulletedList is a bulleted list block
	BlockBulletedList = "bulleted_list"
	// BlockNumberedList is a numbered list block
	BlockNumberedList = "numbered_list"
	// BlockToggle is a toggle block
	BlockToggle = "toggle"
	// BlockTodo is a todo block
	BlockTodo = "to_do"
	// BlockDivider is a divider block
	BlockDivider = "divider"
	// BlockImage is an image block
	BlockImage = "image"
	// BlockHeader is a header block
	BlockHeader = "header"
	// BlockSubHeader is a header block
	BlockSubHeader = "sub_header"
	// BlockSubSubHeader
	BlockSubSubHeader = "sub_sub_header"
	// BlockQuote is a quote block
	BlockQuote = "quote"
	// BlockComment is a comment block
	BlockComment = "comment"
	// BlockCode is a code block
	BlockCode = "code"
	// BlockColumnList is for multi-column. Number of columns is
	// number of content blocks of type TypeColumn
	BlockColumnList = "column_list"
	// BlockColumn is a child of TypeColumnList
	BlockColumn = "column"
	// BlockTable is a table block
	BlockTable = "table"
	// BlockCollectionView is a collection view block
	BlockCollectionView = "collection_view"
	// BlockCollectionViewPage is a page that is a collection
	BlockCollectionViewPage = "collection_view_page"
	// BlockVideo is youtube video embed
	BlockVideo = "video"
	// BlockFile is an embedded file
	BlockFile = "file"
	// BlockPDF is an embedded pdf file
	BlockPDF = "pdf"
	// BlockGist is embedded gist block
	BlockGist = "gist"
	// BlockDrive is embedded Google Drive file
	BlockDrive = "drive"
	// BlockTweet is embedded gist block
	BlockTweet = "tweet"
	// BlockMaps is embedded Google Map block
	BlockMaps = "maps"
	// BlockCodepen is embedded codepen block
	BlockCodepen = "codepen"
	// BlockEmbed is a generic oembed link
	BlockEmbed = "embed"
	// BlockCallout is a callout
	BlockCallout = "callout"
	// BlockTableOfContents is table of contents
	BlockTableOfContents = "table_of_contents"
	// BlockBreadcrumb is breadcrumb block
	BlockBreadcrumb = "breadcrumb"
	// BlockEquation is TeX equation block
	BlockEquation = "equation"
	// BlockFactory represents a factory block
	BlockFactory = "factory"
)

// FormatToggle describes format for BlockToggle
type FormatToggle struct {
	BlockColor string `json:"block_color"`
}

// FormatNumberedList describes format for BlockNumberedList
type FormatNumberedList struct {
	BlockColor string `json:"block_color"`
}

// FormatBulletedList describes format for BlockBulletedList
type FormatBulletedList struct {
	BlockColor string `json:"block_color"`
}

// FormatPage describes format for BlockPage
type FormatPage struct {
	// /images/page-cover/gradients_11.jpg
	PageCover string `json:"page_cover"`
	// e.g. 0.6
	PageCoverPosition float64 `json:"page_cover_position"`
	PageFont          string  `json:"page_font"`
	PageFullWidth     bool    `json:"page_full_width"`
	// it's url like https://s3-us-west-2.amazonaws.com/secure.notion-static.com/8b3930e3-9dfe-4ba7-a845-a8ff69154f2a/favicon-256.png
	// or emoji like "✉️"
	PageIcon      string `json:"page_icon"`
	PageSmallText bool   `json:"page_small_text"`
	BlockColor    string `json:"block_color"`

	// calculated by us
	PageCoverURL string `json:"page_cover_url,omitempty"`
}

// FormatBookmark describes format for BlockBookmark
type FormatBookmark struct {
	Icon  string `json:"bookmark_icon"`
	Cover string `json:"bookmark_cover"`
}

// FormatImage describes format for BlockImage
type FormatImage struct {
	// comes from notion API
	BlockAspectRatio   float64 `json:"block_aspect_ratio"`
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
	BlockWidth         float64 `json:"block_width"`
	DisplaySource      string  `json:"display_source,omitempty"`

	// calculated by us
	ImageURL string `json:"image_url,omitempty"`
}

// FormatVideo describes fromat form BlockVideo
type FormatVideo struct {
	BlockWidth         int64   `json:"block_width"`
	BlockHeight        int64   `json:"block_height"`
	DisplaySource      string  `json:"display_source"`
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockAspectRatio   float64 `json:"block_aspect_ratio"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
}

// FormatText describes format for BlockText
// TODO: possibly more?
type FormatText struct {
	BlockColor string `json:"block_color,omitempty"`
}

// FormatHeader describes format for BlockHeader, BlockSubHeader, BlockSubSubHeader
// TODO: possibly more?
type FormatHeader struct {
	BlockColor string `json:"block_color,omitempty"`
}

// FormatTable describes format for BlockTable
type FormatTable struct {
	TableWrap       bool             `json:"table_wrap"`
	TableProperties []*TableProperty `json:"table_properties"`
}

// FormatColumn describes format for BlockColumn
type FormatColumn struct {
	ColumnRatio float64 `json:"column_ratio"` // e.g. 0.5 for half-sized column
}

// FormatEmbed describes format for BlockEmbed
type FormatEmbed struct {
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockHeight        float64 `json:"block_height"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
	DisplaySource      string  `json:"display_source"`
}

// TableProperty describes property of a table
type TableProperty struct {
	Width    int    `json:"width"`
	Visible  bool   `json:"visible"`
	Property string `json:"property"`
}

// Permission describes user permissions
type Permission struct {
	Role   string  `json:"role"`
	Type   string  `json:"type"`
	UserID *string `json:"user_id,omitempty"`
}

// Block describes a block
type Block struct {
	// values that come from JSON
	// a unique ID of the block
	ID string `json:"id"`
	// if false, the page is deleted
	Alive bool `json:"alive"`
	// List of block ids for that make up content of this block
	// Use Content to get corresponding block (they are in the same order)
	ContentIDs   []string `json:"content,omitempty"`
	CopiedFrom   string   `json:"copied_from,omitempty"`
	CollectionID string   `json:"collection_id,omitempty"` // for BlockCollectionView
	// ID of the user who created this block
	CreatedBy   string `json:"created_by"`
	CreatedTime int64  `json:"created_time"`
	// List of block ids with discussion content
	DiscussionIDs []string `json:"discussion,omitempty"`
	// those ids seem to map to storage in s3
	// https://s3-us-west-2.amazonaws.com/secure.notion-static.com/${id}/${name}
	FileIDs []string `json:"file_ids,omitempty"`

	// TODO: don't know what this means
	IgnoreBlockCount bool `json:"ignore_block_count,omitempty"`

	// ID of the user who last edited this block
	LastEditedBy   string `json:"last_edited_by"`
	LastEditedTime int64  `json:"last_edited_time"`
	// ID of parent Block
	ParentID    string `json:"parent_id"`
	ParentTable string `json:"parent_table"`
	// not always available
	Permissions *[]Permission          `json:"permissions,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	// type of the block e.g. TypeText, TypePage etc.
	Type string `json:"type"`
	// blocks are versioned
	Version int64    `json:"version"`
	ViewIDs []string `json:"view_ids,omitempty"`

	// Parent of this block
	Parent *Block `json:"-"`

	// maps ContentIDs array to Block type
	Content []*Block `json:"-"`
	// this is for some types like TypePage, TypeText, TypeHeader etc.
	InlineContent []*TextSpan `json:"-"`

	// for BlockPage
	Title string `json:"-"`

	// For BlockTodo, a checked state
	IsChecked bool `json:"-"`

	// for BlockBookmark
	Description string `json:"-"`
	Link        string `json:"-"`

	// for BlockBookmark it's the url of the page
	// for BlockGist it's the url for the gist
	// fot BlockImage it's url of the image, but use ImageURL instead
	// because Source is sometimes not accessible
	// for BlockFile it's url of the file
	// for BlockEmbed it's url of the embed
	Source string `json:"-"`

	// for BlockFile
	FileSize string `json:"-"`

	// for BlockImage it's an URL built from Source that is always accessible
	ImageURL string `json:"-"`

	// for BlockCode
	Code         string `json:"-"`
	CodeLanguage string `json:"-"`

	// for BlockCollectionView
	// It looks like the info about which view is selected is stored in browser
	CollectionViews []*CollectionViewInfo `json:"-"`

	Page *Page `json:"-"`

	// RawJSON represents Block as
	RawJSON map[string]interface{} `json:"-"`

	isResolved bool
}

// CollectionViewInfo describes a particular view of the collection
// TODO: same as table?
type CollectionViewInfo struct {
	OriginatingBlock *Block
	CollectionView   *CollectionView
	Collection       *Collection
	CollectionRows   []*Block
	// for serialization of state to JSON
	queryCollectionResponse *QueryCollectionResponse
}

func (b *Block) FormatStringValue(key string) (string, bool) {
	format := jsonGetMap(b.RawJSON, "format")
	if format == nil {
		return "", false
	}
	return jsonGetString(format, key)
}

// CreatedOn return the time the page was created
func (b *Block) CreatedOn() time.Time {
	return time.Unix(b.CreatedTime/1000, 0)
}

// UpdatedOn returns the time the page was last updated
func (b *Block) UpdatedOn() time.Time {
	return time.Unix(b.LastEditedTime/1000, 0)
}

// IsLinkToPage returns true if block element is a link to a page
// (as opposed to embedded page)
func (b *Block) IsLinkToPage() bool {
	if b.Type != BlockPage {
		return false
	}
	return b.ParentTable == TableSpace
}

// IsSubPage returns true if this is a sub-page (as opposed to
// link to a page that is not a child of that page)
func (b *Block) IsSubPage() bool {
	panicIf(b.Type != BlockPage)
	if b.Parent == nil {
		return false
	}
	return b.ParentID == b.Parent.ID
}

// IsPage returns true if block represents a page (either a
// sub-page or a link to a page)
func (b *Block) IsPage() bool {
	return b.Type == BlockPage
}

// IsImage returns true if block represents an image
func (b *Block) IsImage() bool {
	return b.Type == BlockImage
}

// IsCode returns true if block represents a code block
func (b *Block) IsCode() bool {
	return b.Type == BlockCode
}

func getProp(block *Block, name string, toSet *string) bool {
	v, ok := block.Properties[name]
	if !ok {
		return false
	}
	s, err := getFirstInlineBlock(v)
	if err != nil {
		return false
	}
	*toSet = s
	return true
}

func (b *Block) GetProperty(name string) []*TextSpan {
	v, ok := b.Properties[name]
	if !ok {
		return nil
	}
	ts, err := ParseTextSpans(v)
	if err != nil {
		return nil
	}
	return ts
}

func (b *Block) GetCaption() []*TextSpan {
	return b.GetProperty("caption")
}

func (b *Block) GetTitle() []*TextSpan {
	return b.GetProperty("title")
}

func parseProperties(block *Block) error {
	var err error
	props := block.Properties
	if title, ok := props["title"]; ok {
		block.InlineContent, err = ParseTextSpans(title)
		if err != nil {
			return err
		}
		switch block.Type {
		case BlockPage, BlockFile, BlockBookmark:
			block.Title, err = getInlineText(title)
		case BlockCode:
			block.Code, err = getFirstInlineBlock(title)
		default:
		}
		if err != nil {
			return err
		}
	}

	if BlockTodo == block.Type {
		if checked, ok := props["checked"]; ok {
			s, _ := getFirstInlineBlock(checked)
			// fmt.Printf("checked: '%s'\n", s)
			block.IsChecked = strings.EqualFold(s, "Yes")
		}
	}

	// for BlockBookmark
	getProp(block, "description", &block.Description)
	// for BlockBookmark
	getProp(block, "link", &block.Link)

	// for BlockBookmark, BlockImage, BlockGist, BlockFile, BlockEmbed
	// don't over-write if was already set from "source" json field
	if block.Source == "" {
		getProp(block, "source", &block.Source)
	}

	if block.Source != "" && block.IsImage() {
		block.ImageURL = maybeProxyImageURL(block.Source)
	}

	// for BlockCode
	getProp(block, "language", &block.CodeLanguage)

	// for BlockFile
	if block.Type == BlockFile {
		getProp(block, "size", &block.FileSize)
	}

	return nil
}

func (b *Block) panicIfNotOfType(expectedType string) {
	if b.Type != expectedType {
		panic(fmt.Sprintf("operation on invalid block. Block type: %s, expected type: %s", b.Type, expectedType))
	}
}

func (b *Block) unmarshalFormat(expectedType string, v interface{}) bool {
	b.panicIfNotOfType(expectedType)

	formatRaw := jsonGetMap(b.RawJSON, "format")
	if len(formatRaw) == 0 {
		return false
	}
	err := jsonUnmarshalFromMap(formatRaw, v)
	if err != nil {
		panic(err)
	}
	return true
}

// FormatPage returns decoded format property for BlockPage
func (b *Block) FormatPage() *FormatPage {
	var format FormatPage
	if ok := b.unmarshalFormat(BlockPage, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) FormatBookmark() *FormatBookmark {
	var format FormatBookmark
	if ok := b.unmarshalFormat(BlockBookmark, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) FormatImage() *FormatImage {
	// TODO: no longer does
	// format.ImageURL = maybeProxyImageURL(format.DisplaySource)
	var format FormatImage
	if ok := b.unmarshalFormat(BlockImage, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) FormatColumn() *FormatColumn {
	var format FormatColumn
	if ok := b.unmarshalFormat(BlockColumn, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) FormatTable() *FormatTable {
	var format FormatTable
	if ok := b.unmarshalFormat(BlockTable, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) FormatText() *FormatText {
	var format FormatText
	if ok := b.unmarshalFormat(BlockText, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) FormatVideo() *FormatVideo {
	var format FormatVideo
	if ok := b.unmarshalFormat(BlockVideo, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) FormatEmbed() *FormatEmbed {
	var format FormatEmbed
	if ok := b.unmarshalFormat(BlockEmbed, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) FormatHeader() *FormatHeader {
	var format FormatHeader
	if ok := b.unmarshalFormat(BlockHeader, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) FormatToggle() *FormatToggle {
	var format FormatToggle
	if ok := b.unmarshalFormat(BlockToggle, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) FormatNumberedList() *FormatNumberedList {
	var format FormatNumberedList
	if ok := b.unmarshalFormat(BlockNumberedList, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) FormatBulletedList() *FormatBulletedList {
	var format FormatBulletedList
	if ok := b.unmarshalFormat(BlockBulletedList, &format); !ok {
		return nil
	}
	return &format
}

func (b *Block) BlockByID(id string) *Block {
	return b.Page.BlockByID(id)
}

func (b *Block) UserByID(id string) *User {
	return b.Page.UserByID(id)
}

func (b *Block) CollectionByID(id string) *Collection {
	return b.Page.CollectionByID(id)
}

func (b *Block) CollectionViewByID(id string) *CollectionView {
	return b.Page.CollectionViewByID(id)
}

func getBlockIDsSorted(idToBlock map[string]*Block) []string {
	// we want to serialize in a fixed order
	n := len(idToBlock)
	if n == 0 {
		return nil
	}
	ids := make([]string, n, n)
	i := 0
	for id := range idToBlock {
		ids[i] = id
		i++
	}
	sort.Strings(ids)
	return ids
}
