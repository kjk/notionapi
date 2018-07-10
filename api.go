package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
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

// PageInfo describes a single Notion page
type PageInfo struct {
	ID   string
	Page *Block
}

func doNotionAPI(apiURL string, requestData interface{}, parseFn func(d []byte) error) error {
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
	log("POST %s\n", uri)
	if len(js) > 0 {
		log("%s\n", string(js))
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", acceptLang)

	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		log("Error: failed with %s\n", err)
		return err
	}
	if rsp.StatusCode != 200 {
		log("Error: status code %d\n", rsp.StatusCode)
		return fmt.Errorf("http.Post('%s') returned non-200 status code of %d", uri, rsp.StatusCode)
	}
	defer rsp.Body.Close()
	d, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log("Error: ioutil.ReadAll() failed with %s\n", err)
		return err
	}
	logJSON(d)
	err = parseFn(d)
	if err != nil {
		log("Error: json.Unmarshal() failed with %s\n", err)
	}
	return err
}

func apiLoadPageChunk(pageID string, cur *cursor) (*loadPageChunkResponse, error) {
	// emulating notion's website api usage: 50 items on first request,
	// 30 on subsequent requests
	limit := 30
	apiURL := "/api/v3/loadPageChunk"
	if cur == nil {
		cur = &cursor{
			// to mimic browser api which sends empty array for this argment
			Stack: make([][]stack, 0),
		}
		limit = 50
	}
	req := &loadPageChunkRequest{
		PageID:          pageID,
		Limit:           limit,
		Cursor:          *cur,
		VerticalColumns: false,
	}
	var rsp *loadPageChunkResponse
	parse := func(d []byte) error {
		var err error
		rsp, err = parseLoadPageChunk(d)
		return err
	}
	err := doNotionAPI(apiURL, req, parse)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func resolveBlocks(block *Block, blockMap map[string]*Block) error {
	/*
		if block.isResolved {
			return nil
		}
		block.isResolved = true
	*/

	if len(block.InlineContent) == 0 && IsTypeWithBlocks(block.Type) {
		if len(block.Properties) == 0 {
			return fmt.Errorf("block with type '%s' has no Properties", block.Type)
		}
		title, ok := block.Properties["title"]
		if !ok {
			return fmt.Errorf("block with type '%s' is missing Properties['title'] property. Properties: %#v", block.Type, block.Properties)
		}
		var err error
		block.InlineContent, err = parseInlineBlocks(title)
		if err != nil {
			return err
		}
	}

	if TypeTodo == block.Type {
		if v, ok := block.Properties["checked"]; ok {
			if s, ok := v.(string); ok {
				block.IsChecked = strings.EqualFold(s, "Yes")
			}
		}
	}

	if block.Content != nil || len(block.ContentIDs) == 0 {
		return nil
	}
	n := len(block.ContentIDs)
	block.Content = make([]*Block, n, n)
	for i, id := range block.ContentIDs {
		resolved := blockMap[id]
		if resolved == nil {
			return fmt.Errorf("Couldn't resolve block with id '%s'", id)
		}
		block.Content[i] = resolved
		resolveBlocks(resolved, blockMap)
	}
	return nil
}

func apiGetRecordValues(ids []string) (*getRecordValuesResponse, error) {
	req := &getRecordValuesRequest{}

	for _, id := range ids {
		v := getRecordValuesRequestInner{
			Table: TableBlock,
			ID:    id,
		}
		req.Requests = append(req.Requests, v)
	}

	apiURL := "/api/v3/getRecordValues"
	var rsp *getRecordValuesResponse
	parse1 := func(d []byte) error {
		var err error
		rsp, err = parseGetRecordValues(d)
		return err
	}
	err := doNotionAPI(apiURL, req, parse1)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func findMissingBlocks(startIDS []string, idToBlock map[string]*Block) ([]string, error) {
	return []string{}, nil
}

// GetPageInfo returns Noion page data given its id
func GetPageInfo(pageID string) (*PageInfo, error) {
	// TODO: validate pageID
	var res PageInfo
	//fmt.Printf("%#v\n", rsp)
	// change to cannonical version of page id

	{
		recVals, err := apiGetRecordValues([]string{pageID})
		if err != nil {
			return nil, err
		}
		pageID = recVals.Results[0].Value.ID
		res.ID = pageID
		res.Page = recVals.Results[0].Value
	}

	var idToBlock map[string]*Block
	for {
		rsp, err := apiLoadPageChunk(pageID, nil)
		if err != nil {
			return nil, err
		}
		for id, blockWithRole := range rsp.RecordMap.Blocks {
			idToBlock[id] = blockWithRole.Value
		}
		// TODO: handle stack
		break
	}

	// get blocks that are not already loaded, 30 per request
	missing, err := findMissingBlocks(res.Page.ContentIDs, idToBlock)
	if err != nil {
		return nil, err
	}
	dbg("There are %d missing blocks\n", len(missing))
	for len(missing) > 0 {
	}

	err = resolveBlocks(res.Page, idToBlock)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
