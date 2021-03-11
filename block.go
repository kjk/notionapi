package notionapi

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

const (
	// BlockAudio is audio embed (e.g. an mp3 file)
	BlockAudio = "audio"
	// BlockBookmark is a bookmark block
	BlockBookmark = "bookmark"
	// BlockBreadcrumb is breadcrumb block
	BlockBreadcrumb = "breadcrumb"
	// BlockBulletedList is a bulleted list block
	BlockBulletedList = "bulleted_list"
	// BlockCode is a code block
	BlockCode = "code"
	// BlockCodepen is embedded codepen block
	BlockCodepen = "codepen"
	// BlockCallout is a callout
	BlockCallout = "callout"
	// BlockColumn is a child of TypeColumnList
	BlockColumn = "column"
	// BlockColumnList is for multi-column. Number of columns is
	// number of content blocks of type TypeColumn
	BlockColumnList = "column_list"
	// BlockCollectionView is a collection view block for inline collections
	BlockCollectionView = "collection_view"
	// BlockCollectionViewPage is a page that is a collection
	BlockCollectionViewPage = "collection_view_page"
	// BlockComment is a comment block
	BlockComment = "comment"
	// BlockDivider is a divider block
	BlockDivider = "divider"
	// BlockDrive is embedded Google Drive file
	BlockDrive = "drive"
	// BlockEmbed is a generic oembed link
	BlockEmbed = "embed"
	// BlockEquation is TeX equation block
	BlockEquation = "equation"
	// BlockFactory represents a factory block
	BlockFactory = "factory"
	// BlockFigma represents figma embed
	BlockFigma = "figma"
	// BlockFile is an embedded file
	BlockFile = "file"
	// BlockGist is embedded gist block
	BlockGist = "gist"
	// BlockHeader is a header block
	BlockHeader = "header"
	// BlockImage is an image block
	BlockImage = "image"
	// BlockMaps is embedded Google Map block
	BlockMaps = "maps"
	// BlockNumberedList is a numbered list block
	BlockNumberedList = "numbered_list"
	// BlockPDF is an embedded pdf file
	BlockPDF = "pdf"
	// BlockPage is a notion Page
	BlockPage = "page"
	// BlockQuote is a quote block
	BlockQuote = "quote"
	// BlockSubHeader is a header block
	BlockSubHeader = "sub_header"
	// BlockSubSubHeader
	BlockSubSubHeader = "sub_sub_header"
	// BlockTableOfContents is table of contents
	BlockTableOfContents = "table_of_contents"
	// BlockText is a text block
	BlockText = "text"
	// BlockTodo is a todo block
	BlockTodo = "to_do"
	// BlockToggle is a toggle block
	BlockToggle = "toggle"
	// BlockTweet is embedded gist block
	BlockTweet = "tweet"
	// BlockVideo is youtube video embed
	BlockVideo = "video"
)

// FormatBookmark describes format for BlockBookmark
type FormatBookmark struct {
	BlockColor string `json:"block_color"`
	Cover      string `json:"bookmark_cover"`
	Icon       string `json:"bookmark_icon"`
}

// FormatBulletedList describes format for BlockBulletedList
type FormatBulletedList struct {
	BlockColor string `json:"block_color"`
}

// FormatCallout describes format for BlockCallout
type FormatCallout struct {
	BlockColor string `json:"block_color"`
	Icon       string `json:"bookmark_icon"`
}

// FormatCode describes format for BlockCode
type FormatCode struct {
	CodeWrap bool `json:"code_wrap"`
}

type FormatCodepen struct {
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockHeight        float64 `json:"block_height"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
	BlockWidth         float64 `json:"block_width"`
	DisplaySource      string  `json:"display_source,omitempty"`
}

type FormatCollectionView struct {
	BlockFullWidth bool    `json:"block_full_width"`
	BlockHeight    float64 `json:"block_height"`
	BlockPageWidth bool    `json:"block_page_width"`
	BlockWidth     float64 `json:"block_width"`
}

// FormatColumn describes format for BlockColumn
type FormatColumn struct {
	// e.g. 0.5 for half-sized column
	ColumnRatio float64 `json:"column_ratio"`
}

type DriveProperties struct {
	FileID       string `json:"file_id"`
	Icon         string `json:"icon"`
	ModifiedTime int64  `json:"modified_time"`
	Thumbnail    string `json:"thumbnail"` // url
	Title        string `json:"title"`
	Trashed      bool   `json:"trashed"`
	URL          string `json:"url"`
	UserName     string `json:"user_name"`
	Version      int    `json:"version"`
}

type DriveStatus struct {
	Authed      bool  `json:"authed"`
	LastFetched int64 `json:"last_fetched"`
}

type FormatDrive struct {
	DriveProperties *DriveProperties `json:"drive_properties"`
	DriveStatus     *DriveStatus     `json:"drive_status"`
}

// FormatEmbed describes format for BlockEmbed
type FormatEmbed struct {
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockHeight        float64 `json:"block_height"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
	BlockWidth         float64 `json:"block_width"`
	DisplaySource      string  `json:"display_source"`
}

type FormatFigma struct {
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockHeight        float64 `json:"block_height"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
	BlockWidth         float64 `json:"block_width"`
	DisplaySource      string  `json:"display_source"`
}

// FormatHeader describes format for BlockHeader, BlockSubHeader, BlockSubSubHeader
type FormatHeader struct {
	BlockColor string `json:"block_color,omitempty"`
}

// FormatImage describes format for BlockImage
type FormatImage struct {
	// comes from notion API
	BlockAspectRatio   float64 `json:"block_aspect_ratio"`
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockHeight        float64 `json:"block_height"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
	BlockWidth         float64 `json:"block_width"`
	DisplaySource      string  `json:"display_source,omitempty"`

	// calculated by us
	ImageURL string `json:"image_url,omitempty"`
}

type FormatMaps struct {
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockHeight        float64 `json:"block_height"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
	BlockWidth         float64 `json:"block_width"`
	DisplaySource      string  `json:"display_source,omitempty"`
}

// FormatNumberedList describes format for BlockNumberedList
type FormatNumberedList struct {
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

	BlockLocked   bool   `json:"block_locked"`
	BlockLockedBy string `json:"block_locked_by"`

	// calculated by us
	PageCoverURL string `json:"page_cover_url,omitempty"`
}

type FormatPDF struct {
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockHeight        float64 `json:"block_height"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
	BlockWidth         float64 `json:"block_width"`
}

type FormatTableOfContents struct {
	BlockColor string `json:"block_color,omitempty"`
}

// FormatText describes format for BlockText
type FormatText struct {
	BlockColor string `json:"block_color,omitempty"`
}

// FormatToggle describes format for BlockToggle
type FormatToggle struct {
	BlockColor string `json:"block_color"`
}

// FormatVideo describes fromat form BlockVideo
type FormatVideo struct {
	BlockAspectRatio   float64 `json:"block_aspect_ratio"`
	BlockFullWidth     bool    `json:"block_full_width"`
	BlockHeight        int64   `json:"block_height"`
	BlockPageWidth     bool    `json:"block_page_width"`
	BlockPreserveScale bool    `json:"block_preserve_scale"`
	BlockWidth         int64   `json:"block_width"`
	DisplaySource      string  `json:"display_source"`
}

const (
	// value of Permission.Type
	PermissionUser   = "user_permission"
	PermissionPublic = "public_permission"
)

// Permission represents user permissions o
type Permission struct {
	Type string `json:"type"`

	// common to some permission types
	Role string `json:"role"`

	// if Type == "user_permission"
	UserID *string `json:"user_id,omitempty"`

	// if Type == "public_permission"
	AllowDuplicate            bool `json:"allow_duplicate"`
	AllowSearchEngineIndexing bool `json:"allow_search_engine_indexing"`
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

	CreatedByTable    string `json:"created_by_table"`     // e.g. "notion_user"
	CreatedByID       string `json:"created_by_id"`        // e.g. "bb760e2d-d679-4b64-b2a9-03005b21870a",
	LastEditedByTable string `json:"last_edited_by_table"` // e.g. "notion_user"
	LastEditedByID    string `json:"last_edited_by_id"`    // e.g. "bb760e2d-d679-4b64-b2a9-03005b21870a"

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
	Version int64 `json:"version"`
	// for BlockCollectionView
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

	// for BlockCollectionView. There can be multiple views
	// those correspond to ViewIDs
	TableViews []*TableView `json:"-"`

	Page *Page `json:"-"`

	// RawJSON represents Block as
	RawJSON map[string]interface{} `json:"-"`

	isResolved bool
}

func (b *Block) Prop(key string) (interface{}, bool) {
	parts := strings.Split(key, ".")
	js := b.RawJSON
	var ok bool
	lastIdx := len(parts) - 1
	for i := 0; i < lastIdx; i++ {
		key = parts[i]
		v := js[key]
		if v == nil {
			return nil, false
		}
		js, ok = v.(map[string]interface{})
		if !ok {
			return nil, false
		}
	}
	key = parts[lastIdx]
	v, ok := js[key]
	return v, ok
}

func (b *Block) PropAsString(key string) (string, bool) {
	v, ok := b.Prop(key)
	if !ok || v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// CreatedOn return the time the page was created
func (b *Block) CreatedOn() time.Time {
	return time.Unix(b.CreatedTime/1000, 0)
}

// LastEditedOn returns the time the page was last updated
func (b *Block) LastEditedOn() time.Time {
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

// IsEmbeddedType returns true if block represents an embedded type
func (b *Block) IsEmbeddedType() bool {
	switch b.Type {
	case BlockImage, BlockEmbed, BlockAudio,
		BlockFile, BlockVideo:
		return true
	}
	return false
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

func parseTitle(block *Block) error {
	// has already been parsed
	if block.InlineContent != nil {
		return nil
	}
	props := block.Properties
	title, ok := props["title"]
	if !ok {
		return nil
	}
	var err error

	block.InlineContent, err = ParseTextSpans(title)
	if err != nil {
		return err
	}
	switch block.Type {
	case BlockPage, BlockFile, BlockBookmark:
		block.Title, err = getInlineText(title)
	case BlockCode:
		block.Code, err = getInlineText(title)
	default:
	}
	return err
}

func parseProperties(block *Block) error {
	err := parseTitle(block)
	if err != nil {
		return err
	}

	props := block.Properties

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

func (b *Block) FormatBookmark() *FormatBookmark {
	var format FormatBookmark
	if ok := b.unmarshalFormat(BlockBookmark, &format); !ok {
		return nil
	}
	return &format
}

// FormatPage returns decoded format property for BlockPage
// TODO: maybe separate FormatCollectionViewPage
func (b *Block) FormatPage() *FormatPage {
	var format FormatPage
	if b.Type == BlockPage {
		if ok := b.unmarshalFormat(BlockPage, &format); !ok {
			return nil
		}
	} else if b.Type == BlockCollectionViewPage {
		if ok := b.unmarshalFormat(BlockCollectionViewPage, &format); !ok {
			return nil
		}
	} else {
		b.panicIfNotOfType(BlockPage)
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

func (b *Block) FormatCallout() *FormatCallout {
	var format FormatCallout
	if ok := b.unmarshalFormat(BlockCallout, &format); !ok {
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
	ids := make([]string, n)
	i := 0
	for id := range idToBlock {
		ids[i] = id
		i++
	}
	sort.Strings(ids)
	return ids
}
