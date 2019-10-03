package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kjk/notionapi"
)

/*
.\do.bat -tohtml 4c6a54c68b3e4ea2af9cfaabcc88d58d
*/

var (
	// id of notion page looks like this:
	// 4c6a54c68b3e4ea2af9cfaabcc88d58d

	flgToken string
	// id of notion page to download
	flgDownloadPage string

	// id of notion page to download and convert to HTML
	flgToHTML string

	// if true, will try to avoid downloading the page by using
	// cached version saved in log/ directory
	flgNoCache bool

	// if true, will not automatically open a browser to display
	// html generated for a page
	flgNoOpen bool

	flgWc bool

	flgExportPage string
	flgExportType string
	flgRecursive  bool
	flgVerbose    bool
	flgTrace      bool

	// if true, remove cache directories (data/log, data/cache)
	flgCleanCache bool
	flgReExport   bool

	flgSanityTest        bool
	flgSmokeTest         bool
	flgTestToMd          string
	flgTestToHTML        string
	flgTestDownloadCache string
	flgNoFormat          bool
)

var (
	dataDir  = "data"
	cacheDir = filepath.Join("data", "cache")
)

func parseFlags() {
	flag.BoolVar(&flgNoFormat, "no-format", false, "if true, doesn't try to reformat/prettify HTML files during HTML testing")
	flag.BoolVar(&flgCleanCache, "clean-cache", false, "if true, cleans cache directories (data/log, data/cache")
	flag.StringVar(&flgToken, "token", "", "auth token")
	flag.BoolVar(&flgRecursive, "recursive", false, "if true, recursive export")
	flag.BoolVar(&flgVerbose, "verbose", false, "if true, verbose logging")
	flag.StringVar(&flgExportPage, "export-page", "", "id of the page to export")
	flag.BoolVar(&flgTrace, "trace", false, "run node tracenotion/trace.js")
	flag.StringVar(&flgExportType, "export-type", "", "html or markdown")
	flag.StringVar(&flgTestToMd, "test-to-md", "", "test markdown generation")
	flag.StringVar(&flgTestToHTML, "test-to-html", "", "id of start page")
	flag.BoolVar(&flgSanityTest, "sanity", false, "runs a quick sanity tests (fast and basic)")
	flag.BoolVar(&flgSmokeTest, "smoke", false, "run a smoke test (not fast, run after non-trivial changes)")
	flag.StringVar(&flgTestDownloadCache, "test-download-cache", "", "page id to use to test download cache")
	flag.StringVar(&flgDownloadPage, "dlpage", "", "id of notion page to download")
	flag.StringVar(&flgToHTML, "to-html", "", "id of notion page to download and convert to html")
	flag.BoolVar(&flgReExport, "re-export", false, "if true, will re-export from notion")
	flag.BoolVar(&flgNoCache, "no-cache", false, "if true, will not use a cached version in log/ directory")
	flag.BoolVar(&flgNoOpen, "no-open", false, "if true, will not automatically open the browser with html file generated with -tohtml")
	flag.BoolVar(&flgWc, "wc", false, "wc -l on source files")
	flag.Parse()

	// normalize ids early on
	flgDownloadPage = notionapi.ToNoDashID(flgDownloadPage)
	flgToHTML = notionapi.ToNoDashID(flgToHTML)
}

// absolute path of top directory in the repo
func topDir() string {
	dir, err := filepath.Abs(".")
	must(err)
	return dir
}

// we are executed for do/ directory so top dir is parent dir
func cdToTopDir() {
	startDir, err := os.Getwd()
	must(err)
	startDir, err = filepath.Abs(startDir)
	must(err)
	dir := startDir
	for {
		// we're already in top directory
		if filepath.Base(dir) == "notionapi" {
			err = os.Chdir(dir)
			must(err)
			return
		}
		parentDir := filepath.Dir(dir)
		panicIf(dir == parentDir, "invalid startDir: '%s', dir: '%s'", startDir, dir)
		dir = parentDir
	}
}

func removeFilesInDir(dir string) {
	err := os.MkdirAll(dir, 0755)
	must(err)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, fi := range files {
		if !fi.Mode().IsRegular() {
			continue
		}
		path := filepath.Join(dir, fi.Name())
		err = os.Remove(path)
		must(err)
	}
}

func getToken() string {
	if flgToken != "" {
		return flgToken
	}
	return os.Getenv("NOTION_TOKEN")
}

func exportPageToFile(id string, exportType string, recursive bool, path string) error {
	client := &notionapi.Client{
		DebugLog:  flgVerbose,
		Logger:    os.Stdout,
		AuthToken: getToken(),
	}

	if exportType == "" {
		exportType = "html"
	}
	d, err := client.ExportPages(id, exportType, recursive)
	if err != nil {
		logf("client.ExportPages() failed with '%s'\n", err)
		return err
	}

	writeFile(path, d)
	logf("Downloaded exported page of id %s as %s\n", id, path)
	return nil
}

func exportPage(id string, exportType string, recursive bool) {
	client := &notionapi.Client{
		DebugLog:  flgVerbose,
		Logger:    os.Stdout,
		AuthToken: getToken(),
	}

	if exportType == "" {
		exportType = "html"
	}
	d, err := client.ExportPages(id, exportType, recursive)
	if err != nil {
		logf("client.ExportPages() failed with '%s'\n", err)
		return
	}
	name := notionapi.ToNoDashID(id) + "-" + exportType + ".zip"
	writeFile(name, d)
	logf("Downloaded exported page of id %s as %s\n", id, name)
}

func runGoTests() {
	cmd := exec.Command("go", "test", "-v", "./...")
	logf("Running: %s\n", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}

func testSubPages() {
	// test that GetSubPages() only returns direct children
	// of a page, not link to pages
	client := &notionapi.Client{}
	uri := "https://www.notion.so/Test-sub-pages-in-mono-font-381243f4ba4d4670ac491a3da87b8994"
	pageID := "381243f4ba4d4670ac491a3da87b8994"
	page, err := client.DownloadPage(pageID)
	must(err)
	subPages := page.GetSubPages()
	nExp := 7
	panicIf(len(subPages) != nExp, "expected %d sub-pages of '%s', got %d", nExp, uri, len(subPages))
}

func traceNotionAPI() {
	nodeModulesDir := filepath.Join("tracenotion", "node_modules")
	if !dirExists(nodeModulesDir) {
		cmd := exec.Command("yarn")
		cmd.Dir = "tracenotion"
		err := cmd.Run()
		must(err)
	}
	scriptPath := filepath.Join("tracenotion", "trace.js")
	cmd := exec.Command("node", scriptPath)
	cmd.Args = append(cmd.Args, flag.Args()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	must(err)
}

func main() {
	cdToTopDir()
	logf("topDir: '%s'\n", topDir())

	parseFlags()
	must(os.MkdirAll(cacheDir, 0755))

	if false {
		testSubPages()
		return
	}

	if flgWc {
		doLineCount()
		return
	}

	if flgCleanCache {
		removeFilesInDir(cacheDir)
	}

	if flgSanityTest {
		sanityTests()
		return
	}

	if flgSmokeTest {
		// smoke test includes sanity test
		sanityTests()
		smokeTest()
		return
	}

	if flgTrace {
		traceNotionAPI()
		return
	}

	if flgTestToMd != "" {
		testToMarkdown(flgTestToMd)
		return
	}

	if flgExportPage != "" {
		exportPage(flgExportPage, flgExportType, flgRecursive)
		return
	}

	if flgTestDownloadCache != "" {
		testCachingDownloads(flgTestDownloadCache)
		return
	}

	if flgTestToHTML != "" {
		testToHTML(flgTestToHTML)
		return
	}

	if flgDownloadPage != "" {
		client := makeNotionClient()
		downloadPage(client, flgDownloadPage)
		return
	}

	if flgToHTML != "" {
		flgNoCache = true
		toHTML(flgToHTML)
		return
	}

	flag.Usage()
}
