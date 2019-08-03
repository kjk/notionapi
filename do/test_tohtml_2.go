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
	}

	zipPath := filepath.Join(topDir(), "data", "testdata", "Export-html-a352c43e-0545-481d-a935-57d4a3330bca.zip")
	zipFiles := readZipFile(zipPath)
	fmt.Printf("There are %d files in zip file\n", len(zipFiles))

	startPage := "0367c2db381a4f8b9ce360f388a6b2e3"
	//startPage = ""
	testToHTMLRecur(startPage, toSkip, zipFiles)
}
