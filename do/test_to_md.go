package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tomarkdown"
	"github.com/kjk/u"
)

var knownBadMarkdown = [][]string{
	{
		"3b617da409454a52bc3a920ba8832bf7",

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
	},
}

func normalizeID(s string) string {
	return notionapi.ToNoDashID(s)
}

func getKnownBadMarkdown(pageID string) []string {
	for _, a := range knownBadMarkdown {
		if a[0] == pageID {
			return a[1:]
		}
	}
	return nil
}

func isPageIDInArray(a []string, pageID string) bool {
	pageID = notionapi.ToNoDashID(pageID)
	for _, s := range a {
		if notionapi.ToNoDashID(s) == pageID {
			return true
		}
	}
	return false
}

func toMarkdown(page *notionapi.Page) (string, []byte) {
	name := tomarkdown.MarkdownFileNameForPage(page)
	r := tomarkdown.NewConverter(page)
	d := r.ToMarkdown()
	return name, d
}

func isReferenceMarkdownName(referenceName string, name string, id string) bool {
	id = notionapi.ToDashID(id)
	if strings.Contains(referenceName, id) {
		return true
	}
	return false
}

func findReferenceMarkdownData(referenceFiles map[string][]byte, name string, id string) ([]byte, bool) {
	for referenceName, d := range referenceFiles {
		if isReferenceMarkdownName(referenceName, name, id) {
			return d, true
		}
	}
	return nil, false
}

func exportPages(pageID string, exportType string) map[string][]byte {
	var ext string
	switch exportType {
	case notionapi.ExportTypeMarkdown:
		ext = "md"
	case notionapi.ExportTypeHTML:
		ext = "html"
	}
	name := pageID + "-" + ext + ".zip"
	zipPath := filepath.Join(dataDir, name)
	if flgReExport {
		os.Remove(zipPath)
	}

	if _, err := os.Stat(zipPath); err != nil {
		if getToken() == "" {
			fmt.Printf("Must provide token with -token arg or by setting NOTION_TOKEN env variable\n")
			os.Exit(1)
		}
		fmt.Printf("Downloading %s\n", zipPath)
		must(exportPageToFile(pageID, exportType, true, zipPath))
	}

	return u.ReadZipFileMust(zipPath)
}

func testToMarkdown(startPageID string) {
	startPageID = notionapi.ToNoDashID(startPageID)

	knownBad := getKnownBadMarkdown(startPageID)

	referenceFiles := exportPages(startPageID, notionapi.ExportTypeMarkdown)
	fmt.Printf("There are %d files in zip file\n", len(referenceFiles))

	client := &notionapi.Client{
		DebugLog:  flgVerbose,
		AuthToken: getToken(),
	}
	seenPages := map[string]bool{}
	pages := []string{startPageID}
	nPage := 0

	hasDirDiff := getDiffToolPath() != ""
	diffDir := filepath.Join(dataDir, "diff")
	expDiffDir := filepath.Join(diffDir, "exp")
	gotDiffDir := filepath.Join(diffDir, "got")
	if hasDirDiff {
		must(os.MkdirAll(expDiffDir, 0755))
		must(os.MkdirAll(gotDiffDir, 0755))
		u.RemoveFilesInDirMust(expDiffDir)
		u.RemoveFilesInDirMust(gotDiffDir)
	}
	nDifferent := 0

	for len(pages) > 0 {
		pageID := pages[0]
		pages = pages[1:]

		pageIDNormalized := notionapi.ToNoDashID(pageID)
		if seenPages[pageIDNormalized] {
			continue
		}
		seenPages[pageIDNormalized] = true
		nPage++

		page, err := downloadPage(client, pageID)
		must(err)
		pages = append(pages, page.GetSubPages()...)
		name, pageMd := toMarkdown(page)
		fmt.Printf("%02d: '%s'", nPage, name)

		expData, ok := findReferenceMarkdownData(referenceFiles, name, pageID)
		if !ok {
			fmt.Printf("\n'%s' from '%s' doesn't seem correct as it's not present in referenceFiles\n", name, page.Root().Title)
			fmt.Printf("Names in referenceFiles:\n")
			for s := range referenceFiles {
				fmt.Printf("  %s\n", s)
			}
			os.Exit(1)
		}

		if bytes.Equal(pageMd, expData) {
			if isPageIDInArray(knownBad, pageID) {
				fmt.Printf(" ok (AND ALSO WHITELISTED)\n")
				continue
			}
			fmt.Printf(" ok\n")
			continue
		}

		// if we can diff dirs, run through all files and save files that are
		// differetn in in dirs
		if hasDirDiff {
			fileName := fmt.Sprintf("%s.md", notionapi.ToNoDashID(pageID))
			expPath := filepath.Join(expDiffDir, fileName)
			u.WriteFileMust(expPath, expData)
			gotPath := filepath.Join(gotDiffDir, fileName)
			u.WriteFileMust(gotPath, pageMd)
			fmt.Printf(" https://notion.so/%s doesn't match\n", notionapi.ToNoDashID(pageID))
			if nDifferent == 0 {
				dirDiff(expDiffDir, gotDiffDir)
			}
			nDifferent++
			continue
		}

		if isPageIDInArray(knownBad, pageID) {
			fmt.Printf(" doesn't match but whitelisted\n")
			continue
		}

		fmt.Printf("\nMarkdown in https://notion.so/%s doesn't match\n", notionapi.ToNoDashID(pageID))

		fileName := fmt.Sprintf("%s.md", notionapi.ToNoDashID(pageID))
		expPath := "exp-" + fileName
		gotPath := "got-" + fileName
		u.WriteFileMust(expPath, expData)
		u.WriteFileMust(gotPath, pageMd)
		openCodeDiff(expPath, gotPath)
		os.Exit(1)
	}
}
