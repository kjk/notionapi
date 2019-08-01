package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
)

func toHTML2(page *notionapi.Page) (string, []byte) {
	name := tohtml.HTMLFileNameForPage(page)
	r := tohtml.NewHTMLRenderer(page)
	d := r.ToHTML()
	return name, d
}

func testToHTMLRecur(startPageID string, referenceFiles map[string][]byte) {
	client := &notionapi.Client{
		DebugLog: true,
	}
	seenPages := map[string]bool{}
	pages := []string{startPageID}
	nPage := 0
	for len(pages) > 0 {
		pageID := pages[0]
		pages = pages[1:]

		pageIDNormalized := notionapi.ToNoDashID(pageID)
		if seenPages[pageIDNormalized] {
			continue
		}
		seenPages[pageIDNormalized] = true
		nPage++

		page, err := dl(client, pageID)
		must(err)
		name, pageMd := toHTML2(page)
		fmt.Printf("%02d: '%s'", nPage, name)
		//fmt.Printf("page as markdown:\n%s\n", string(pageMd))
		expData, ok := referenceFiles[name]
		if !ok {
			fmt.Printf("\n'%s' from '%s' doesn't seem correct as it's not present in referenceFiles\n", name, page.Root.Title)
			fmt.Printf("Names in referenceFiles:\n")
			for s := range referenceFiles {
				fmt.Printf("  %s\n", s)
			}
			os.Exit(1)
		}
		if bytes.Equal(pageMd, expData) {
			fmt.Printf(" ok\n")
			pages = append(pages, notionapi.GetSubPages(page.Root.Content)...)
			continue
		}
		if len(pageMd) == len(expData) {
			for i, b := range pageMd {
				bExp := expData[i]
				if b != bExp {
					fmt.Printf("Bytes different at pos %d, got: 0x%x '%c', exp: 0x%x '%c'\n", i, b, b, bExp, bExp)
					goto endloop
				}
			}
		}
	endloop:
		if isMdWhitelisted(pageID) {
			fmt.Printf(" doesn't match but whitelisted\n")
			continue
		}
		fmt.Printf("\nMarkdown in https://notion.so/%s doesn't match\n", notionapi.ToNoDashID(pageID))
		writeFile("exp.md", expData)
		writeFile("got.md", pageMd)
		openCodeDiff(`.\exp.md`, `.\got.md`)
		os.Exit(1)
	}
}

var htmlWhiteListed = []string{}

func isHTMLWhitelisted(pageID string) bool {
	for _, s := range htmlWhiteListed {
		if normalizeID(s) == normalizeID(pageID) {
			return true
		}
	}
	return false
}

func testToHTML() int {
	zipPath := filepath.Join(topDir(), "data", "testdata", "Export-html-6f6dae04-a337-419e-81ca-f82de3202b9e.zip")
	zipFiles := readZipFile(zipPath)
	fmt.Printf("There are %d files in zip file\n", len(zipFiles))

	startPage := "3b617da409454a52bc3a920ba8832bf7" // top-level page for blendle handbok
	//startPage := "2bf22b99850b402882bb885a41cfd981"
	testToMdRecur(startPage, zipFiles)
	return 0
}
