package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml2"
	"github.com/yosssi/gohtml"
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

func testToHTMLRecur(startPageID string, firstToTest string, validBad []string, referenceFiles map[string][]byte) {
	client := &notionapi.Client{
		DebugLog:  true,
		AuthToken: getToken(),
	}
	seenPages := map[string]bool{}
	pages := []string{startPageID}
	nPage := 0
	isDoing := (firstToTest == "")
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
			if idsEqual(pageID, firstToTest) {
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

		if isPageIDInArray(validBad, pageID) {
			fmt.Printf(" doesn't match but whitelisted\n")
			continue
		}

		expDataFormatted := ppHTML(expData)
		gotDataFormatted := ppHTML(pageHTML)

		if bytes.Equal(expDataFormatted, gotDataFormatted) {
			fmt.Printf(", files same after formatting\n")
			continue
		}

		writeFile("exp.html", expDataFormatted)
		writeFile("got.html", gotDataFormatted)
		fmt.Printf("\nHTML in https://notion.so/%s doesn't match\n", notionapi.ToNoDashID(pageID))
		openCodeDiff(`.\exp.html`, `.\got.html`)
		os.Exit(1)
	}
}

func ppHTML(d []byte) []byte {
	s := gohtml.Format(string(d))
	return []byte(s)
}
