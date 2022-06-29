package main

import (
	"encoding/json"
	"fmt"

	"github.com/kjk/notionapi"
)

// https://www.notion.so/Comparing-prices-of-VPS-servers-c30393989ae549c3a39f21ca5a681d72
func testGetBlockRecords() {
	c := newClient()
	ids := []string{"c30393989ae549c3a39f21ca5a681d72"}
	blocks, err := c.GetBlockRecords(ids)
	must(err)
	panicIf(len(blocks) == 1)
	dashID := notionapi.ToDashID(ids[0])
	panicIf(blocks[0].ID != dashID)
	for _, r := range blocks {
		logf("testSyncRecordValues: id: '%s'\n", r.ID)
	}
}

// https://www.notion.so/Test-text-4c6a54c68b3e4ea2af9cfaabcc88d58d
func testLoadCachePageChunk() {
	c := newClient()
	pageID := notionapi.ToDashID("4c6a54c68b3e4ea2af9cfaabcc88d58d")
	rsp, err := c.LoadCachedPageChunk(pageID, 0, nil)
	must(err)
	fmt.Printf("rsp:\n%#v\n\n", rsp)
	for blockID, block := range rsp.RecordMap.Blocks {
		fmt.Printf("blockID: %s, block.ID: %s\n", blockID, block.ID)
		panicIf(blockID != block.ID)
	}
}

func testQueryDecode() {
	s := `{
    "aggregate": [
      {
        "aggregation_type": "count",
        "id": "count",
        "property": "title",
        "type": "title",
        "view_type": "table"
      }
    ]
  }`
	var v notionapi.Query
	err := json.Unmarshal([]byte(s), &v)
	must(err)
}

func testSubPages() {
	// test that GetSubPages() only returns direct children
	// of a page, not link to pages
	client := newClient()
	uri := "https://www.notion.so/Test-sub-pages-in-mono-font-381243f4ba4d4670ac491a3da87b8994"
	pageID := "381243f4ba4d4670ac491a3da87b8994"
	page, err := client.DownloadPage(pageID)
	must(err)
	subPages := page.GetSubPages()
	nExp := 7
	panicIf(len(subPages) != nExp, "expected %d sub-pages of '%s', got %d", nExp, uri, len(subPages))
	logf("ok\ttestSubPages()\n")
}

// TODO: this fails now
func testQueryCollection() {
	// test for table on https://www.notion.so/Comparing-prices-of-VPS-servers-c30393989ae549c3a39f21ca5a681d72
	c := newClient()
	spaceID := "bc202e06-6caa-4e3f-81eb-f226ab5deef7"
	collectionID := "0567b270-3cb1-44e4-847c-34a843f55dfc"
	collectionViewID := "74e9cd84-ff2d-4259-bd56-5f8478da8839"
	req := notionapi.QueryCollectionRequest{}
	req.Collection.ID = collectionID
	req.Collection.SpaceID = spaceID
	req.CollectionView.ID = collectionViewID
	req.CollectionView.SpaceID = spaceID
	sort := notionapi.QuerySort{
		ID:        "6e89c507-e0da-47c7-b8c8-fe2b336e0985",
		Type:      "number",
		Property:  "E13y",
		Direction: "ascending",
	}
	q := notionapi.Query{
		Sort: []notionapi.QuerySort{sort},
	}
	res, err := c.QueryCollection(req, &q)
	must(err)
	colRes := res.Result.ReducerResults.CollectionGroupResults
	panicIf(colRes.Total != 18, "colRes.Total == %d", colRes.Total)
	panicIf(len(colRes.BlockIds) != 18)
	panicIf(colRes.Type != "results")
	//fmt.Printf("%#v\n", colRes)
}

// sanity tests are basic tests to validate changes
// meant to not take too long
func sanityTests() {
	logf("Running sanity tests\n")
	testQueryDecode()

	runGoTests()
	testGetBlockRecords()
	testLoadCachePageChunk()
	testSubPages()

	// TODO: something must have changed on the server and this test fails now
	// testQueryCollection()

	// queryCollectionApi changed
	pageID := "c30393989ae549c3a39f21ca5a681d72"
	testCachingDownloads(pageID)
	logf("ok\ttestCachingDownloads() of %s ok!\n", pageID)
	// TODO: more tests?
}
