package main

import (
	"github.com/kjk/notionapi"
)

// https://www.notion.so/Comparing-prices-of-VPS-servers-c30393989ae549c3a39f21ca5a681d72
func testSyncRecordValues() {
	client := &notionapi.Client{}
	//client.DebugLog = true
	//client.Logger = os.Stdout
	ids := []string{"c30393989ae549c3a39f21ca5a681d72"}
	res, err := client.SyncBlockRecords(ids)
	must(err)
	for table, records := range res.RecordMap {
		panicIf(table != "block")
		for id, r := range records {
			logf("testSyncRecordValues: id: %s, id: '%s'\n", id, r.ID)
			panicIf(id != r.ID)
		}
	}
}

// sanity tests are basic tests to validate changes
// meant to not take too long
func sanityTests() {
	logf("Running sanity tests\n")
	testSyncRecordValues()

	if true {
		runGoTests()
		testSubPages()
	}
	if false {
		// queryCollectionApi changed
		pageID := "c30393989ae549c3a39f21ca5a681d72"
		testCachingDownloads(pageID)
		logf("ok\ttestCachingDownloads() of %s ok!\n", pageID)
	}
	// TODO: more tests?
}
