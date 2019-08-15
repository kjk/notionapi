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
	flgVerbose    bool

	// if true, remove cache directories (data/log, data/cache)
	flgCleanCache bool
	flgReExport   bool

	flgSanityTest          bool
	flgTestToMd            string
	flgTestToHTML          string
	flgTestPageJSONMarshal string
	flgNoFormat            bool
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
	flag.StringVar(&flgExportType, "export-type", "", "html or markdown")
	flag.StringVar(&flgTestToMd, "test-to-md", "", "test markdown generation")
	flag.StringVar(&flgTestToHTML, "test-to-html", "", "id of start page")
	flag.BoolVar(&flgSanityTest, "sanity", false, "if true, runs a sanity tests")
	flag.StringVar(&flgTestPageJSONMarshal, "test-json-marshal", "", "test marshalling of a given page to/from JSON")
	flag.StringVar(&flgDownloadPage, "dlpage", "", "id of notion page to download")
	flag.StringVar(&flgToHTML, "to-html", "", "id of notion page to download and convert to html")
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
		DebugLog:  flgVerbose,
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
		DebugLog:  flgVerbose,
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
	cmd := exec.Command("go", "test", "./...")
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
	must(os.MkdirAll(cacheDir, 0755))

	if flgCleanCache {
		removeFilesInDir(cacheDir)
	}

	if flgSanityTest {
		sanityTests()
		return
	}

	if false {
		flgTestToMd = "0367c2db381a4f8b9ce360f388a6b2e3"
	}

	if flgTestToMd != "" {
		testToMarkdown(flgTestToMd)
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
