package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kjk/notionapi"
)

// https://www.notion.so/Test-pages-for-notionapi-0367c2db381a4f8b9ce360f388a6b2e3
func testToHTML2() {
	validBad := []string{
		// TODO: Notion doen't export link to page
		"86b5223576104fa69dc03675e44571b7",
		// TODO: a date with time zone not formatted correctly
		"97100f9c17324fd7ba3d3c5f1832104d",
		// TODO: bad indent in toc
		"c969c9455d7c4dd79c7f860f3ace6429",
		// TODO: Notion exports a column "Title" marked as "not visible"
		"92dd7aedf1bb4121aaa8986735df3d13",
		// TODO: don't have name of the page
		"f97ffca91f8949b48004999df34ab1f7",
	}

	// top-level page for blendle handbok
	startPage := "0367c2db381a4f8b9ce360f388a6b2e3"
	os.MkdirAll(cacheDir, 0755)
	name := startPage + "-html.zip"
	zipPath := filepath.Join("data", name)
	if _, err := os.Stat(zipPath); err != nil {
		fmt.Printf("Downloading %s\n", zipPath)
		must(exportPageToFile(startPage, notionapi.ExportTypeHTML, true, zipPath))
	}

	zipFiles := readZipFile(zipPath)
	fmt.Printf("There are %d files in zip file\n", len(zipFiles))

	testToHTMLRecur(startPage, validBad, zipFiles)
}
