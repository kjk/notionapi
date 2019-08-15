package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjk/notionapi/caching_downloader"

	"github.com/kjk/notionapi"
)

func makeNotionClient() *notionapi.Client {
	client := &notionapi.Client{
		DebugLog:  flgVerbose,
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

func pathForPageHTML(pageID string) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.page.html", pageID)
	return filepath.Join(cacheDir, name)
}

func pathForPageSimpleStructure(pageID string) string {
	name := fmt.Sprintf("%s.page.structure.txt", pageID)
	return filepath.Join(cacheDir, name)
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

func eventObserver(ev interface{}) {
	switch v := ev.(type) {
	case caching_downloader.EventError:
		log(v.Error)
	case caching_downloader.EventDidDownload:
		log("'%s' : downloaded in %s\n", v.PageID, v.Duration)
	case caching_downloader.EventDidReadFromCache:
		// TODO: only verbose
		log("'%s' : read from cache in %s\n", v.PageID, v.Duration)
	case caching_downloader.EventGotVersions:
		log("downloaded info about %d versions in %s\n", v.Count, v.Duration)
	}
}
func downloadPage(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	cache, err := caching_downloader.NewDirectoryCache(cacheDir)
	if err != nil {
		return nil, err
	}
	d := caching_downloader.New(cache, client)
	if err != nil {
		return nil, err
	}
	d.EventObserver = eventObserver
	d.NoReadCache = flgNoCache
	return d.DownloadPage(pageID)
}
