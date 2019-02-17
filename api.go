package notionapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	notionHost = "https://www.notion.so"
	// modern Chrome
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3483.0 Safari/537.36"
	acceptLang = "en-US,en;q=0.9"
)

// HTTPInterceptor allows intercepting HTTP request so that a client
// of this library can provide e.g. a caching system for requests
// instead
type HTTPInterceptor interface {
	// OnRequest is called before http request is sent tot he server
	// If it returns non-nil response, it'll be used instead of sending
	// a request to the server
	OnRequest(*http.Request) *http.Response
	// OnResponse is called after getting a response from the server
	// to allow e.g. caching of responses
	// Only called if the request was sent to the server (i.e. doesn't come
	// from OnRequest)
	OnResponse(*http.Response)
}

// Client is client for invoking Notion API
type Client struct {
	// AuthToken allows accessing non-public pages.
	AuthToken string
	// HTTPClient allows over-riding http.Client
	HTTPClient *http.Client
	// HTTPIntercept allows intercepting http requests
	// e.g. to implement caching
	HTTPIntercept HTTPInterceptor
	// Logger is used to log requests and responses for debugging.
	// By default is not set.
	Logger io.Writer
	// DebugLog enables debug logging
	DebugLog bool
}

// Page describes a single Notion page
type Page struct {
	ID string
	// Root is a root block representing a page
	Root *Block
	// Users allows to find users that Page refers to by their ID
	Users  []*User
	Tables []*Table
}

// Table represents a table (i.e. CollectionView)
type Table struct {
	CollectionView *CollectionView `json:"collection_view"`
	Collection     *Collection     `json:"collection"`
	Data           []*Block
}

func doNotionAPI(c *Client, apiURL string, requestData interface{}, result interface{}) error {
	var js []byte
	var err error
	if requestData != nil {
		js, err = json.Marshal(requestData)
		if err != nil {
			return err
		}
	}
	uri := notionHost + apiURL
	body := bytes.NewBuffer(js)
	log(c, "POST %s\n", uri)
	if len(js) > 0 {
		logJSON(c, js)
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", acceptLang)
	if c.AuthToken != "" {
		req.Header.Set("cookie", fmt.Sprintf("token_v2=%v", c.AuthToken))
	}
	var rsp *http.Response
	if c.HTTPIntercept != nil {
		rsp = c.HTTPIntercept.OnRequest(req)
	}

	realHTTPRequest := false
	if rsp == nil {
		realHTTPRequest = true
		httpClient := c.HTTPClient
		if httpClient == nil {
			httpClient = http.DefaultClient
		}
		rsp, err = httpClient.Do(req)
	}

	if err != nil {
		log(c, "http.DefaultClient.Do() failed with %s\n", err)
		return err
	}
	if c.HTTPIntercept != nil && realHTTPRequest {
		c.HTTPIntercept.OnResponse(rsp)
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		d, _ := ioutil.ReadAll(rsp.Body)
		log(c, "Error: status code %s\nBody:\n%s\n", rsp.Status, ppJSON(d))
		return fmt.Errorf("http.Post('%s') returned non-200 status code of %d", uri, rsp.StatusCode)
	}
	d, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log(c, "Error: ioutil.ReadAll() failed with %s\n", err)
		return err
	}
	logJSON(c, d)
	err = json.Unmarshal(d, result)
	if err != nil {
		log(c, "Error: json.Unmarshal() failed with %s\n. Body:\n%s\n", err, string(d))
	}
	return err
}

func getFirstInline(inline []*InlineBlock) string {
	if len(inline) == 0 {
		return ""
	}
	return inline[0].Text
}

func getFirstInlineBlock(v interface{}) (string, error) {
	inline, err := parseInlineBlocks(v)
	if err != nil {
		return "", err
	}
	return getFirstInline(inline), nil
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

func parseProperties(block *Block) error {
	var err error
	props := block.Properties

	if title, ok := props["title"]; ok {
		if block.Type == BlockPage {
			block.Title, err = getFirstInlineBlock(title)
		} else if block.Type == BlockCode {
			block.Code, err = getFirstInlineBlock(title)
		} else {
			block.InlineContent, err = parseInlineBlocks(title)
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
		block.ImageURL = makeImageURL(block.Source)
	}

	// for BlockCode
	getProp(block, "language", &block.CodeLanguage)

	// for BlockFile
	if block.Type == BlockFile {
		getProp(block, "size", &block.FileSize)
	}

	return nil
}

// sometimes image url in "source" is not accessible but can
// be accessed when proxied via notion server as
// www.notion.so/image/${source}
// This also allows resizing via ?width=${n} arguments
//
// from: /images/page-cover/met_vincent_van_gogh_cradle.jpg
// =>
// https://www.notion.so/image/https%3A%2F%2Fwww.notion.so%2Fimages%2Fpage-cover%2Fmet_vincent_van_gogh_cradle.jpg?width=3290
func makeImageURL(uri string) string {
	if uri == "" || strings.Contains(uri, "//www.notion.so/image/") {
		return uri
	}
	// if the url has https://, it's already in s3.
	// If not, it's only a relative URL (like those for built-in
	// cover pages)
	if !strings.HasPrefix(uri, "https://") {
		uri = "https://www.notion.so" + uri
	}
	return "https://www.notion.so/image/" + url.PathEscape(uri)
}

func parseFormat(block *Block) error {
	if len(block.FormatRaw) == 0 {
		// TODO: maybe if BlockPage, set to default &FormatPage{}
		return nil
	}
	var err error
	switch block.Type {
	case BlockPage:
		var format FormatPage
		err = json.Unmarshal(block.FormatRaw, &format)
		if err == nil {
			format.PageCoverURL = makeImageURL(format.PageCover)
			block.FormatPage = &format
		}
	case BlockBookmark:
		var format FormatBookmark
		err = json.Unmarshal(block.FormatRaw, &format)
		if err == nil {
			block.FormatBookmark = &format
		}
	case BlockImage:
		var format FormatImage
		err = json.Unmarshal(block.FormatRaw, &format)
		if err == nil {
			format.ImageURL = makeImageURL(format.DisplaySource)
			block.FormatImage = &format
		}
	case BlockColumn:
		var format FormatColumn
		err = json.Unmarshal(block.FormatRaw, &format)
		if err == nil {
			block.FormatColumn = &format
		}
	case BlockTable:
		var format FormatTable
		err = json.Unmarshal(block.FormatRaw, &format)
		if err == nil {
			block.FormatTable = &format
		}
	case BlockText:
		var format FormatText
		err = json.Unmarshal(block.FormatRaw, &format)
		if err == nil {
			block.FormatText = &format
		}
	case BlockVideo:
		var format FormatVideo
		err = json.Unmarshal(block.FormatRaw, &format)
		if err == nil {
			block.FormatVideo = &format
		}
	case BlockEmbed:
		var format FormatEmbed
		err = json.Unmarshal(block.FormatRaw, &format)
		if err == nil {
			block.FormatEmbed = &format
		}
	}

	if err != nil {
		fmt.Printf("parseFormat: json.Unamrshal() failed with '%s', format: '%s'\n", err, string(block.FormatRaw))
		return err
	}
	return nil
}

var segments = []int{8, 4, 4, 4}

// NormalizeID converts 2131b10cebf64938a1277089ff02dbe4
// 2131b10c-ebf6-4938-a127-7089ff02dbe4
func NormalizeID(s string) (string, bool) {
	s = strings.Replace(s, "-", "", -1)
	if len(s) != 32 {
		return s, false
	}
	var parts [5]string
	for i, n := range segments {
		part := s[0:n]
		s = s[n:]
		parts[i] = part
	}
	parts[4] = s
	return strings.Join(parts[:], "-"), true
}

func resolveBlocks(block *Block, idToBlock map[string]*Block) error {
	err := parseProperties(block)
	if err != nil {
		return err
	}
	err = parseFormat(block)
	if err != nil {
		return err
	}

	if block.Content != nil || len(block.ContentIDs) == 0 {
		return nil
	}
	n := len(block.ContentIDs)
	block.Content = make([]*Block, n, n)
	notResolved := []int{}
	for i, id := range block.ContentIDs {
		resolved := idToBlock[id]
		if resolved == nil {
			// This can happen e.g. for page fa3fc358e5644f39b89c57f13d426d54
			notResolved = append(notResolved, i)
			//return fmt.Errorf("Couldn't resolve block with id '%s'", id)
			continue
		}
		block.Content[i] = resolved
		resolveBlocks(resolved, idToBlock)
	}
	// remove blocks that are not resolved
	for idx, toRemove := range notResolved {
		i := toRemove - idx
		{
			a := block.ContentIDs
			block.ContentIDs = append(a[:i], a[i+1:]...)
		}
		{
			a := block.Content
			block.Content = append(a[:i], a[i+1:]...)
		}
	}
	return nil
}

// recursively find blocks that we don't have yet
func findMissingBlocks(startIds []string, idToBlock map[string]*Block, blocksToSkip map[string]struct{}) []string {
	var missing []string
	seen := map[string]struct{}{}
	toCheck := append([]string{}, startIds...)
	for len(toCheck) > 0 {
		id := toCheck[0]
		toCheck = toCheck[1:]
		if _, ok := seen[id]; ok {
			continue
		}
		if _, ok := blocksToSkip[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		block := idToBlock[id]
		if block == nil {
			missing = append(missing, id)
			continue
		}

		// we don't have the content of this block.
		// get it unless this is a page block becuase this is only
		// a link to a page
		switch block.Type {
		case BlockPage:
		// skip those blocks
		default:
			toCheck = append(toCheck, block.ContentIDs...)
		}
	}
	return missing
}

// DownloadPage returns Noion page data given its id
func (c *Client) DownloadPage(pageID string) (*Page, error) {
	normalizedPageID, ok := NormalizeID(pageID)
	if !ok {
		return nil, fmt.Errorf("%s is not a valid Notion page id", pageID)
	}
	pageID = normalizedPageID

	var page Page
	{
		recVals, err := c.GetRecordValues([]string{pageID})
		if err != nil {
			return nil, err
		}
		res := recVals.Results[0]
		// this might happen e.g. when a page is not publicly visible
		if res.Value == nil {
			return nil, fmt.Errorf("Couldn't retrieve page with id %s", pageID)
		}
		pageID = res.Value.ID
		page.ID = pageID
		page.Root = recVals.Results[0].Value
	}

	idToBlock := map[string]*Block{}
	idToCollection := map[string]*Collection{}
	idToCollectionView := map[string]*CollectionView{}
	idToUser := map[string]*User{}
	// not alive or when server doesn't return "value" for this block id
	blocksToSkip := map[string]struct{}{}
	chunkNo := 0
	var cur *cursor
	for {
		rsp, err := c.LoadPageChunk(pageID, chunkNo, cur)
		chunkNo++
		if err != nil {
			return nil, err
		}
		for id, v := range rsp.RecordMap.Blocks {
			if v.Value.Alive {
				idToBlock[id] = v.Value
			} else {
				blocksToSkip[id] = struct{}{}
			}
		}
		for id, v := range rsp.RecordMap.Collections {
			if v.Value.Alive {
				idToCollection[id] = v.Value
			}
			// TODO: what to do for not alive?
		}
		for id, v := range rsp.RecordMap.CollectionViews {
			if v.Value.Alive {
				idToCollectionView[id] = v.Value
			}
			// TODO: what to do for not alive?
		}
		for id, v := range rsp.RecordMap.Users {
			idToUser[id] = v.Value
		}

		cursor := rsp.Cursor
		//dbg("GetPaDownloadPagegeInfo: len(cursor.Stack)=%d\n", len(cursor.Stack))
		if len(cursor.Stack) == 0 {
			break
		}
		cur = &rsp.Cursor
	}

	// get blocks that are not already loaded
	missingIter := 1
	for {
		missing := findMissingBlocks(page.Root.ContentIDs, idToBlock, blocksToSkip)
		if len(missing) == 0 {
			break
		}
		dbg(c, "DownloadPage: %d missing blocks in iteration %d\n", len(missing), missingIter)
		missingIter++

		// the API worked even with 6k items, but I'll split it into many
		// smaller requrests anyway
		maxToGet := 128 * 10
		for len(missing) > 0 {
			toGet := missing
			if len(toGet) > maxToGet {
				toGet = missing[:maxToGet]
				missing = missing[maxToGet:]
			} else {
				missing = nil
			}

			recVals, err := c.GetRecordValues(toGet)
			if err != nil {
				return nil, err
			}
			for n, blockWithRole := range recVals.Results {
				block := blockWithRole.Value
				// This can happen e.g. in 157765353f2c4705bd45474e5ba8b46c
				// Server returns { "role": "none" },
				if block == nil {
					expectedID := toGet[n]
					blocksToSkip[expectedID] = struct{}{}
					if n > 0 {
						prevBlock := recVals.Results[n-1]
						prevBlockID := prevBlock.Value.ID
						dbg(c, "block is nil at position n = %v with expected id %s. Prev block id: %s\n", n, expectedID, prevBlockID)
					} else {
						dbg(c, "block is nil at position n = %v with expected id %s.\n", n, expectedID)
					}
					continue
				}

				id := block.ID
				idToBlock[id] = block
			}
		}
	}

	for _, v := range idToUser {
		page.Users = append(page.Users, v)
	}

	err := resolveBlocks(page.Root, idToBlock)
	if err != nil {
		return nil, err
	}

	for _, block := range page.Root.Content {
		if block.Type != BlockCollectionView {
			continue
		}
		if len(block.ViewIDs) == 0 {
			return nil, fmt.Errorf("collection_view has no ViewIDs")
		}
		// TODO: should fish out the user based on block.CreatedBy
		if len(page.Users) == 0 {
			return nil, fmt.Errorf("no users when trying to resolve collection_view")
		}

		collectionID := block.CollectionID
		for _, collectionViewID := range block.ViewIDs {
			user := page.Users[0]
			collectionView, ok := idToCollectionView[collectionViewID]
			if !ok {
				return nil, fmt.Errorf("Didn't find collection_view with id '%s'", collectionViewID)
			}
			collection, ok := idToCollection[collectionID]
			if !ok {
				return nil, fmt.Errorf("Didn't find collection with id '%s'", collectionID)
			}
			var agg []*AggregateQuery
			if collectionView.Query != nil {
				agg = collectionView.Query.Aggregate
			}
			res, err := c.QueryCollection(collectionID, collectionViewID, agg, user)
			if err != nil {
				return nil, err
			}
			blockIds := res.Result.BlockIDS
			collInfo := &CollectionViewInfo{
				CollectionView: collectionView,
				Collection:     collection,
			}
			for _, id := range blockIds {
				rowBlock, ok := res.RecordMap.Blocks[id]
				if !ok {
					return nil, fmt.Errorf("didn't find block with id '%s' for collection view with id '%s'", id, collectionViewID)
				}
				collInfo.CollectionRows = append(collInfo.CollectionRows, rowBlock.Value)
			}
			block.CollectionViews = append(block.CollectionViews, collInfo)
		}
	}
	return &page, nil
}
