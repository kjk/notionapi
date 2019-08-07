package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml2"
)

func shouldFormat() bool {
	return !flgNoFormat
}

func toHTML2(page *notionapi.Page) (string, []byte) {
	name := tohtml2.HTMLFileNameForPage(page)
	r := tohtml2.NewConverter(page)
	r.FullHTML = true
	d := r.ToHTML()
	return name, d
}

func idsEqual(id1, id2 string) bool {
	id1 = notionapi.ToDashID(id1)
	id2 = notionapi.ToDashID(id2)
	return id1 == id2
}

func testToHTMLRecur(startPageID string, startWith string, validBad []string, referenceFiles map[string][]byte) {
	client := &notionapi.Client{
		DebugLog:  true,
		AuthToken: getToken(),
	}
	seenPages := map[string]bool{}
	pages := []string{startPageID}
	nPage := 0
	isDoing := (startWith == "")
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
		pages = append(pages, notionapi.GetSubPages(page.Root().Content)...)
		name, pageHTML := toHTML2(page)
		fmt.Printf("%02d: %s '%s'", nPage, pageID, name)

		if !isDoing {
			if idsEqual(pageID, startWith) {
				isDoing = true
			}
		}
		if !isDoing {
			fmt.Printf(" skipped\n")
			continue
		}

		//fmt.Printf("page as html:\n%s\n", string(pageHTML))
		var expData []byte
		for refName, d := range referenceFiles {
			if strings.HasSuffix(refName, name) {
				expData = d
				break
			}
		}
		if len(expData) == 0 {
			fmt.Printf("\n'%s' from '%s' doesn't seem correct as it's not present in referenceFiles\n", name, page.Root().Title)
			fmt.Printf("Names in referenceFiles:\n")
			for s := range referenceFiles {
				fmt.Printf("  %s\n", s)
			}
			os.Exit(1)
		}
		if bytes.Equal(pageHTML, expData) {
			fmt.Printf(" ok\n")
			continue
		}
		if len(pageHTML) == len(expData) {
			for i, b := range pageHTML {
				bExp := expData[i]
				if b != bExp {
					fmt.Printf("Bytes different at pos %d, got: 0x%x '%c', exp: 0x%x '%c'\n", i, b, b, bExp, bExp)
					goto endloop
				}
			}
		}
	endloop:
		if isPageIDInArray(validBad, pageID) {
			fmt.Printf(" doesn't match but whitelisted\n")
			continue
		}

		writeFile("exp.html", expData)
		writeFile("got.html", pageHTML)
		if shouldFormat() {
			formatHTMLFile("exp.html")
			formatHTMLFile("got.html")
			if areFilesEuqal("exp.html", "got.html") {
				fmt.Printf(", files same after formatting\n")
				pages = append(pages, notionapi.GetSubPages(page.Root().Content)...)
				continue
			}
		}
		fmt.Printf("\nHTML in https://notion.so/%s doesn't match\n", notionapi.ToNoDashID(pageID))
		openCodeDiff(`.\exp.html`, `.\got.html`)
		os.Exit(1)
	}
}
