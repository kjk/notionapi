package main

// sanity tests are basic tests to validate changes
// meant to not take too long
func sanityTests() {
	logf("Running sanity tests\n")
	runGoTests()
	testSubPages()
	logf("ok\ttestSubPages()\n")
	pageID := "c30393989ae549c3a39f21ca5a681d72"
	testCachingDownloads(pageID)
	logf("ok\ttestCachingDownloads() of %s ok!\n", pageID)
	// TODO: more tests?
}
