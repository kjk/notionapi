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

/*
https://www.notion.so/kjkpublic/Test-page-c969c9455d7c4dd79c7f860f3ace6429
https://www.notion.so/kjkpublic/Test-page-text-4c6a54c68b3e4ea2af9cfaabcc88d58d
https://www.notion.so/kjkpublic/Test-page-text-not-simple-f97ffca91f8949b48004999df34ab1f7
https://www.notion.so/kjkpublic/blog-300db9dc27c84958a08b8d0c37f4cfe5

c969c9455d7c4dd79c7f860f3ace6429   test all
f97ffca91f8949b48004999df34ab1f7   test text not simple
6682351e44bb4f9ca0e149b703265bdb   test header
fd9338a719a24f02993fcfbcf3d00bb0   test todo list and page style
484919a1647144c29234447ce408ff6b   test toggle and bullet list
c969c9455d7c4dd79c7f860f3ace6429
300db9dc27c84958a08b8d0c37f4cfe5   large page (my blog)
0367c2db381a4f8b9ce360f388a6b2e3   index page for test pages
25b6ac21d67744f18a4dc071b21a86fe   test code and favorite
70ecbf1f5abc41d48a4e4320aeb38d10   test todo
97100f9c17324fd7ba3d3c5f1832104d   test dates
0fa8d15a16134f0c9fad1aa0a7232374   test comments, icon, cover
57cb49183ee44eb9a4fcc37817473b54   test deleted page
157765353f2c4705bd45474e5ba8b46c   notion "what's new" page
72fd504c58984cc5a5dfb86b6f8617dc   test nested toggle

available args:
 -recursive -no-cache
*/

/*
.\do.bat -dlpage 4c6a54c68b3e4ea2af9cfaabcc88d58d

Options:
 -use-cache : use on-disk cache to maybe avoid downloading
              data from the server

For testing: downloads a page with a given notion id
and saves http requests and responses in
log/${notionid}.log.txt so that we can look at them
It will also save log/${notionid}.page.json which is
JSON-serialized Page structure.
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

	// if true, remove cache directories (data/log, data/cache)
	flgCleanCache bool

	flgTestToMd        bool
	flgTestToHTML      bool
	flgTestPageMarshal bool
	flgNoFormat        bool
)

var (
	logDir   = filepath.Join("data", "log")
	cacheDir = filepath.Join("data", "cache")
)

var (
	useCache = true
)

func parseFlags() {
	flag.BoolVar(&flgNoFormat, "no-format", false, "if true, doesn't try to reformat/prettify HTML files during HTML testing")
	flag.BoolVar(&flgCleanCache, "clean-cache", false, "if true, cleans cache directories (data/log, data/cache")
	flag.BoolVar(&flgTestToMd, "test-to-md", false, "test markdown generation")
	flag.BoolVar(&flgTestToHTML, "test-to-html", false, "test html generation")
	flag.BoolVar(&flgTestPageMarshal, "test-page-marshal", false, "test marshalling of Page to/from JSON")
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

	parseFlags()

	if flgCleanCache {
		os.RemoveAll(logDir)
		os.RemoveAll(cacheDir)
	}

	must(os.MkdirAll(logDir, 0755))
	must(os.MkdirAll(cacheDir, 0755))

	if flgTestToMd {
		if false {
			os.RemoveAll(logDir)
			must(os.MkdirAll(logDir, 0755))
			os.RemoveAll(cacheDir)
			must(os.MkdirAll(cacheDir, 0755))
		}
		testToMarkdown1()
		return
	}

	if flgTestPageMarshal {
		testPageMarshal()
		return
	}

	if true || flgTestToHTML {
		if true {
			os.RemoveAll(logDir)
			must(os.MkdirAll(logDir, 0755))
			os.RemoveAll(cacheDir)
			must(os.MkdirAll(cacheDir, 0755))
		}

		ensurePrettierExists()
		//testToHTML1()
		//testToHTML2()
		testToHTML3()
		return
	}

	if flgDownloadPage != "" {
		emptyLogDir()
		downloadPageMaybeCached(flgDownloadPage)
		return
	}
	if flgToHTML != "" {
		recreateDir(logDir)
		toHTML(flgToHTML)
		return
	}

	flag.Usage()
}
