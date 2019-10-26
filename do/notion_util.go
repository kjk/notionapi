package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kjk/notionapi/caching_downloader"

	"github.com/kjk/notionapi"
)

var (
	didPrintTokenStatus bool
)

func makeNotionClient() *notionapi.Client {
	client := &notionapi.Client{
		AuthToken: getToken(),
		DebugLog:  flgVerbose,
		Logger:    os.Stdout,
	}

	if !didPrintTokenStatus {
		didPrintTokenStatus = true
		if client.AuthToken == "" {
			logf("NOTION_TOKEN env variable not set. Can only access public pages\n")
		} else {
			// TODO: validate that the token looks legit
			logf("NOTION_TOKEN env variable set, can access private pages\n")
		}
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

var (
	eventsPerID = map[string]string{}
)

func eventObserver(ev interface{}) {
	switch v := ev.(type) {
	case *caching_downloader.EventError:
		logf(v.Error)
	case *caching_downloader.EventDidDownload:
		s := fmt.Sprintf("downloaded in %s", v.Duration)
		eventsPerID[v.PageID] = s
	case *caching_downloader.EventDidReadFromCache:
		s := fmt.Sprintf("from cache in %s", v.Duration)
		eventsPerID[v.PageID] = s
	case *caching_downloader.EventGotVersions:
		logf("downloaded info about %d versions in %s\n", v.Count, v.Duration)
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

const (
	idNoDashLength = 32
)

// only hex chars seem to be valid
func isValidNoDashIDChar(c byte) bool {
	switch {
	case c >= '0' && c <= '9':
		return true
	case c >= 'a' && c <= 'f':
		return true
	case c >= 'A' && c <= 'F':
		// currently not used but just in case notion starts using them
		return true
	}
	return false
}

// given e.g.:
// /p/foo-395f6c6af50d44e48919a45fcc064d3e
// returns:
// 395f6c6af50d44e48919a45fcc064d3e
func extractNotionIDFromURL(uri string) string {
	n := len(uri)
	if n < idNoDashLength {
		return ""
	}

	s := ""
	for i := n - 1; i > 0; i-- {
		c := uri[i]
		if c == '-' {
			continue
		}
		if isValidNoDashIDChar(c) {
			s = string(c) + s
			if len(s) == idNoDashLength {
				return s
			}
		}
	}
	return ""
}
