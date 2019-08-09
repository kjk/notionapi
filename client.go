package notionapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	notionHost = "https://www.notion.so"
	// modern Chrome
	userAgent  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3483.0 Safari/537.36"
	acceptLang = "en-US,en;q=0.9"
)

// Client is client for invoking Notion API
type Client struct {
	// AuthToken allows accessing non-public pages.
	AuthToken string
	// HTTPClient allows over-riding http.Client to e.g. implement caching
	// on a per-request level
	HTTPClient *http.Client
	// Logger is used to log requests and responses for debugging.
	// By default is not set.
	Logger io.Writer
	// DebugLog enables debug logging
	DebugLog bool
}

func (c *Client) getHTTPClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	httpClient := *http.DefaultClient
	httpClient.Timeout = time.Second * 30
	return &httpClient
}

func doNotionAPI(c *Client, apiURL string, requestData interface{}, result interface{}) (map[string]interface{}, error) {
	var js []byte
	var err error
	if requestData != nil {
		js, err = json.Marshal(requestData)
		if err != nil {
			return nil, err
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
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", acceptLang)
	if c.AuthToken != "" {
		req.Header.Set("cookie", fmt.Sprintf("token_v2=%v", c.AuthToken))
	}
	var rsp *http.Response

	httpClient := c.getHTTPClient()
	rsp, err = httpClient.Do(req)

	if err != nil {
		log(c, "http.DefaultClient.Do() failed with %s\n", err)
		return nil, err
	}
	defer rsp.Body.Close()
	if rsp.StatusCode != 200 {
		d, _ := ioutil.ReadAll(rsp.Body)
		log(c, "Error: status code %s\nBody:\n%s\n", rsp.Status, ppJSON(d))
		return nil, fmt.Errorf("http.Post('%s') returned non-200 status code of %d", uri, rsp.StatusCode)
	}
	d, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log(c, "Error: ioutil.ReadAll() failed with %s\n", err)
		return nil, err
	}
	logJSON(c, d)
	err = json.Unmarshal(d, result)
	if err != nil {
		log(c, "Error: json.Unmarshal() failed with %s\n. Body:\n%s\n", err, string(d))
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(d, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

var (
	dashIDLen   = len("2131b10c-ebf6-4938-a127-7089ff02dbe4")
	noDashIDLen = len("2131b10cebf64938a1277089ff02dbe4")
)

// only hex chars seem to be valid
func isValidNoDashIDChar(c byte) bool {
	switch {
	case c >= '0' && c <= '9':
		return true
	case c >= 'a' && c <= 'f':
		return true
	case c >= 'A' && c <= 'F':
		// currently not used but just in case notion starts using them
		return true
	}
	return false
}

func isValidDashIDChar(c byte) bool {
	if c == '-' {
		return true
	}
	return isValidNoDashIDChar(c)
}

// IsValidDashID returns true if id looks like a valid Notion dash id
func IsValidDashID(id string) bool {
	if len(id) != dashIDLen {
		return false
	}
	if id[8] != '-' ||
		id[13] != '-' ||
		id[18] != '-' ||
		id[23] != '-' {
		return false
	}
	for i := range id {
		if !isValidDashIDChar(id[i]) {
			return false
		}
	}
	return true
}

// IsValidNoDashID returns true if id looks like a valid Notion no dash id
func IsValidNoDashID(id string) bool {
	if len(id) != noDashIDLen {
		return false
	}
	for i := range id {
		if !isValidNoDashIDChar(id[i]) {
			return false
		}
	}
	return true
}

// ToNoDashID converts 2131b10c-ebf6-4938-a127-7089ff02dbe4
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
	if IsValidDashID(id) {
		return id
	}
	s := strings.Replace(id, "-", "", -1)
	if len(s) != noDashIDLen {
		return id
	}
	res := id[:8] + "-" + id[8:12] + "-" + id[12:16] + "-" + id[16:20] + "-" + id[20:]
	return res
}

func isSafeChar(r rune) bool {
	if r >= '0' && r <= '9' {
		return true
	}
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	return false
}

// SafeName returns a file-system safe name
func SafeName(s string) string {
	var res string
	for _, r := range s {
		if !isSafeChar(r) {
			res += "-"
		} else {
			res += string(r)
		}
	}
	// replace multi-dash with single dash
	for strings.Contains(res, "--") {
		res = strings.Replace(res, "--", "-", -1)
	}
	res = strings.TrimLeft(res, "-")
	res = strings.TrimRight(res, "-")
	return res
}

// ExtractNoDashIDFromNotionURL tries to extract notion page id from
// notion URL, e.g. given:
// https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
// returns ea07db1b9bff415ab180b0525f3898f6
// returns "" if didn't detect valid notion id in the url
func ExtractNoDashIDFromNotionURL(uri string) string {
	maybeID := ToNoDashID(uri)
	if IsValidNoDashID(maybeID) {
		return maybeID
	}
	trimmed := strings.TrimPrefix(uri, "https://www.notion.so/")
	if uri == trimmed {
		return ""
	}
	// could be c674bebe8adf44d18c3a36cc18c131e2 from https://www.notion.so/c674bebe8adf44d18c3a36cc18c131e2
	id := trimmed
	parts := strings.Split(trimmed, "-")
	n := len(parts)
	if n >= 2 {
		// could be ea07db1b9bff415ab180b0525f3898f6 from Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
		id = parts[n-1]
	}
	id = ToNoDashID(id)
	if IsValidNoDashID(id) {
		return id
	}
	return ""
}

// recursively find blocks that we don't have yet
func (p *Page) findMissingBlocks(startIds []string) []string {
	var missing []string
	seen := map[string]struct{}{}
	toCheck := append([]string{}, startIds...)
	for len(toCheck) > 0 {
		id := toCheck[0]
		toCheck = toCheck[1:]
		if _, ok := seen[id]; ok {
			continue
		}
		if _, ok := p.blocksToSkip[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		block := p.idToBlock[id]
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
	id := ToDashID(pageID)
	if !IsValidDashID(id) {
		return nil, fmt.Errorf("%s is not a valid Notion page id", id)
	}
	pageID = id

	p := &Page{
		ID:                 pageID,
		client:             c,
		idToBlock:          map[string]*Block{},
		idToCollection:     map[string]*Collection{},
		idToCollectionView: map[string]*CollectionView{},
		idToUser:           map[string]*User{},
		blocksToSkip:       map[string]struct{}{},
	}

	var root *Block
	// get page's root block and then recursively download referenced blocks
	{
		recVals, err := c.GetRecordValues([]string{pageID})
		if err != nil {
			return nil, err
		}
		res := recVals.Results[0]
		// this might happen e.g. when a page is not publicly visible
		root = res.Value
		if root == nil {
			return nil, fmt.Errorf("Couldn't retrieve page with id %s", pageID)
		}
		panicIf(p.ID != root.ID, "%s != %s", p.ID, root.ID)
		p.idToBlock[root.ID] = root
	}

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
				p.idToBlock[id] = v.Value
			} else {
				p.blocksToSkip[id] = struct{}{}
			}
		}
		for id, v := range rsp.RecordMap.Collections {
			if v.Value.Alive {
				p.idToCollection[id] = v.Value
			}
			// TODO: what to do for not alive?
		}
		for id, v := range rsp.RecordMap.CollectionViews {
			if v.Value.Alive {
				p.idToCollectionView[id] = v.Value
			}
			// TODO: what to do for not alive?
		}
		for id, v := range rsp.RecordMap.Users {
			p.idToUser[id] = v.Value
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
		missing := p.findMissingBlocks(root.ContentIDs)
		if len(missing) == 0 {
			break
		}
		dbg(c, "DownloadPage: %d missing blocks in iteration %d\n", len(missing), missingIter)
		missingIter++

		// the API worked even with 6k items, but I'll split it into many
		// smaller requests anyway
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
					p.blocksToSkip[expectedID] = struct{}{}
					if n > 0 {
						prevBlock := recVals.Results[n-1]
						if prevBlock == nil || prevBlock.Value == nil {
							// this can happen if we don't have access to this page
							dbg(c, "prevBlock.Value is nil at position n = %d with expected id %s.\n", n, expectedID)
						} else {
							prevBlockID := prevBlock.Value.ID
							dbg(c, "block is nil at position n = %d with expected id %s. Prev block id: %s\n", n, expectedID, prevBlockID)
						}
					} else {
						dbg(c, "block is nil at position n = %v with expected id %s.\n", n, expectedID)
					}
					continue
				}

				id := block.ID
				p.idToBlock[id] = block
			}
		}
	}

	for _, v := range p.idToUser {
		p.Users = append(p.Users, v)
	}

	p.resolveBlocks()

	blockIDs := getBlockIDsSorted(p.idToBlock)
	for _, id := range blockIDs {
		block := p.idToBlock[id]
		if block.Type != BlockCollectionView {
			continue
		}
		if len(block.ViewIDs) == 0 {
			return nil, fmt.Errorf("collection_view has no ViewIDs")
		}
		// TODO: should fish out the user based on block.CreatedBy
		if len(p.Users) == 0 {
			return nil, fmt.Errorf("no users when trying to resolve collection_view")
		}

		collectionID := block.CollectionID
		for _, collectionViewID := range block.ViewIDs {
			user := p.Users[0]
			collectionView, ok := p.idToCollectionView[collectionViewID]
			if !ok {
				return nil, fmt.Errorf("Didn't find collection_view with id '%s'", collectionViewID)
			}
			collection, ok := p.idToCollection[collectionID]
			if !ok {
				//return nil, fmt.Errorf("Didn't find collection with id '%s'", collectionID)
				continue
			}
			var agg []*AggregateQuery
			if collectionView.Query != nil {
				agg = collectionView.Query.Aggregate
			}
			res, err := c.QueryCollection(collectionID, collectionViewID, agg, user)
			if err != nil {
				return nil, err
			}

			// this is repeated in unmarshalCollectionViewInfo
			collInfo := &CollectionViewInfo{
				CollectionView:          collectionView,
				Collection:              collection,
				queryCollectionResponse: res,
			}
			blockIds := res.Result.BlockIDS
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

	for _, b := range p.idToBlock {
		parseProperties(b)
		b.Page = p
	}

	return p, nil
}
