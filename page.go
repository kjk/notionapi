package notionapi

import (
	"errors"
	"fmt"
	"sort"
)

var (
	// TODO: add more values, see FormatPage struct
	validFormatValues = map[string]struct{}{
		"page_full_width": {},
		"page_small_text": {},
	}
)

// Page describes a single Notion page
type Page struct {
	ID       string
	NotionID *NotionID

	// expose raw records for all data associated with this page
	BlockRecords          []*Record
	UserRecords           []*Record
	CollectionRecords     []*Record
	CollectionViewRecords []*Record
	DiscussionRecords     []*Record
	CommentRecords        []*Record
	SpaceRecords          []*Record

	// for every block of type collection_view and its view_ids
	// we } TableView representing that collection view_id
	TableViews []*TableView

	idToBlock          map[string]*Block
	idToUser           map[string]*User
	idToCollection     map[string]*Collection
	idToCollectionView map[string]*CollectionView
	idToComment        map[string]*Comment
	idToDiscussion     map[string]*Discussion
	idToSpace          map[string]*Space

	blocksToSkip map[string]struct{} // not alive or when server doesn't return "value" for this block id

	client   *Client
	subPages []*NotionID
}

func (p *Page) GetNotionID() *NotionID {
	if p.NotionID == nil {
		p.NotionID = NewNotionID(p.ID)
	}
	return p.NotionID
}

// SpaceByID returns a space by its id
func (p *Page) SpaceByID(nid *NotionID) *Space {
	return p.idToSpace[nid.DashID]
}

// BlockByID returns a block by its id
func (p *Page) BlockByID(nid *NotionID) *Block {
	return p.idToBlock[nid.DashID]
}

// UserByID returns a user by its id
func (p *Page) UserByID(nid *NotionID) *User {
	return p.idToUser[nid.DashID]
}

// CollectionByID returns a collection by its id
func (p *Page) CollectionByID(nid *NotionID) *Collection {
	return p.idToCollection[nid.DashID]
}

// CollectionViewByID returns a collection view by its id
func (p *Page) CollectionViewByID(nid *NotionID) *CollectionView {
	return p.idToCollectionView[nid.DashID]
}

// DiscussionByID returns a discussion by its id
func (p *Page) DiscussionByID(nid *NotionID) *Discussion {
	return p.idToDiscussion[nid.DashID]
}

// CommentByID returns a comment by its id
func (p *Page) CommentByID(nid *NotionID) *Comment {
	return p.idToComment[nid.DashID]
}

// Root returns a root block representing a page
func (p *Page) Root() *Block {
	return p.BlockByID(p.GetNotionID())
}

// SetTitle changes page title
func (p *Page) SetTitle(s string) error {
	op := p.Root().SetTitleOp(s)
	ops := []*Operation{op}
	return p.client.SubmitTransaction(ops)
}

// SetFormat changes format properties of a page. Valid values are:
// page_full_width (bool), page_small_text (bool)
func (p *Page) SetFormat(args map[string]interface{}) error {
	if len(args) == 0 {
		return errors.New("args can't be empty")
	}
	for k := range args {
		if _, ok := validFormatValues[k]; !ok {
			return fmt.Errorf("'%s' is not a valid page format property", k)
		}
	}
	op := p.Root().UpdateFormatOp(args)
	ops := []*Operation{op}
	return p.client.SubmitTransaction(ops)
}

// NotionURL returns url of this page on notion.so
func (p *Page) NotionURL() string {
	if p == nil {
		return ""
	}
	id := ToNoDashID(p.ID)
	// TODO: maybe add title?
	return "https://www.notion.so/" + id
}

func forEachBlockWithParent(seen map[string]bool, blocks []*Block, parent *Block, cb func(*Block)) {
	for _, block := range blocks {
		id := block.ID
		if seen[id] {
			// crash rather than have infinite recursion
			panic("seen the same page again")
		}
		if parent != nil && (block.Type == BlockPage || block.Type == BlockCollectionViewPage) {
			// skip sub-pages to avoid infnite recursion
			continue
		}
		seen[id] = true
		block.Parent = parent
		cb(block)
		forEachBlockWithParent(seen, block.Content, block, cb)
	}
}

// ForEachBlock traverses the tree of blocks and calls cb on every block
// in depth-first order. To traverse every blocks in a Page, do:
// ForEachBlock([]*notionapi.Block{page.Root}, cb)
func ForEachBlock(blocks []*Block, cb func(*Block)) {
	seen := map[string]bool{}
	forEachBlockWithParent(seen, blocks, nil, cb)
}

// ForEachBlock recursively calls cb for each block in the page
func (p *Page) ForEachBlock(cb func(*Block)) {
	seen := map[string]bool{}
	blocks := []*Block{p.Root()}
	forEachBlockWithParent(seen, blocks, nil, cb)
}

func panicIf(cond bool, args ...interface{}) {
	if !cond {
		return
	}
	if len(args) == 0 {
		panic("condition failed")
	}
	format := args[0].(string)
	if len(args) == 1 {
		panic(format)
	}
	panic(fmt.Sprintf(format, args[1:]))
}

// IsSubPage returns true if a given block is BlockPage and
// a direct child of this page (as opposed to a link to
// arbitrary page)
func (p *Page) IsSubPage(block *Block) bool {
	if block == nil || !isPageBlock(block) {
		return false
	}

	for {
		parentID := block.ParentID
		if parentID == p.ID {
			return true
		}
		parent := p.BlockByID(block.GetParentNotionID())
		if parent == nil {
			return false
		}
		// parent is page but not our page, so it can't be sub-page
		if parent.Type == BlockPage {
			return false
		}
		block = parent
	}
}

// IsRoot returns true if this block is root block of the page
// i.e. of type BlockPage and very first block
func (p *Page) IsRoot(block *Block) bool {
	if block == nil || block.Type != BlockPage {
		return false
	}
	// a block can be a link to its parent, causing infinite loop
	// https://github.com/kjk/notionapi/issues/21
	// TODO: why block.ID == block.ParentID doesn't work?
	if block == block.Parent {
		return false
	}
	return block.ID == p.ID
}

func isPageBlock(block *Block) bool {
	switch block.Type {
	case BlockPage, BlockCollectionViewPage:
		return true
	}
	return false
}

// GetSubPages return list of ids for direct sub-pages of this page
func (p *Page) GetSubPages() []*NotionID {
	if len(p.subPages) > 0 {
		return p.subPages
	}
	root := p.Root()
	panicIf(!isPageBlock(root))
	subPages := map[*NotionID]struct{}{}
	seenBlocks := map[string]struct{}{}
	var blocksToVisit []*NotionID
	for _, id := range root.ContentIDs {
		nid := NewNotionID(id)
		blocksToVisit = append(blocksToVisit, nid)
	}
	for len(blocksToVisit) > 0 {
		nid := blocksToVisit[0]
		id := nid.DashID
		blocksToVisit = blocksToVisit[1:]
		if _, ok := seenBlocks[id]; ok {
			continue
		}
		seenBlocks[id] = struct{}{}
		block := p.BlockByID(nid)
		if p.IsSubPage(block) {
			subPages[nid] = struct{}{}
		}
		// need to recursively scan blocks with children
		for _, id := range block.ContentIDs {
			nid := NewNotionID(id)
			blocksToVisit = append(blocksToVisit, nid)
		}
	}
	res := []*NotionID{}
	for id := range subPages {
		res = append(res, id)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].DashID < res[j].DashID
	})
	p.subPages = res
	return res
}

func makeUserName(user *User) string {
	s := user.GivenName
	if len(s) > 0 {
		s += " "
	}
	s += user.FamilyName
	if len(s) > 0 {
		return s
	}
	return user.ID
}

// GetUserNameByID returns a full user name given user id
// it's a helper function
func GetUserNameByID(page *Page, userID string) string {
	for _, r := range page.UserRecords {
		user := r.User
		if user.ID == userID {
			return makeUserName(user)
		}
	}
	return userID
}

func (p *Page) resolveBlocks() error {
	for _, block := range p.idToBlock {
		err := resolveBlock(p, block)
		if err != nil {
			return err
		}
	}
	return nil
}

func resolveBlock(p *Page, block *Block) error {
	if block.isResolved {
		return nil
	}
	block.isResolved = true
	err := parseProperties(block)
	if err != nil {
		return err
	}

	var contentIDs []string
	var content []*Block
	for _, id := range block.ContentIDs {
		b := p.idToBlock[id]
		if b == nil {
			continue
		}
		contentIDs = append(contentIDs, id)
		content = append(content, b)
	}
	block.ContentIDs = contentIDs
	block.Content = content
	return nil
}
