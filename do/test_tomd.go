package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tomd"
)

func toMD(page *notionapi.Page) (string, []byte) {
	name := tomd.MarkdownFileNameForPage(page)
	r := tomd.NewMarkdownRenderer(page)
	d := r.ToMarkdown()
	return name, d
}

func testToMdRecur(startPageID string, referenceFiles map[string][]byte) {
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
		name, pageMd := toMD(page)
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

func normalizeID(s string) string {
	return notionapi.ToNoDashID(s)
}

var whiteListed = []string{
	// bad rendering of inlines in Notion
	"36430bf6-1c2a-4dec-8621-a10f220155b5",
	// mine misses one newline, neither mine nor notion is correct
	"4f5ee5cf-4850-4846-8db8-dfbf5924409c",
	// difference in inlines rendering and I have extra space
	"5fea966407204d9080a5b989360b205f",
	// bad link & bold render in Notion
	"619286e4fb4f4198957341b66c98cfb9",
	// can't resolve user name. Not sure why, the users are in cached
	// .json of the file but not in the Page object
	"7a5df17b32e84686ae33bf01fa367da9",
	// different bold/link render (Notion seems to try to merge adjacent links)
	"7afdcc4fbede49bc9582469ad6e86fd3",
	// difference in bold rendering, both valid
	"7e0814fa4a7f415db820acbbb0112aca",
	// missing newline in numbered list. Maybe the rule should be
	// to put newline if there are non-empty children
	// or: supress newline before lists if previous is also list
	// without children
	"949f33cdba814fc4a288d81c6e7c810d",
	// bold/url, newline
	"94c94534e403472f80baeef87ae3efcf",
	// bold (Notion collapses multiple)
	"d0464f97636448fd8dab5497f68394c2",
	// bold
	"d1fe3bd9514a4543ae43194333f3cbd2",
	// bold
	"d82df6d6fafe47d590cd40f33a06e263",
	// bold, newline
	"f2d97c9cba804583838acf5d571313f5",
	// italic, bold
	"f495439c3d54409ca714fc3c7cc5711f",
	// bold
	"bf5d1c1f793a443ca4085cc99186d32f",
	// newline
	"b2a41db3032049f6a5e2ff66642268b7",
	// Notion has a bug in (undefined), bold
	"13b8fb98f56848c2814eaf453c2da1e7",
	// missing newline in mine
	"143d0aef49d54e7ca19eac7b912b5b40",
	// bold, newline
	"473db4b892c942648d3e3e041c2945d9",
	// "undefined"
	"c29a8c69877442278c04ce8cdd49a0a0",
}

func isMdWhitelisted(pageID string) bool {
	for _, s := range whiteListed {
		if normalizeID(s) == normalizeID(pageID) {
			return true
		}
	}
	return false
}

func testToMarkdown() int {
	zipPath := filepath.Join(topDir(), "testdata", "Export-b676ebbf-10ea-465f-aa21-158fc9b2ec82.zip")
	zipFiles := readZipFile(zipPath)
	fmt.Printf("There are %d files in zip file\n", len(zipFiles))

	startPage := "3b617da409454a52bc3a920ba8832bf7" // top-level page for blendle handbok
	//startPage := "2bf22b99850b402882bb885a41cfd981"
	testToMdRecur(startPage, zipFiles)
	return 0
}
