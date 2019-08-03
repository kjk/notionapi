package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml2"
)

func toHTML2(page *notionapi.Page) (string, []byte) {
	name := tohtml2.HTMLFileNameForPage(page)
	r := tohtml2.NewHTMLRenderer(page)
	r.FullHTML = true
	d := r.ToHTML()
	return name, d
}

// to speed up iteration, we skip pages that we know we render correctly
var toSkipHTML = []string{
	"3b617da409454a52bc3a920ba8832bf7",
	"023663a53df242f9aaf44f192c952754",
	"078cc0bf15a6450dac7b6c061f94f86d",
	"13aa42a5a95d4357aa830c3e7ff35ae1",
	"23b0ea84114b483b96887f30bc453675",
	"2bf22b99850b402882bb885a41cfd981",
	"36430bf61c2a4dec8621a10f220155b5",
	// TODO: those differ but both fail pretty-printing, so no idea
	// what the difference is
	"4f5ee5cf485048468db8dfbf5924409c",
	// TODO: need to redo how title property is handled as my
	// current system seems to loose information
	"5fea966407204d9080a5b989360b205f",
	"619286e4fb4f4198957341b66c98cfb9",
	"6c3b0ff40d8546d5a190ffd26a51be8d",
	"6d25f4e53b914df68630c98ea5523692",
	"745c70bc880a4f88a9f988df70a12eed",
	"772c732082154d47b6f6832a472ba746",
	// TODO: mine is malstructured
	"7a5df17b32e84686ae33bf01fa367da9",
	// TODO: both malformed
	"7afdcc4fbede49bc9582469ad6e86fd3",
	// TODO: need inline redo
	"7e0814fa4a7f415db820acbbb0112aca",
	// TODO: both malformed
	"949f33cdba814fc4a288d81c6e7c810d",
	"8ae3770614e543bf82dba518e61ced66",
	"94a2bcc47fde4dab922968733b9a2a94",
	"94c94534e403472f80baeef87ae3efcf",
	// TODO: inline redo
	"9a00460355b149cd9f9450826c8bebb2",
}

func shouldSkipHTML(pageID string) bool {
	pageID = notionapi.ToNoDashID(pageID)
	for _, s := range toSkipHTML {
		if pageID == s {
			return true
		}
	}
	return false
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
		fmt.Printf("%02d: %s '%s'", nPage, pageID, name)
		if shouldSkipHTML(pageID) {
			fmt.Printf(" skipping known good\n")
			pages = append(pages, notionapi.GetSubPages(page.Root.Content)...)
			continue
		}
		//fmt.Printf("page as markdown:\n%s\n", string(pageMd))
		var expData []byte
		for refName, d := range referenceFiles {
			if strings.HasSuffix(refName, name) {
				expData = d
				break
			}
		}
		if len(expData) == 0 {
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
		if isHTMLWhitelisted(pageID) {
			fmt.Printf(" doesn't match but whitelisted\n")
			continue
		}
		writeFile("exp.html", expData)
		writeFile("got.html", pageMd)
		if shouldFormat() {
			formatHTMLFile("exp.html")
			formatHTMLFile("got.html")
			if areFilesEuqal("exp.html", "got.html") {
				fmt.Printf(", files same after formatting\n")
				pages = append(pages, notionapi.GetSubPages(page.Root.Content)...)
				continue
			}
		}
		fmt.Printf("\nHTML in https://notion.so/%s doesn't match\n", notionapi.ToNoDashID(pageID))
		openCodeDiff(`.\exp.html`, `.\got.html`)
		os.Exit(1)
	}
}

func shouldFormat() bool {
	return !flgNoFormat
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
	if shouldFormat() {
		ensurePrettierExists()
	}
	zipPath := filepath.Join(topDir(), "data", "testdata", "Export-html-6f6dae04-a337-419e-81ca-f82de3202b9e.zip")
	zipFiles := readZipFile(zipPath)
	fmt.Printf("There are %d files in zip file\n", len(zipFiles))

	startPage := "3b617da409454a52bc3a920ba8832bf7" // top-level page for blendle handbok
	//startPage = "13aa42a5a95d4357aa830c3e7ff35ae1"
	testToHTMLRecur(startPage, zipFiles)
	return 0
}
