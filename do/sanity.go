package main

import (
	"encoding/json"

	"github.com/kjk/notionapi"
)

// https://www.notion.so/Comparing-prices-of-VPS-servers-c30393989ae549c3a39f21ca5a681d72
func testSyncRecordValues() {
	c := newClient()
	ids := []string{"c30393989ae549c3a39f21ca5a681d72"}
	res, err := c.SyncBlockRecords(ids)
	must(err)
	for table, records := range res.RecordMap {
		panicIf(table != "block")
		for id, r := range records {
			logf("testSyncRecordValues: id: %s, id: '%s'\n", id, r.ID)
			panicIf(id != r.ID)
		}
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
	panicIf(colRes.Total != 18)
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
	testSyncRecordValues()
	testSubPages()

	// TODO: something must have changed on the server and this test fails now
	// testQueryCollection()

	// queryCollectionApi changed
	pageID := "c30393989ae549c3a39f21ca5a681d72"
	testCachingDownloads(pageID)
	logf("ok\ttestCachingDownloads() of %s ok!\n", pageID)
	// TODO: more tests?
}
