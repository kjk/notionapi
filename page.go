package notionapi

import (
	"errors"
	"fmt"
	"sort"
)

var (
	// TODO: add more values, see FormatPage struct
	validFormatValues = map[string]struct{}{
		"page_full_width": struct{}{},
		"page_small_text": struct{}{},
	}
)

// Page describes a single Notion page
type Page struct {
	ID string
	// Users allows to find users that Page refers to by their ID
	Users  []*UserWithRole
	Tables []*Table

	idToBlock          map[string]*Block
	idToUser           map[string]*UserWithRole
	idToCollection     map[string]*Collection
	idToCollectionView map[string]*Block
	blocksToSkip       map[string]struct{} // not alive or when server doesn't return "value" for this block id

	client *Client
}

// BlockByID returns a block by its id
func (p *Page) BlockByID(id string) *Block {
	return p.idToBlock[ToDashID(id)]
}

// UserByID returns a user by its id
func (p *Page) UserByID(id string) *User {
	res := p.idToUser[ToDashID(id)]
	if res != nil {
		return res.Value
	}
	return nil
}

// CollectionByID returns a collection by its id
func (p *Page) CollectionByID(id string) *Collection {
	return p.idToCollection[ToDashID(id)]
}

// CollectionViewByID returns a collection view by its id
func (p *Page) CollectionViewByID(id string) *Block {
	return p.idToCollectionView[ToDashID(id)]
}

// Root returns a root block representing a page
func (p *Page) Root() *Block {
	return p.BlockByID(p.ID)
}

// Table represents a table (i.e. CollectionView)
type Table struct {
	CollectionView *Block      `json:"collection_view"`
	Collection     *Collection `json:"collection"`
	Data           []*Block
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

func forEachBlockWithParent(blocks []*Block, parent *Block, cb func(*Block)) {
	for _, block := range blocks {
		block.Parent = parent
		cb(block)
		forEachBlockWithParent(block.Content, block, cb)
	}
}

// ForEachBlock traverses the tree of blocks and calls cb on every block
// in depth-first order. To traverse every blocks in a Page, do:
// ForEachBlock([]*notionapi.Block{page.Root}, cb)
func ForEachBlock(blocks []*Block, cb func(*Block)) {
	forEachBlockWithParent(blocks, nil, cb)
}

func (p *Page) ForEachBlock(cb func(*Block)) {
	root := p.Root()
	cb(root)
	forEachBlockWithParent(root.Content, nil, cb)
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
	if block == nil || block.Type != BlockPage {
		return false
	}
	for {
		parentID := block.ParentID
		if parentID == p.ID {
			return true
		}
		parent := p.BlockByID(block.ParentID)
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
	return block.ID == p.ID
}

// GetSubPages return list of ids for direct sub-pages of this page
// TDOO: add pages that are title properties of collection_view
func (p *Page) GetSubPages() []string {
	root := p.Root()
	panicIf(root.Type != BlockPage)
	subPages := map[string]struct{}{}
	seenBlocks := map[string]struct{}{}
	blocksToVisit := append([]string{}, root.ContentIDs...)
	for len(blocksToVisit) > 0 {
		id := ToDashID(blocksToVisit[0])
		blocksToVisit = blocksToVisit[1:]
		if _, ok := seenBlocks[id]; ok {
			continue
		}
		seenBlocks[id] = struct{}{}
		block := p.BlockByID(id)
		if p.IsSubPage(block) {
			subPages[id] = struct{}{}
		}
		// need to recursively scan blocks with children
		blocksToVisit = append(blocksToVisit, block.ContentIDs...)
	}
	res := []string{}
	for id := range subPages {
		res = append(res, id)
	}
	sort.Strings(res)
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

func ResolveUser(page *Page, userID string) string {
	// TODO: either scan for user ids when initially downloading a page
	// or do a query if not found
	for _, u := range page.Users {
		user := u.Value
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
