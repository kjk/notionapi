package main

import (
	"fmt"
	"path/filepath"
)

// https://www.notion.so/Notion-Pok-dex-d6eb49cfc68f402881af3aef391443e6
func testToHTML3() {
	// to speed up iteration, we skip pages that we know we render correctly
	validBad := []string{}

	zipPath := filepath.Join(topDir(), "data", "testdata", "Export-html-3b2938f6-675b-4107-8fbf-e9505478a292.zip")
	zipFiles := readZipFile(zipPath)
	fmt.Printf("There are %d files in zip file\n", len(zipFiles))

	startPage := "d6eb49cfc68f402881af3aef391443e6"
	//startPage = ""
	testToHTMLRecur(startPage, "", validBad, zipFiles)
}
