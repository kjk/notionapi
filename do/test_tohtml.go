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

	// TODO(1): Notion renders <div> inside <p> which is illegal and makes pretty-printing
	// not work, so can't compare results. Probably because they render children
	//  <p> inside </p> instead of after </p>
	"4f5ee5cf485048468db8dfbf5924409c",

	"5fea966407204d9080a5b989360b205f",
	"619286e4fb4f4198957341b66c98cfb9",
	"6c3b0ff40d8546d5a190ffd26a51be8d",
	"6d25f4e53b914df68630c98ea5523692",
	"745c70bc880a4f88a9f988df70a12eed",
	"772c732082154d47b6f6832a472ba746",

	// TODO: mine has extract link-to-page and bad user rendering
	"7a5df17b32e84686ae33bf01fa367da9",

	// TODO: Notion's malformed
	"7afdcc4fbede49bc9582469ad6e86fd3",
	"7e0814fa4a7f415db820acbbb0112aca",
	"8ae3770614e543bf82dba518e61ced66",
	// TODO: Notion's malformed
	"949f33cdba814fc4a288d81c6e7c810d",
	"94a2bcc47fde4dab922968733b9a2a94",
	"94c94534e403472f80baeef87ae3efcf",
	"9a00460355b149cd9f9450826c8bebb2",
	"9cc14382e3c34037bf80a4936a9b6674",
	"a881aeee28254ecb8490188e248019ae",
	"ab2af85726b94440904826eb37192dca",
	// TODO: mine has extra link-to-page
	"b1b31f6d3405466c988676f996ce03ad",
	"cddcb453eaa5435a92a364d147425b9e",
	"d0464f97636448fd8dab5497f68394c2",
	"d1fe3bd9514a4543ae43194333f3cbd2",
	"d82df6d6fafe47d590cd40f33a06e263",
	"dc31f1bfdd7146fba42986365c33c37e",
	// TODO: extra link-to-page
	"ef850413bb53491eaebccaa09eeb8630",
	// TODO: Notion's malformed
	"f2d97c9cba804583838acf5d571313f5",
	"f495439c3d54409ca714fc3c7cc5711f",
	"8048fc1b43994344af9979ead9017aef",
	"bf5d1c1f793a443ca4085cc99186d32f",
	// TODO: Notion's malformed
	"3c892714f4dc4d2194619fdccba48fc6",
	"b2a41db3032049f6a5e2ff66642268b7",
	"8f12cc5182a6437aac4dc518cb28b681",
	"11147d69498a40aeb9ea706f4428c38d",
	"3350e580b2174d589aa3edfd70741f44",
	"bf7ba51334644a6498a196b99106f682",
	"13b8fb98f56848c2814eaf453c2da1e7",
	"143d0aef49d54e7ca19eac7b912b5b40",
	"473db4b892c942648d3e3e041c2945d9",
	"c29a8c69877442278c04ce8cdd49a0a0",
	"d9ccc316dc0b4fe88166e364d014c5fe",
	"e35f07fbb0c541399350e80cb0252530",
	"076006f590e046508e27b08b017b64bd",
	"4974956cb53149719bce4dc62199a908",
	"a7457b45915246b8be4efef8d3529da8",
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
