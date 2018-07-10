package notion

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

var (
	// Logger is used to log requests and responses for debugging.
	// By default is not set.
	Logger io.Writer
)

// PageInfo describes a single Notion page
type PageInfo struct {
	ID   string
	Page *BlockValue
}

// pretty-print if valid JSON. If not, return unchanged
func ppJSON(js []byte) []byte {
	var m map[string]interface{}
	err := json.Unmarshal(js, &m)
	if err != nil {
		return js
	}
	d, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return js
	}
	return d
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
	if Logger != nil {
		_, _ = fmt.Fprintf(Logger, "POST %s\n", uri)
		if len(js) > 0 {
			Logger.Write(js)
			io.WriteString(Logger, "\n")
		}
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
		if Logger != nil {
			fmt.Fprintf(Logger, "Error: failed with %s\n", err)
		}
		return err
	}
	if rsp.StatusCode != 200 {
		if Logger != nil {
			fmt.Fprintf(Logger, "Error: status code %d\n", rsp.StatusCode)
		}
		return fmt.Errorf("http.Post('%s') returned non-200 status code of %d", uri, rsp.StatusCode)
	}
	defer rsp.Body.Close()
	d, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		if Logger != nil {
			fmt.Fprintf(Logger, "Error: ioutil.ReadAll() failed with %s\n", err)
		}
		return err
	}
	if Logger != nil {
		Logger.Write(ppJSON(d))
		io.WriteString(Logger, "\n")
	}
	err = parseFn(d)
	if err != nil {
		if Logger != nil {
			fmt.Fprintf(Logger, "Error: json.Unmarshal() failed with %s\n", err)
		}
	}
	return err
}

func apiLoadPageChunk(pageID string, cur *cursor) (*loadPageChunkResponse, error) {
	apiURL := "/api/v3/loadPageChunk"
	if cur == nil {
		cur = &cursor{
			// to mimic browser api which sends empty array for this argment
			Stack: make([][]stack, 0),
		}
	}
	req := &loadPageChunkRequest{
		PageID:          pageID,
		Limit:           50,
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

func resolveBlocks(block *BlockValue, blockMap map[string]*BlockInfo) error {
	/*
		if block.isResolved {
			return nil
		}
		block.isResolved = true
	*/

	if len(block.Blocks) == 0 && IsTypeWithBlocks(block.Type) {
		if len(block.Properties) == 0 {
			return fmt.Errorf("block with type '%s' has no Properties", block.Type)
		}
		title, ok := block.Properties["title"]
		if !ok {
			return fmt.Errorf("block with type '%s' is missing Properties['title'] property. Properties: %#v", block.Type, block.Properties)
		}
		var err error
		block.Blocks, err = parseInlineBlocks(title)
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
	block.Content = make([]*BlockValue, n, n)
	for i, id := range block.ContentIDs {
		resolved := blockMap[id]
		if resolved == nil {
			return fmt.Errorf("Couldn't resolve block with id '%s'", id)
		}
		block.Content[i] = resolved.Value
		resolveBlocks(resolved.Value, blockMap)
	}
	return nil
}

// GetPageInfo returns Noion page data given its id
func GetPageInfo(pageID string) (*PageInfo, error) {
	// TODO: validate pageID
	req := &getRecordValuesRequest{
		Requests: []getRecordValuesRequestInner{
			{
				Table: TableBlock,
				ID:    pageID,
			},
		},
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
	var res PageInfo
	//fmt.Printf("%#v\n", rsp)
	// change to cannonical version of page id
	pageID = rsp.Results[0].Value.ID
	res.ID = pageID
	res.Page = rsp.Results[0].Value
	rsp2, err := apiLoadPageChunk(pageID, nil)
	if err != nil {
		return nil, err
	}
	// TODO: handle limit
	err = resolveBlocks(res.Page, rsp2.RecordMap.Blocks)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
