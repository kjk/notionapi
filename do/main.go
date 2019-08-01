package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kjk/notionapi"
)

/*
.\do.bat -tohtml 4c6a54c68b3e4ea2af9cfaabcc88d58d

Options:
  -no-open   : won't automatically open the web browser
  -use-cache : use on-disk cache to maybe avoid downloading
               data from the server

For testing: downloads a page with a given notion id
and converts to HTML. Saves the html to log/ directory
and opens browser with that page
*/

var (
	// id of notion page looks like this:
	// 4c6a54c68b3e4ea2af9cfaabcc88d58d

	// id of notion page to download
	flgDownloadPage string

	// id of notion page to download and convert to HTML
	flgToHTML string

	// if true, will try to avoid downloading the page by using
	// cached version sved in log/ directory
	flgUseCache bool

	// if true, will not automatically open a browser to display
	// html generated for a page
	flgNoOpen bool

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
	flag.StringVar(&flgDownloadPage, "dlpage", "", "id of notion page to download")
	flag.StringVar(&flgToHTML, "tohtml", "", "id of notion page to download and convert to html")
	flag.BoolVar(&flgUseCache, "use-cache", false, "if true will try to avoid downloading the page by using cached version saved in log/ directory")
	flag.BoolVar(&flgNoOpen, "no-open", false, "if true will not automatically open the browser with html file generated with -tohtml")
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
	err := os.Chdir("..")
	must(err)
}

func main() {
	cdToTopDir()
	fmt.Printf("topDir: '%s'\n", topDir())
	must(os.MkdirAll(logDir, 0755))
	must(os.MkdirAll(cacheDir, 0755))

	parseFlags()
	if flgTestToMd {
		testToMarkdown()
		return
	}

	if flgDownloadPage != "" {
		emptyLogDir()
		downloadPageMaybeCached(flgDownloadPage)
		return
	}
	if flgToHTML != "" {
		emptyLogDir()
		toHTML(flgToHTML)
		return
	}

	flag.Usage()
}
