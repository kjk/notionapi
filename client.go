package notionapi

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
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
	// HTTPClient allows over-riding http.Client
	HTTPClient *http.Client
	// Logger is used to log requests and responses for debugging.
	// By default is not set.
	Logger io.Writer
	// DebugLog enables debug logging
	DebugLog bool
	// MinRequestDelay is for controlling rate limiting. it's 333 ms by default
	// because https://developers.notion.com/reference/errors#rate-limits
	// says rate limit is, on average, 3 requests per second
	MinRequestDelay time.Duration
	// simplest rate limiting: track last request time and wait at least
	// MinRequestDelay between requests
	lastRequestTime time.Time

	httpPostOverride func(uri string, body []byte) ([]byte, error)
}

// vlogf is for verbose logging
func (c *Client) vlogf(format string, args ...interface{}) {
	if !c.DebugLog {
		return
	}
	c.logf(format, args...)
}

func (c *Client) logf(format string, args ...interface{}) {
	if c.Logger == nil {
		return
	}
	if len(args) == 0 {
		fmt.Fprint(c.Logger, format)
		return
	}
	fmt.Fprintf(c.Logger, format, args...)
}

func (c *Client) getHTTPClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	httpClient := *http.DefaultClient
	httpClient.Timeout = time.Second * 30
	return &httpClient
}

func (c *Client) rateLimitRequest() {
	if !c.lastRequestTime.IsZero() {
		minDelay := c.MinRequestDelay
		if minDelay == 0 {
			minDelay = time.Millisecond * 360
		}
		since := time.Since(c.lastRequestTime)
		if minDelay > since {
			time.Sleep(minDelay - since)
		}
	}
	c.lastRequestTime = time.Now()
}

func (c *Client) doPost(uri string, body []byte) ([]byte, error) {
	if c.httpPostOverride != nil {
		return c.httpPostOverride(uri, body)
	}
	return c.doPostInternal(uri, body)
}

func (c *Client) doPostInternal(uri string, body []byte) ([]byte, error) {
	c.rateLimitRequest()

	// try to back-off exponentially
	// note: backing off doesn't seem to work i.e. I get 429 from subsequent requests as well
	nRepeats := 0
	timeouts := []time.Duration{time.Second * 3, time.Second * 5, time.Second * 10}
repeatRequest:
	br := bytes.NewBuffer(body)
	req, err := http.NewRequest("POST", uri, br)
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
		c.logf("httpClient.Do() failed with %s\n", err)
		return nil, err
	}

	if rsp.StatusCode == http.StatusTooManyRequests {
		if nRepeats < 3 {
			closeNoError(rsp.Body)
			c.logf("retrying '%s' because httpClient.Do() returned %d (%s)\n", uri, rsp.StatusCode, rsp.Status)
			time.Sleep(timeouts[nRepeats])
			nRepeats++
			goto repeatRequest
		}
	}

	defer closeNoError(rsp.Body)

	if rsp.StatusCode != 200 {
		d, _ := ioutil.ReadAll(rsp.Body)
		c.logf("Error: status code %s\nBody:\n%s\n", rsp.Status, PrettyPrintJS(d))
		return nil, fmt.Errorf("http.Post('%s') returned non-200 status code of %d", uri, rsp.StatusCode)
	}
	d, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		c.logf("Error: ioutil.ReadAll() failed with %s\n", err)
		return nil, err
	}
	return d, nil
}

func (c *Client) doNotionAPI(apiURL string, requestData interface{}, result interface{}) (map[string]interface{}, error) {
	var body []byte
	var err error
	if requestData != nil {
		body, err = jsonit.MarshalIndent(requestData, "", "  ")
		if err != nil {
			return nil, err
		}
	}
	uri := notionHost + apiURL
	c.logf("POST %s\n", uri)
	if len(body) > 0 {
		logJSON(c, body)
	}

	d, err := c.doPost(uri, body)
	if err != nil {
		return nil, err
	}
	logJSON(c, d)

	err = jsonit.Unmarshal(d, result)
	if err != nil {
		c.logf("Error: json.Unmarshal() failed with %s\n. Body:\n%s\n", err, string(d))
		return nil, err
	}
	var m map[string]interface{}
	err = jsonit.Unmarshal(d, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ExtractNoDashIDFromNotionURL tries to extract notion page id from
// notion URL, e.g. given:
// https://www.notion.so/Advanced-web-spidering-with-Puppeteer-ea07db1b9bff415ab180b0525f3898f6
// returns ea07db1b9bff415ab180b0525f3898f6
// returns "" if didn't detect valid notion id in the url
func ExtractNoDashIDFromNotionURL(uri string) string {
	maybeID := ToNoDashID(uri)
	if maybeID != "" {
		return maybeID
	}
	id := uri
	// only look at the last part of the url
	parts := strings.Split(id, "/")
	id = parts[len(parts)-1]
	// look at last '-' part
	parts = strings.Split(id, "-")
	id = parts[len(parts)-1]
	// remove url fragment
	parts = strings.Split(id, "#")
	id = parts[0]
	return ToNoDashID(id)
}

func (p *Page) findInlinePageReferences(block *Block) []string {
	// TODO: maybe note which blocks were already processed
	// to avoid checking things multiple times
	parseTitle(block)
	if len(block.InlineContent) == 0 {
		return nil
	}
	var res []string

	for _, ts := range block.InlineContent {
		for _, attr := range ts.Attrs {
			switch AttrGetType(attr) {
			case AttrPage:
				pageID := AttrGetPageID(attr)
				res = append(res, pageID)
			}
		}
	}
	return res
}

// find referenced blocks that we don't have yet
func (p *Page) findMissingBlocks() []string {
	missing := map[string]struct{}{}
	for _, block := range p.idToBlock {
		if !block.Alive {
			continue
		}
		// don't want to recursively pull information about sub-pages
		// or linked pages
		if block.Type == BlockPage {
			continue
		}

		for _, id := range block.ContentIDs {
			if _, ok := p.idToBlock[id]; !ok {
				missing[id] = struct{}{}
			}
		}
		referencedPages := p.findInlinePageReferences(block)
		for _, id := range referencedPages {
			if _, ok := p.idToBlock[id]; !ok {
				missing[id] = struct{}{}
			}
		}
	}

	var res []string
	for id := range missing {
		if _, ok := p.blocksToSkip[id]; ok {
			continue
		}
		res = append(res, id)
	}
	sort.Strings(res)
	return res
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
		idToComment:        map[string]*Comment{},
		idToDiscussion:     map[string]*Discussion{},
		idToUser:           map[string]*User{},
		idToSpace:          map[string]*Space{},
		blocksToSkip:       map[string]struct{}{},
	}

	var root *Block
	// get page's root block and then recursively download referenced blocks
	{
		recVals, err := c.GetBlockRecords([]string{pageID})
		if err != nil {
			return nil, err
		}
		res := recVals.Results[0]
		// this might happen e.g. when a page is not publicly visible
		root = res.Block
		if root == nil {
			return nil, newErrPageNotFound(pageID)
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
		recordMap := rsp.RecordMap
		for id, v := range recordMap.Blocks {
			b := v.Block
			if b.Alive {
				p.idToBlock[id] = b
			} else {
				p.blocksToSkip[id] = struct{}{}
			}
		}
		for id, r := range recordMap.Collections {
			p.CollectionRecords = append(p.CollectionRecords, r)
			p.idToCollection[id] = r.Collection
		}
		for id, r := range recordMap.CollectionViews {
			p.CollectionViewRecords = append(p.CollectionViewRecords, r)
			p.idToCollectionView[id] = r.CollectionView
		}
		for id, r := range recordMap.Discussions {
			p.DiscussionRecords = append(p.DiscussionRecords, r)
			p.idToDiscussion[id] = r.Discussion
		}
		for id, r := range recordMap.Comments {
			p.CommentRecords = append(p.CommentRecords, r)
			p.idToComment[id] = r.Comment
		}
		for id, r := range recordMap.Spaces {
			p.SpaceRecords = append(p.SpaceRecords, r)
			p.idToSpace[id] = r.Space
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
		missing := p.findMissingBlocks()
		if len(missing) == 0 {
			break
		}
		c.vlogf("DownloadPage: %d missing blocks in iteration %d\n", len(missing), missingIter)
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

			recVals, err := c.GetBlockRecords(toGet)
			if err != nil {
				return nil, err
			}
			for n, recordValue := range recVals.Results {
				block := recordValue.Block
				// This can happen e.g. in 157765353f2c4705bd45474e5ba8b46c
				// Server returns { "role": "none" },
				expectedID := toGet[n]
				if block != nil {
					// Happens when block is from a relation column thus it's view might be in a different page
					viewInsideOfPage := len(block.ViewIDs) == 0 // Don't consider blocks without a view
					for _, collectionViewID := range block.ViewIDs {
						_, ok := p.idToCollectionView[collectionViewID]
						if ok {
							viewInsideOfPage = true
						} else {
							c.vlogf("collection view id = %s block id = %s is outside of page.\n", collectionViewID, block.ID)
						}
					}
					if viewInsideOfPage {
						p.idToBlock[block.ID] = block
					} else {
						p.blocksToSkip[expectedID] = struct{}{}
					}

					continue
				}
				p.blocksToSkip[expectedID] = struct{}{}
				if n > 0 {
					prevRecord := recVals.Results[n-1]
					if prevRecord == nil || prevRecord.Block == nil {
						// this can happen if we don't have access to this page
						c.vlogf("prevBlock.Value is nil at position n = %d with expected id %s.\n", n, expectedID)
					} else {
						prevBlockID := prevRecord.Block.ID
						c.vlogf("block is nil at position n = %d with expected id %s. Prev block id: %s\n", n, expectedID, prevBlockID)
					}
				} else {
					c.vlogf("block is nil at position n = %v with expected id %s.\n", n, expectedID)
				}
			}
		}
	}

	err := p.resolveBlocks()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve blocks on page '%s': %s", p.ID, err)
	}

	/*
		TODO: use loadUserContent to get info about users
			for id, r := range recordMap.Users {
				p.UserRecords = append(p.UserRecords, r)
				p.idToUser[id] = r.User
			}
	*/

	blockIDs := getBlockIDsSorted(p.idToBlock)
	for _, id := range blockIDs {
		block := p.idToBlock[id]
		if block.Type != BlockCollectionView && block.Type != BlockCollectionViewPage {
			continue
		}
		if len(block.ViewIDs) == 0 {
			return nil, fmt.Errorf("collection_view has no ViewIDs")
		}

		// TODO: should fish out the user based on block.CreatedBy
		// TODO: notion changed the api and User is no long returned in loadPageChunk
		// need to use syncRecordValues
		if false && len(p.UserRecords) == 0 {
			return nil, fmt.Errorf("no users when trying to resolve collection_view")
		}

		collectionID := block.CollectionID
		for _, collectionViewID := range block.ViewIDs {
			collectionView, ok := p.idToCollectionView[collectionViewID]
			if !ok {
				return nil, fmt.Errorf("didn't find collection_view with id '%s'", collectionViewID)
			}
			collection, ok := p.idToCollection[collectionID]
			if !ok {
				//return nil, fmt.Errorf("Didn't find collection with id '%s'", collectionID)
				continue
			}
			spaceID := block.SpaceID
			req := QueryCollectionRequest{}
			req.Collection.ID = collectionID
			req.Collection.SpaceID = spaceID
			req.CollectionView.ID = collectionViewID
			req.CollectionView.SpaceID = spaceID
			res, err := c.QueryCollection(req, collectionView.Query)
			if err != nil {
				return nil, err
			}

			tableView := &TableView{
				Page:           p,
				CollectionView: collectionView,
				Collection:     collection,
			}
			if err := c.buildTableView(tableView, res); err != nil {
				return nil, err
			}
			block.TableViews = append(block.TableViews, tableView)
			p.TableViews = append(p.TableViews, tableView)
		}
	}

	for _, b := range p.idToBlock {
		err := parseProperties(b)
		if err != nil {
			return nil, fmt.Errorf("failed to parse properties of block '%s', err: '%s'", b.ID, err)
		}
		b.Page = p

		switch b.ParentTable {
		case TableSpace:
			// TODO: Support parent table space
			continue
		case TableCollection:
			// TODO: Support parent table collection
			continue
		case TableBlock:
			// Page's parent is outside of this page
			if isPageBlock(b) && !p.IsSubPage(b) {
				continue
			}

			b.Parent = p.BlockByID(b.GetParentNotionID())
			if b.Parent == nil {
				return nil, fmt.Errorf("could not find parent '%s' of id '%s' of block '%s'", b.ParentTable, b.ParentID, b.ID)
			}
		default:
			c.vlogf("unsupported parent table type %s of block %s", b.ParentTable, b.ID)
		}
	}

	return p, nil
}
