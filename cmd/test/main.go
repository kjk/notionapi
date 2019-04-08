package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjk/notionapi"
)

const (
	logDir   = "log"
	cacheDir = "cache"
)

var (
	// id of notion page to download (e.g. "4c6a54c68b3e4ea2af9cfaabcc88d58d")
	flgDownloadPage string
	useCache        = true
)

func parseFlags() {
	flag.StringVar(&flgDownloadPage, "dlpage", "", "id of notion page to download e.g. '4c6a54c68b3e4ea2af9cfaabcc88d58d'")
	flag.Parse()
}

func usageAndExit() {
	flag.Usage()
	os.Exit(1)
}

func logFilePathForPageID(pageID string) string {
	name := fmt.Sprintf("%s.log.txt", pageID)
	return filepath.Join(logDir, name)
}

func pageJSONFilePathForPageID(pageID string) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.page.json", pageID)
	return filepath.Join(logDir, name)
}

func pageSimpleStructureFilePathForPageID(pageID string) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.page.structure.txt", pageID)
	return filepath.Join(logDir, name)
}

func openLogFileForPageID(pageID string) (io.WriteCloser, error) {
	path := logFilePathForPageID(pageID)
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
				fmt.Printf("Got data for pageID %s from cache file %s\n", pageID, cachedPath)
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

func downlaodPageLogged(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
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

func emptyLogDir() {
	os.RemoveAll(logDir)
	os.MkdirAll(logDir, 0755)
}

func notionURLForPageID(pageID string) string {
	pageID = notionapi.ToNoDashID(pageID)
	return "https://notion.so/" + pageID
}

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// returns path of the created file
func savePageAsJSON(page *notionapi.Page) string {
	d, err := json.Marshal(page)
	must(err)
	path := pageJSONFilePathForPageID(page.ID)
	err = ioutil.WriteFile(path, d, 0644)
	must(err)
	return path
}

// returns path of the created file
func savePageAsSimpleStructure(page *notionapi.Page) string {
	path := pageSimpleStructureFilePathForPageID(page.ID)
	f, err := os.Create(path)
	must(err)
	defer f.Close()
	notionapi.Dump(f, page)
	return path
}

func downloadOnePage(pageID string) {
	client := &notionapi.Client{
		DebugLog: true,
	}
	notionToken := strings.TrimSpace(os.Getenv("NOTION_TOKNE"))
	if notionToken == "" {
		fmt.Print("NOTION_TOKEN env variable not set. Can only access public pages\n")
	} else {
		fmt.Print("NOTION_TOKEN env variable set, can access private pages\n")
		// TODO: validate that the token looks legit
		client.AuthToken = notionToken
	}

	pageURL := notionURLForPageID(pageID)
	page, err := downlaodPageLogged(client, pageID)
	if err != nil {
		fmt.Printf("Download of page '%s' failed with %s\n", pageURL, err)
		return
	}

	fmt.Printf("Downloaded page %s\n", pageURL)
	jsonPath := savePageAsJSON(page)
	simplePath := savePageAsSimpleStructure(page)
	fmt.Printf("%s : log of HTTP traffic\n", logFilePathForPageID(pageID))
	fmt.Printf("%s : notionapi.Page serialized as JSON\n", jsonPath)
	fmt.Printf("%s : notionapi.Page serialized in simple format\n", simplePath)
}

func main() {
	emptyLogDir()

	parseFlags()

	if flgDownloadPage != "" {
		downloadOnePage(flgDownloadPage)
		return
	}

	usageAndExit()

}
