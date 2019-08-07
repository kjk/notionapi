package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kjk/notionapi"
)

// https://www.notion.so/Notion-Pok-dex-d6eb49cfc68f402881af3aef391443e6
func testToHTML3() {
	// to speed up iteration, we skip pages that we know we render correctly
	validBad := []string{
		// TODO: I'm not formatting table correctly
		"00f68316d03c4830b00c453e542a1df7",
		// TODO: I'm not formatting table correctly
		"02bfca37eae5484ba942a00c99076b7a",
		// TODO: I'm not formatting table correctly
		"09e9c8f5c9df445f94d1cf3f39a1039f",
		// TODO: totally different export
		"0e684b2e45ea434293274c802b5ad702",
		// TODO: I'm not exporting a table the right way
		"141c2ef1718b471896c915ae622dae83",
		// TODO: Bad export
		"14d22d99fb074352a59d78751646cf3d",
	}

	startWith := ""

	// top-level page for Notion Pok Dex
	startPage := "d6eb49cfc68f402881af3aef391443e6"
	os.MkdirAll(cacheDir, 0755)
	name := startPage + "-html.zip"
	zipPath := filepath.Join("data", name)
	if _, err := os.Stat(zipPath); err != nil {
		fmt.Printf("Downloading %s\n", zipPath)
		must(exportPageToFile(startPage, notionapi.ExportTypeHTML, true, zipPath))
	}

	zipFiles := readZipFile(zipPath)
	fmt.Printf("There are %d files in zip file\n", len(zipFiles))

	testToHTMLRecur(startPage, startWith, validBad, zipFiles)
}
