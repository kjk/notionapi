package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjk/notionapi"
)

func makeNotionClient() *notionapi.Client {
	client := &notionapi.Client{
		DebugLog:  true,
		AuthToken: getToken(),
	}
	notionToken := strings.TrimSpace(os.Getenv("NOTION_TOKNE"))
	if notionToken == "" {
		log("NOTION_TOKEN env variable not set. Can only access public pages\n")
	} else {
		log("NOTION_TOKEN env variable set, can access private pages\n")
		// TODO: validate that the token looks legit
		client.AuthToken = notionToken
	}
	return client
}

func notionURLForPageID(pageID string) string {
	return "https://notion.so/" + pageID
}

func pathForPageLog(pageID string) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.log.txt", pageID)
	return filepath.Join(logDir, name)
}

func pathForPageRequestsCache(pageID string) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.page.json", pageID)
	return filepath.Join(cacheDir, name)
}

func pathForPageHTML(pageID string) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.page.html", pageID)
	return filepath.Join(cacheDir, name)
}

func pathForPageSimpleStructure(pageID string) string {
	name := fmt.Sprintf("%s.page.structure.txt", pageID)
	return filepath.Join(cacheDir, name)
}

func openLogFileForPage(pageID string) (io.WriteCloser, error) {
	path := pathForPageLog(pageID)
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		return nil, err
	}
	return f, nil
}

func loadRequestCacheForPage(pageID string) *notionapi.HTTPCache {
	if flgNoCache {
		return nil
	}

	path := pathForPageRequestsCache(pageID)
	d, err := ioutil.ReadFile(path)
	if err != nil {
		// it's ok if file doesn't exit
		return nil
	}
	var httpCache *notionapi.HTTPCache
	err = json.Unmarshal(d, &httpCache)
	if err == nil {
		return httpCache
	}
	if err != nil {
		log("json.Unmarshal() failed with %s decoding file %s\n", err, path)
		err = os.Remove(path)
		must(err)
		log("Deleted file %s\n", path)
	}
	return nil
}

// returns path of the created file
func savePageRequestsCacheAsJSON(pageID string, cache *notionapi.HTTPCache) string {
	d, err := json.MarshalIndent(cache, "", "  ")
	must(err)
	path := pathForPageRequestsCache(pageID)
	err = ioutil.WriteFile(path, d, 0644)
	must(err)
	return path
}

// returns path of the created file
func savePageAsSimpleStructure(page *notionapi.Page) string {
	path := pathForPageSimpleStructure(page.ID)
	f, err := os.Create(path)
	must(err)
	defer f.Close()
	notionapi.Dump(f, page)
	return path
}

func downloadPageCached(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	pageID = notionapi.ToNoDashID(pageID)
	httpCache := loadRequestCacheForPage(pageID)
	if httpCache == nil {
		httpCache = notionapi.NewHTTPCache()
	}
	httpClient := notionapi.NewCachingHTTPClient(httpCache)
	prevClient := client.HTTPClient
	client.HTTPClient = httpClient
	defer func() {
		client.HTTPClient = prevClient
	}()

	client.Logger, _ = openLogFileForPage(pageID)
	if client.Logger != nil {
		defer func() {
			f := client.Logger.(*os.File)
			f.Close()
			client.Logger = nil
		}()
	}

	res, err := client.DownloadPage(pageID)
	if err != nil {
		return nil, err
	}
	savePageRequestsCacheAsJSON(pageID, httpCache)
	return res, nil
}

func downloadPage(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	page, err := downloadPageCached(client, pageID)
	if err != nil {
		fmt.Printf("downloadPageCached('%s') failed with %s\n", pageID, err)
		return nil, err
	}
	return page, nil
}
