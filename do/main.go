package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/kjk/notionapi"
	"github.com/kjk/u"
)

var (
	dataDir  = "tmpdata"
	cacheDir = filepath.Join(dataDir, "cache")

	flgToken   string
	flgVerbose bool

	// if true, will try to avoid downloading the page by using
	// cached version saved in log/ directory
	flgNoCache bool

	// if true, will not automatically open a browser to display
	// html generated for a page
	flgNoOpen bool

	flgNoFormat bool
	flgReExport bool
)

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

	u.WriteFileMust(path, d)
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
	u.WriteFileMust(name, d)
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
	u.PanicIf(len(subPages) != nExp, "expected %d sub-pages of '%s', got %d", nExp, uri, len(subPages))
}

func traceNotionAPI() {
	nodeModulesDir := filepath.Join("tracenotion", "node_modules")
	if !u.DirExists(nodeModulesDir) {
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

func bench() {
	cmd := exec.Command("go", "test", "-bench=.")
	cmd.Dir = "caching_downloader"
	u.RunCmdMust(cmd)
}

func main() {
	u.CdUpDir("notionapi")
	logf("currDirAbs: '%s'\n", u.CurrDirAbsMust())

	var (
		//flgToken string
		// id of notion page to download
		flgDownloadPage string

		// id of notion page to download and convert to HTML
		flgToHTML     string
		flgToMarkdown string

		flgPreviewHTML     string
		flgPreviewMarkdown string

		flgWc bool

		flgExportPage string
		flgExportType string
		flgRecursive  bool
		flgTrace      bool

		// if true, remove cache directories (data/log, data/cache)
		flgCleanCache bool

		flgSanityTest        bool
		flgSmokeTest         bool
		flgTestToMd          string
		flgTestToHTML        string
		flgTestDownloadCache string
		flgBench             bool
	)

	{
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

		flag.StringVar(&flgPreviewHTML, "preview-html", "", "id of start page")
		flag.StringVar(&flgPreviewMarkdown, "preview-md", "", "id of start page")

		flag.BoolVar(&flgSanityTest, "sanity", false, "runs a quick sanity tests (fast and basic)")
		flag.BoolVar(&flgSmokeTest, "smoke", false, "run a smoke test (not fast, run after non-trivial changes)")
		flag.StringVar(&flgTestDownloadCache, "test-download-cache", "", "page id to use to test download cache")
		flag.StringVar(&flgDownloadPage, "dlpage", "", "id of notion page to download")
		flag.StringVar(&flgToHTML, "to-html", "", "id of notion page to download and convert to html")
		flag.StringVar(&flgToMarkdown, "to-md", "", "id of notion page to download and convert to markdown")
		flag.BoolVar(&flgReExport, "re-export", false, "if true, will re-export from notion")
		flag.BoolVar(&flgNoCache, "no-cache", false, "if true, will not use a cached version in log/ directory")
		flag.BoolVar(&flgNoOpen, "no-open", false, "if true, will not automatically open the browser with html file generated with -tohtml")
		flag.BoolVar(&flgWc, "wc", false, "wc -l on source files")
		flag.BoolVar(&flgBench, "bench", false, "run benchmark")
		flag.Parse()
	}

	must(os.MkdirAll(cacheDir, 0755))

	if false {
		testDownloadFile()
		return
	}

	if false {
		adhocTests()
		return
	}

	if false {
		flgToHTML = "97c2435193d245688bb5da7f84edfe45"
	}

	if false {
		testSubPages()
		return
	}

	// normalize ids early on
	flgDownloadPage = notionapi.ToNoDashID(flgDownloadPage)
	flgToHTML = notionapi.ToNoDashID(flgToHTML)
	flgToMarkdown = notionapi.ToNoDashID(flgToMarkdown)

	if flgWc {
		doLineCount()
		return
	}

	if flgCleanCache {
		{
			dir := filepath.Join(dataDir, "diff")
			os.RemoveAll(dir)
		}
		{
			dir := filepath.Join(dataDir, "smoke")
			os.RemoveAll(dir)
		}
		u.RemoveFilesInDirMust(cacheDir)
		return
	}

	if flgBench {
		bench()
		return
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

	if flgToMarkdown != "" {
		flgNoCache = true
		toMd(flgToMarkdown)
		return
	}

	if flgPreviewHTML != "" {
		uri := "/previewhtml/" + flgPreviewHTML
		startHTTPServer(uri)
		return
	}

	if flgPreviewMarkdown != "" {
		uri := "/previewmd/" + flgPreviewMarkdown
		startHTTPServer(uri)
		return
	}

	flag.Usage()
}

func startHTTPServer(uri string) {
	flgHTTPAddr := "localhost:8503"
	httpSrv := makeHTTPServer()
	httpSrv.Addr = flgHTTPAddr

	logf("Starting on addr: %v\n", flgHTTPAddr)

	chServerClosed := make(chan bool, 1)
	go func() {
		err := httpSrv.ListenAndServe()
		// mute error caused by Shutdown()
		if err == http.ErrServerClosed {
			err = nil
		}
		must(err)
		logf("HTTP server shutdown gracefully\n")
		chServerClosed <- true
	}()

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt /* SIGINT */, syscall.SIGTERM)

	u.OpenBrowser("http://" + flgHTTPAddr + uri)
	time.Sleep(time.Second * 2)

	sig := <-c
	logf("Got signal %s\n", sig)

	if httpSrv != nil {
		// Shutdown() needs a non-nil context
		_ = httpSrv.Shutdown(context.Background())
		select {
		case <-chServerClosed:
			// do nothing
		case <-time.After(time.Second * 5):
			// timeout
		}
	}

}
