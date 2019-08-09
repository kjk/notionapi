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
	Users  []*User
	Tables []*Table

	idToBlock          map[string]*Block
	idToUser           map[string]*User
	idToCollection     map[string]*Collection
	idToCollectionView map[string]*CollectionView
	blocksToSkip       map[string]struct{} // not alive or when server doesn't return "value" for this block id

	client *Client
}

// BlockByID returns a block by its id
func (p *Page) BlockByID(id string) *Block {
	return p.idToBlock[ToDashID(id)]
}

// UserByID returns a user by its id
func (p *Page) UserByID(id string) *User {
	return p.idToUser[ToDashID(id)]
}

// CollectionByID returns a collection by its id
func (p *Page) CollectionByID(id string) *Collection {
	return p.idToCollection[ToDashID(id)]
}

// CollectionViewByID returns a collection view by its id
func (p *Page) CollectionViewByID(id string) *CollectionView {
	return p.idToCollectionView[ToDashID(id)]
}

// Root returns a root block representing a page
func (p *Page) Root() *Block {
	return p.BlockByID(p.ID)
}

// Table represents a table (i.e. CollectionView)
type Table struct {
	CollectionView *CollectionView `json:"collection_view"`
	Collection     *Collection     `json:"collection"`
	Data           []*Block
}

// SetTitle changes page title
func (p *Page) SetTitle(s string) error {
	op := buildSetTitleOp(p.ID, s)
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
	op := buildSetPageFormat(p.ID, args)
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

// GetSubPages return list of ids for pages reachable from those block
func GetSubPages(blocks []*Block) []string {
	pageIDs := map[string]struct{}{}
	seen := map[string]struct{}{}
	toVisit := blocks
	for len(toVisit) > 0 {
		block := toVisit[0]
		toVisit = toVisit[1:]
		id := ToNoDashID(block.ID)
		if block.Type == BlockPage {
			pageIDs[id] = struct{}{}
			seen[id] = struct{}{}
		}
		for _, b := range block.Content {
			if b == nil {
				continue
			}
			id := ToNoDashID(block.ID)
			if _, ok := seen[id]; ok {
				continue
			}
			toVisit = append(toVisit, b)
		}
	}
	res := []string{}
	for id := range pageIDs {
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
		if u.ID == userID {
			return makeUserName(u)
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
