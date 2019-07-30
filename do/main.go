package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tomd"
)

var (
	flgTestToMd bool
)

const (
	logDir   = "log"
	cacheDir = "cache"
)

var (
	useCache = true
)

func parseFlags() {
	flag.BoolVar(&flgTestToMd, "test-to-md", false, "test markdown generation")
	flag.Parse()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func readZipFile(path string) map[string][]byte {
	r, err := zip.OpenReader(path)
	must(err)
	defer r.Close()
	res := map[string][]byte{}
	for _, f := range r.File {
		rc, err := f.Open()
		must(err)
		d, err := ioutil.ReadAll(rc)
		must(err)
		rc.Close()
		res[f.Name] = d
	}
	return res
}

func reacreateDir(dir string) {
	os.RemoveAll(dir)
	err := os.MkdirAll(dir, 0755)
	must(err)
}

func topDir() string {
	// we start inside "do" directory so topDir is
	// one dir above
	dir, err := filepath.Abs(".")
	must(err)
	return dir
}

func openLogFileForPageID(pageID string) (io.WriteCloser, error) {
	name := fmt.Sprintf("%s.go.log.txt", pageID)
	path := filepath.Join(logDir, name)
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		return nil, err
	}
	return f, nil
}

func downloadPageCached(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	var page notionapi.Page
	cachedPath := filepath.Join(cacheDir, pageID+".json")
	if useCache {
		d, err := ioutil.ReadFile(cachedPath)
		if err == nil {
			err = json.Unmarshal(d, &page)
			if err == nil {
				//fmt.Printf("Got data for pageID %s from cache file %s\n", pageID, cachedPath)
				return &page, nil
			}
			// not a fatal error, just a warning
			fmt.Printf("json.Unmarshal() on '%s' failed with %s\n", cachedPath, err)
		}
	}
	res, err := client.DownloadPage(pageID)
	if err != nil {
		return nil, err
	}
	d, err := json.MarshalIndent(res, "", "  ")
	if err == nil {
		err = ioutil.WriteFile(cachedPath, d, 0644)
		if err != nil {
			// not a fatal error, just a warning
			fmt.Printf("ioutil.WriteFile(%s) failed with %s\n", cachedPath, err)
		}
	} else {
		// not a fatal error, just a warning
		fmt.Printf("json.Marshal() on pageID '%s' failed with %s\n", pageID, err)
	}
	return res, nil
}

func dl(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	client.Logger, _ = openLogFileForPageID(pageID)
	if client.Logger != nil {
		defer func() {
			f := client.Logger.(*os.File)
			f.Close()
		}()
	}
	page, err := downloadPageCached(client, pageID)
	if err != nil {
		fmt.Printf("downloadPageCached('%s') failed with %s\n", pageID, err)
		return nil, err
	}
	return page, nil
}

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
			pages = append(pages, findSubPageIDs(page.Root.Content)...)
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
		if isWhitelisted(pageID) {
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

func isWhitelisted(pageID string) bool {
	for _, s := range whiteListed {
		if normalizeID(s) == normalizeID(pageID) {
			return true
		}
	}
	return false
}

// TODO: make public as it's useful for recursive downloading of pages
func findSubPageIDs(blocks []*notionapi.Block) []string {
	pageIDs := map[string]struct{}{}
	seen := map[string]struct{}{}
	toVisit := blocks
	for len(toVisit) > 0 {
		block := toVisit[0]
		toVisit = toVisit[1:]
		id := normalizeID(block.ID)
		if block.Type == notionapi.BlockPage {
			pageIDs[id] = struct{}{}
			seen[id] = struct{}{}
		}
		for _, b := range block.Content {
			if b == nil {
				continue
			}
			id := normalizeID(block.ID)
			if _, ok := seen[id]; ok {
				continue
			}
			toVisit = append(toVisit, b)
		}
	}
	res := []string{}
	for id := range pageIDs {
		res = append(res, id)
	}
	sort.Strings(res)
	return res
}

func writeFile(path string, data []byte) {
	err := ioutil.WriteFile(path, data, 0666)
	must(err)
}

func openCodeDiff(path1, path2 string) {
	cmd := exec.Command("code", "--new-window", "--diff", path1, path2)
	err := cmd.Start()
	must(err)
}

func openNotepadWithFile(path string) {
	cmd := exec.Command("notepad.exe", path)
	err := cmd.Start()
	must(err)
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

func cdToTopDir() {
	err := os.Chdir("..")
	must(err)
}

func main() {
	cdToTopDir()
	fmt.Printf("topDir: '%s'\n", topDir())
	must(os.MkdirAll(logDir, 0755))
	must(os.MkdirAll(cacheDir, 0755))

	parseFlags()
	if true || flgTestToMd {
		os.Exit(testToMarkdown())
	}
	flag.Usage()
}
