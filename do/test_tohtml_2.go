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
	}

	zipPath := filepath.Join(topDir(), "data", "testdata", "Export-html-a352c43e-0545-481d-a935-57d4a3330bca.zip")
	zipFiles := readZipFile(zipPath)
	fmt.Printf("There are %d files in zip file\n", len(zipFiles))

	startPage := "0367c2db381a4f8b9ce360f388a6b2e3"
	//startPage = ""
	testToHTMLRecur(startPage, toSkip, zipFiles)
}
