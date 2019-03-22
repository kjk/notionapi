package notionapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
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

var (
    dashIDLen = len("2131b10c-ebf6-4938-a127-7089ff02dbe4")
    noDashIDLen = len("2131b10cebf64938a1277089ff02dbe4")
)

// NormalizeID is deprecated. Use ToDashID instead.
func NormalizeID(s string) (string, bool) {
    return ToDashID(s), true
}

// ToNoDashID converts  2131b10c-ebf6-4938-a127-7089ff02dbe4
// to 2131b10cebf64938a1277089ff02dbe4.
// If not in expected format, we leave it untouched
func ToNoDashID(id string) string {
	s := strings.Replace(id, "-", "", -1)
	if len(s) != 32 {
		return id
	}
    return s
}

// ToDashID convert id in format bb760e2dd6794b64b2a903005b21870a
// to bb760e2d-d679-4b64-b2a9-03005b21870a
// If id is not in that format, we leave it untouched.
func ToDashID(id string) string {
	s := strings.Replace(id, "-", "", -1)
	if len(s) != noDashIDLen {
		return id
	}
	res := id[:8] + "-" + id[8:12] + "-" + id[12:16] + "-" + id[16:20] + "-" + id[20:]
	return res
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

// DownloadPage returns Notion page data given its id
func (c *Client) DownloadPage(pageID string) (*Page, error) {
	normalizedPageID, ok := NormalizeID(pageID)
	if !ok {
		return nil, fmt.Errorf("%s is not a valid Notion page id", pageID)
	}
	pageID = normalizedPageID

	page := &Page{
		client: c,
	}

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
	return page, nil
}
