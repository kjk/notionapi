package main

import (
	"fmt"
	"path/filepath"
)

// https://www.notion.so/Blendle-s-Employee-Handbook-3b617da409454a52bc3a920ba8832bf7
func testToHTML1() {
	// to speed up iteration, we skip pages that we know we render correctly
	validBad := []string{
		// Notion is missing one link to page
		"13aa42a5a95d4357aa830c3e7ff35ae1",
		// TODO(1): Notion renders <div> inside <p> which is illegal and makes pretty-printing
		// not work, so can't compare results. Probably because they render children
		//  <p> inside </p> instead of after </p>
		"4f5ee5cf485048468db8dfbf5924409c",
		// Notion is missing one link to page
		"7a5df17b32e84686ae33bf01fa367da9",
		// Notion is malformed
		"7afdcc4fbede49bc9582469ad6e86fd3",
		// Notion is malformed
		"949f33cdba814fc4a288d81c6e7c810d",
		// Notion is missing one link to page
		"b1b31f6d3405466c988676f996ce03ad",
		// Notion is missong some link to page
		"ef850413bb53491eaebccaa09eeb8630",
		// Notion is malformed
		"f2d97c9cba804583838acf5d571313f5",
		// Notion is malformed
		"3c892714f4dc4d2194619fdccba48fc6",
		// Different ids
		"8f12cc5182a6437aac4dc518cb28b681",
	}
	startWith := "8f12cc5182a6437aac4dc518cb28b681"

	zipPath := filepath.Join(topDir(), "data", "testdata", "Export-html-6f6dae04-a337-419e-81ca-f82de3202b9e.zip")
	zipFiles := readZipFile(zipPath)
	fmt.Printf("There are %d files in zip file\n", len(zipFiles))

	startPage := "3b617da409454a52bc3a920ba8832bf7"
	//startPage = "13aa42a5a95d4357aa830c3e7ff35ae1"
	testToHTMLRecur(startPage, startWith, validBad, zipFiles)
}
