package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/kjk/notionapi"
)

var (
	smokeDir  string
	smokeSeen map[string]bool
)

var (
	collectionSchemaTypes = map[string]bool{}
	collectionSchemeProps = map[string]bool{}
	formatPropsPerBlock   = map[string]map[string]bool{}
)

// TODO: doesn't seem to get format for all blocks
// Is ForEachBlock not traversing every block?
func recordFormatProps(block *notionapi.Block) {
	format := block.RawJSON["format"]
	if format == nil {
		return
	}
	fm, ok := format.(map[string]interface{})
	if !ok {
		return
	}

	typ := block.Type
	m := formatPropsPerBlock[typ]
	if m == nil {
		m = map[string]bool{}
		formatPropsPerBlock[typ] = m
	}
	for k := range fm {
		m[k] = true
	}
}

func printFormatPropsPerBlock() {
	var blocks []string
	for k := range formatPropsPerBlock {
		blocks = append(blocks, k)
	}
	sort.Strings(blocks)
	for _, typ := range blocks {
		m := formatPropsPerBlock[typ]
		s := fmt.Sprintf("format fields for block: '%s'", typ)
		printMapKeys(m, s)
	}
}

// make sure we can decode the format
func validateFormat(block *notionapi.Block) {
	switch block.Type {
	case notionapi.BlockBookmark:
		block.FormatBookmark()
	}
}

func collectCollectionsInfo(page *notionapi.Page) {
	fn := func(block *notionapi.Block) {
		validateFormat(block)

		recordFormatProps(block)

		if block.Type == notionapi.BlockCollectionView {
			viewInfo := block.CollectionViews[0]
			collection := viewInfo.Collection
			schema := collection.CollectionSchema
			for _, colInfo := range schema {
				typ := colInfo.Type
				collectionSchemaTypes[typ] = true
				/*
					for k := range colInfo.RawJSON {
						collectionSchemeProps[k] = true
					}
				*/
			}
		}
	}
	page.ForEachBlock(fn)

	/*
		for _, table := range page.Tables {
			coll := table.Collection
			for _, schema := range coll.CollectionSchema {
				typ := schema.Type
				collectionSchemaTypes[typ] = true
			}
		}
	*/
}

func printMapKeys(m map[string]bool, s string) {
	var types []string
	for k := range m {
		types = append(types, k)
	}
	sort.Strings(types)
	fmt.Printf("%d %s:\n", len(types), s)
	for _, typ := range types {
		fmt.Printf("  %s\n", typ)
	}
}

// load the page, render to md and html. repeat for all sub-children
func loadAndRenderPageRecur(pageID string) {
	id := notionapi.ToNoDashID(pageID)
	if smokeSeen[id] {
		return
	}
	smokeSeen[id] = true
	page := toHTML(pageID)
	collectCollectionsInfo(page)
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

	smokeSeen = map[string]bool{}
	flgNoOpen = true

	// https://www.notion.so/49d988a60c4a4592bce09938918e8e5b?v=ade5945063da49a3bc79128b06a0683e
	// collection_view_page
	loadAndRenderPageRecur("49d988a60c4a4592bce09938918e8e5b")

	// https://www.notion.so/Relations-rollups-fd56bfc6a3f0471a9f0cc3110ff19a79
	// complicated table, used to crash
	loadAndRenderPageRecur("fd56bfc6a3f0471a9f0cc3110ff19a79")
	// https://www.notion.so/Test-pages-for-notionapi-0367c2db381a4f8b9ce360f388a6b2e3
	// root page of my test pages
	loadAndRenderPageRecur("0367c2db381a4f8b9ce360f388a6b2e3")

	// ad-hoc code to gather stats on blocks and print them
	//printMapKeys(collectionSchemaTypes, "column types")
	//printMapKeys(collectionSchemeProps, "column props")
	//printFormatPropsPerBlock()
}
