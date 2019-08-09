package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	flgExportPage string
	flgExportType string
	flgRecursive  bool

	// if true, remove cache directories (data/log, data/cache)
	flgCleanCache bool
	flgReExport   bool

	flgSanityTest          bool
	flgTestToMd            string
	flgTestToHTML          string
	flgTestToHTML1         bool
	flgTestToHTML2         bool
	flgTestToHTML3         bool
	flgTestPageJSONMarshal string
	flgNoFormat            bool
)

var (
	dataDir  = "data"
	cacheDir = filepath.Join("data", "cache")
	logDir   = filepath.Join(cacheDir, "log")
)

func parseFlags() {
	flag.BoolVar(&flgNoFormat, "no-format", false, "if true, doesn't try to reformat/prettify HTML files during HTML testing")
	flag.BoolVar(&flgCleanCache, "clean-cache", false, "if true, cleans cache directories (data/log, data/cache")
	flag.StringVar(&flgToken, "token", "", "auth token")
	flag.BoolVar(&flgRecursive, "recursive", false, "if true, recursive export")
	flag.StringVar(&flgExportPage, "export-page", "", "id of the page to export")
	flag.StringVar(&flgExportType, "export-type", "", "html or markdown")
	flag.StringVar(&flgTestToMd, "test-to-md", "", "test markdown generation")
	flag.StringVar(&flgTestToHTML, "test-to-html", "", "id of start page")
	flag.BoolVar(&flgTestToHTML1, "test-to-html1", false, "test html 1 generation")
	flag.BoolVar(&flgTestToHTML2, "test-to-html2", false, "test html 2 generation")
	flag.BoolVar(&flgTestToHTML3, "test-to-html3", false, "test html 3 generation")
	flag.BoolVar(&flgSanityTest, "sanity", false, "if true, runs a sanity tests")
	flag.StringVar(&flgTestPageJSONMarshal, "test-json-marshal", "", "test marshalling of a given page to/from JSON")
	flag.StringVar(&flgDownloadPage, "dlpage", "", "id of notion page to download")
	flag.StringVar(&flgToHTML, "tohtml", "", "id of notion page to download and convert to html")
	flag.BoolVar(&flgReExport, "re-export", false, "if true, will re-export from notion")
	flag.BoolVar(&flgNoCache, "no-cache", false, "if true, will not use a cached version in log/ directory")
	flag.BoolVar(&flgNoOpen, "no-open", false, "if true, will not automatically open the browser with html file generated with -tohtml")
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

func removeFilesInDir(dir string) {
	os.MkdirAll(dir, 0755)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, fi := range files {
		if !fi.Mode().IsRegular() {
			continue
		}
		path := filepath.Join(dir, fi.Name())
		os.Remove(path)
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
		DebugLog:  true,
		Logger:    os.Stdout,
		AuthToken: getToken(),
	}

	if exportType == "" {
		exportType = "html"
	}
	d, err := client.ExportPages(id, exportType, recursive)
	if err != nil {
		fmt.Printf("client.ExportPages() failed with '%s'\n", err)
		return err
	}

	err = ioutil.WriteFile(path, d, 0755)
	if err != nil {
		fmt.Printf("ioutil.WriteFile() failed with '%s'\n", err)
		return err
	}
	fmt.Printf("Downloaded exported page of id %s as %s\n", id, path)
	return nil
}

func panicIf(cond bool, args ...interface{}) {
	if !cond {
		return
	}
	if len(args) == 0 {
		panic("condition failed")
	}
	format := args[0].(string)
	if len(args) == 1 {
		panic(format)
	}
	panic(fmt.Sprintf(format, args[1:]))
}

func exportPage(id string, exportType string, recursive bool) {
	client := &notionapi.Client{
		DebugLog:  true,
		Logger:    os.Stdout,
		AuthToken: getToken(),
	}

	if exportType == "" {
		exportType = "html"
	}
	d, err := client.ExportPages(id, exportType, recursive)
	if err != nil {
		fmt.Printf("client.ExportPages() failed with '%s'\n", err)
		return
	}
	name := notionapi.ToNoDashID(id) + "-" + exportType + ".zip"
	err = ioutil.WriteFile(name, d, 0755)
	if err != nil {
		fmt.Printf("ioutil.WriteFile() failed with '%s'\n", err)
	}
	fmt.Printf("Downloaded exported page of id %s as %s\n", id, name)
}

func runGoTests() {
	cmd := exec.Command("go", "test", "-v", "./...")
	fmt.Printf("Running: %s\n", strings.Join(cmd.Args, " "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	must(cmd.Run())
}

// sanity tests are basic tests to validate changes
// meant to not take too long
func sanityTests() {
	fmt.Printf("Running sanity tests\n")
	runGoTests()
	testPageJSONMarshal("dd5c0a813dfe4487a6cd432f82c0c2fc")
	// TODO: more tests?
}

func main() {
	cdToTopDir()
	fmt.Printf("topDir: '%s'\n", topDir())

	parseFlags()

	if flgCleanCache {
		removeFilesInDir(cacheDir)
	}

	if flgTestToMd != "" {
		testToMarkdown(flgTestToMd)
		return
	}

	if flgSanityTest {
		sanityTests()
		return
	}

	if flgExportPage != "" {
		exportPage(flgExportPage, flgExportType, flgRecursive)
		return
	}

	if flgTestPageJSONMarshal != "" {
		testPageJSONMarshal(flgTestPageJSONMarshal)
		return
	}

	if false {
		removeFilesInDir(cacheDir)
	}

	if flgTestToHTML != "" {
		testToHTML(flgTestToHTML)
		return
	}

	if flgTestToHTML1 {
		testToHTML("3b617da409454a52bc3a920ba8832bf7")
		return
	}
	if flgTestToHTML2 {
		testToHTML("0367c2db381a4f8b9ce360f388a6b2e3")
		return
	}

	if flgTestToHTML3 {
		testToHTML("d6eb49cfc68f402881af3aef391443e6")
		return
	}

	if flgDownloadPage != "" {
		client := makeNotionClient()
		downloadPage(client, flgDownloadPage)
		return
	}

	if flgToHTML != "" {
		recreateDir(logDir)
		toHTML(flgToHTML)
		return
	}

	flag.Usage()
}
