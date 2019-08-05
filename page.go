package notionapi

import (
	"encoding/json"
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
	idToCollection     map[string]*Collection
	idToCollectionView map[string]*CollectionView
	idToUser           map[string]*User
	blocksToSkip       map[string]struct{} // not alive or when server doesn't return "value" for this block id

	client *Client
}

// BlockByID returns a block by its id
func (p *Page) BlockByID(id string) *Block {
	return p.idToBlock[ToDashID(id)]
}

// Root returns a root block representing a page
func (p *Page) Root() *Block {
	return p.BlockByID(p.ID)
}

const (
	// current version of Page JSON serialization
	// allows changing the format in the future
	currPageJSONVersion = "1"
)

type pageMarshaled struct {
	Version    string
	RootPageID string
	Blocks     []map[string]interface{}
}

func (p *Page) MarshalJSON() ([]byte, error) {
	v := pageMarshaled{
		Version:    currPageJSONVersion,
		RootPageID: p.ID,
	}
	// we want to serialize in a fixed order
	var ids []string
	for id := range p.idToBlock {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		block := p.idToBlock[id]
		v.Blocks = append(v.Blocks, block.RawJSON)
	}
	return json.MarshalIndent(v, "", "  ")
}

func (p *Page) UnmarshalJSON(data []byte) error {
	var v pageMarshaled
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}
	if v.Version != currPageJSONVersion {
		return fmt.Errorf("expected serialization format version '%s', got '%s'", currPageJSONVersion, v.Version)
	}

	p.idToBlock = map[string]*Block{}
	p.ID = v.RootPageID
	for _, blockJSON := range v.Blocks {
		var b Block
		err = jsonUnmarshalFromMap(blockJSON, &b)
		if err != nil {
			return err
		}
		b.RawJSON = blockJSON
		p.idToBlock[b.ID] = &b
	}
	return nil
}

// GetBlockByID returns Block given it's id
func (p *Page) GetBlockByID(blockID string) *Block {
	return p.idToBlock[blockID]
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
