package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml2"
	"github.com/yosssi/gohtml"
)

// detect location of https://winmerge.org/
// if present, we can do directory diffs
// only works on windows
func getWinMergePath() string {
	path, err := exec.LookPath("WinMergeU")
	if err == nil {
		return path
	}
	dir, err := os.UserHomeDir()
	if err == nil {
		path := filepath.Join(dir, "AppData", "Local", "Programs", "WinMerge", "WinMergeU.exe")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func dirDiff(dir1, dir2 string) {
	winMerge := getWinMergePath()
	cmd := exec.Command(winMerge, "/r", dir1, dir2)
	err := cmd.Start()
	must(err)
}

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

func testToHTML(startPageID string) {
	os.MkdirAll(cacheDir, 0755)
	os.MkdirAll(logDir, 0755)
	startPageID = notionapi.ToNoDashID(startPageID)

	knownBad := findKnownBadHTML(startPageID)

	referenceFiles := exportPages(startPageID, notionapi.ExportTypeHTML)
	fmt.Printf("There are %d files in zip file\n", len(referenceFiles))

	client := &notionapi.Client{
		DebugLog:  true,
		AuthToken: getToken(),
	}
	seenPages := map[string]bool{}
	pages := []string{startPageID}
	nPage := 0

	hasDirDiff := getWinMergePath() != ""
	diffDir := filepath.Join(dataDir, "diff")
	expDiffDir := filepath.Join(diffDir, "exp")
	gotDiffDir := filepath.Join(diffDir, "got")
	if hasDirDiff {
		must(os.MkdirAll(expDiffDir, 0755))
		must(os.MkdirAll(gotDiffDir, 0755))
		removeFilesInDir(expDiffDir)
		removeFilesInDir(gotDiffDir)
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
		pages = append(pages, notionapi.GetSubPages(page.Root().Content)...)
		name, pageHTML := toHTML2(page)
		fmt.Printf("%02d: %s '%s'", nPage, pageID, name)

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
			if isPageIDInArray(knownBad, pageID) {
				fmt.Printf(" ok (AND ALSO WHITELISTED)\n")
				continue
			}
			fmt.Printf(" ok\n")
			continue
		}

		expDataFormatted := ppHTML(expData)
		gotDataFormatted := ppHTML(pageHTML)

		if bytes.Equal(expDataFormatted, gotDataFormatted) {
			if isPageIDInArray(knownBad, pageID) {
				fmt.Printf(" ok after formatting (AND ALSO WHITELISTED)\n")
				continue
			}
			fmt.Printf(", files same after formatting\n")
			continue
		}

		// if we can diff dirs, run through all files and save files that are
		// differetn in in dirs
		if hasDirDiff {
			fileName := fmt.Sprintf("%s.html", notionapi.ToNoDashID(pageID))
			expPath := filepath.Join(expDiffDir, fileName)
			writeFile(expPath, expDataFormatted)
			gotPath := filepath.Join(gotDiffDir, fileName)
			writeFile(gotPath, gotDataFormatted)
			fmt.Printf("\nHTML in https://notion.so/%s doesn't match\n", notionapi.ToNoDashID(pageID))
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

		fileName := fmt.Sprintf("%s.html", notionapi.ToNoDashID(pageID))
		expPath := "exp-" + fileName
		gotPath := "got-" + fileName
		writeFile(expPath, expDataFormatted)
		writeFile(gotPath, gotDataFormatted)
		fmt.Printf("\nHTML in https://notion.so/%s doesn't match\n", notionapi.ToNoDashID(pageID))
		openCodeDiff(expPath, gotPath)
		os.Exit(1)
	}
}

func ppHTML(d []byte) []byte {
	s := gohtml.Format(string(d))
	return []byte(s)
}
