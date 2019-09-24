package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kjk/notionapi"
)

var (
	smokeDir  string
	smokeSeen map[string]bool
)

// load the page, render to md and html. repeat for all sub-children
func loadAndRenderPageRecur(pageID string) {
	id := notionapi.ToNoDashID(pageID)
	if smokeSeen[id] {
		return
	}
	smokeSeen[id] = true
	page := toHTML(pageID)
	_, md := toMarkdown(page)
	mdName := fmt.Sprintf("%s.page.md", id)
	mdPath := filepath.Join(cacheDir, mdName)
	writeFile(mdPath, md)
	logf("%s : md version of the page\n", mdPath)
	for _, pageID := range page.GetSubPages() {
		loadAndRenderPageRecur(pageID)
	}
}

// smoke test is meant to be run after non-trivial changes
// it tries to exercise as many features as possible while still
// being reasonably fast
func smokeTest() {
	smokeDir = filepath.Join("data", "smoke")
	recreateDir(smokeDir)
	// over-write cacheDir
	defer func(curr string) {
		cacheDir = curr
	}(cacheDir)

	// over-write cache dir location
	cacheDir = filepath.Join(smokeDir, "cache")
	err := os.MkdirAll(cacheDir, 0755)
	must(err)

	logFilePath := filepath.Join(smokeDir, "log.txt")
	logf("Running smokeTest(), log file: '%s', cache dir: '%s'\n", logFilePath, cacheDir)
	f, err := os.Create(logFilePath)
	must(err)
	defer f.Close()
	logFile = f

	// https://www.notion.so/Test-pages-for-notionapi-0367c2db381a4f8b9ce360f388a6b2e3
	// root page of my test pages
	smokeSeen = map[string]bool{}
	flgNoOpen = true
	loadAndRenderPageRecur("0367c2db381a4f8b9ce360f388a6b2e3")
}