package main

import (
	"fmt"
	"path/filepath"
)

// https://www.notion.so/Test-pages-for-notionapi-0367c2db381a4f8b9ce360f388a6b2e3
func testToHTML2() {
	// to speed up iteration, we skip pages that we know we render correctly
	var toSkip = []string{
		"0367c2db381a4f8b9ce360f388a6b2e3",
		"0fa8d15a16134f0c9fad1aa0a7232374",
		"1f062365d2f34c19b6c69b1f250ff4b4",
		"24341592002b40159933e6bc3f31f359",
		"25b6ac21d67744f18a4dc071b21a86fe",
		"484919a1647144c29234447ce408ff6b",
		"4c6a54c68b3e4ea2af9cfaabcc88d58d",
		"5007504c406840dcbc3353364ede3d02",
		"6682351e44bb4f9ca0e149b703265bdb",
		"70ecbf1f5abc41d48a4e4320aeb38d10",
		"72fd504c58984cc5a5dfb86b6f8617dc",
		"7b9cdf3ab2cf405692e9810b0ac8322e",
		"7e825831be07487e87e756e52914233b",
		"8511412cbfde432ba226648e9bdfbec2",
		// TODO: Notion missing link to page
		"86b5223576104fa69dc03675e44571b7",
		// TODO: Notion exports a column "Title" marked as "not visible"
		"92dd7aedf1bb4121aaa8986735df3d13",
		"94167af6567043279811dc923edd1f04",
		// TODO: a single date time value
		"97100f9c17324fd7ba3d3c5f1832104d",
		"99031183f223417988241fdad218ceba",
		"b0a87a5a9c304534bf85c40f6aa29176",
		"c4abe71e76084cc78249502c60f3ff59",
		// TODO: Notion missing link to page
		"c969c9455d7c4dd79c7f860f3ace6429",
	}

	zipPath := filepath.Join(topDir(), "data", "testdata", "Export-html-a352c43e-0545-481d-a935-57d4a3330bca.zip")
	zipFiles := readZipFile(zipPath)
	fmt.Printf("There are %d files in zip file\n", len(zipFiles))

	startPage := "0367c2db381a4f8b9ce360f388a6b2e3"
	//startPage = ""
	testToHTMLRecur(startPage, toSkip, zipFiles)
}
