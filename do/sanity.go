package main

// sanity tests are basic tests to validate changes
// meant to not take too long
func sanityTests() {
	logf("Running sanity tests\n")
	runGoTests()
	testSubPages()
	logf("ok\ttestSubPages()\n")
	pageID := "dd5c0a813dfe4487a6cd432f82c0c2fc"
	testCachingDownloads(pageID)
	logf("ok\testCachingDownloads() of %s ok!\n", pageID)
	// TODO: more tests?
}
