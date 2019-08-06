package main

import (
	"fmt"
	"path/filepath"
)

// https://www.notion.so/Test-pages-for-notionapi-0367c2db381a4f8b9ce360f388a6b2e3
func testToHTML2() {
	// to speed up iteration, we skip pages that we know we render correctly
	/*
		var toSkip = []string{
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
			"cec279f5f21748aa80d8976670d402ce",
			"de7a4da288664a6ab81121afc1a5447c",
			"e0cffcb524324661a152d9427ef5c842",
			// TODO: don't know the name of the page when rendering inline page
			"f97ffca91f8949b48004999df34ab1f7",
			"fd9338a719a24f02993fcfbcf3d00bb0",
			"fd9780323287488794ebff8956e0e7fc",
			"5ce4c6a8870e4a629983316c32568f41",
			"7bb4f0dea9024ce5a1b6d7a88ce5a024",
		}
	*/
	validBad := []string{
		// Notion missing link to page
		"86b5223576104fa69dc03675e44571b7",
	}
	startWith := "86b5223576104fa69dc03675e44571b7"

	zipPath := filepath.Join(topDir(), "data", "testdata", "Export-html-884455e1-98f4-4c77-8733-5373a4b47b85.zip")

	zipFiles := readZipFile(zipPath)
	fmt.Printf("There are %d files in zip file\n", len(zipFiles))

	startPage := "92dd7aedf1bb4121aaa8986735df3d13"
	//startPage = ""
	testToHTMLRecur(startPage, startWith, validBad, zipFiles)
}
