package main

import (
	"bytes"

	"github.com/kjk/caching_http_client"
	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
	"github.com/kjk/notionapi/tomarkdown"
)

func pageToHTML(page *notionapi.Page) []byte {
	converter := tohtml.NewConverter(page)
	d, _ := converter.ToHTML()
	return d
}

func pageToMarkdown(page *notionapi.Page) []byte {
	converter := tomarkdown.NewConverter(page)
	d := converter.ToMarkdown()
	return d
}

func testCachingDownloads(pageID string) {
	// Test that caching downloader works:
	// - download page using empty cache
	// - format as html and md
	// - download again using cache from previous download
	// - format as html and md
	// - compare they are identical
	logf("testCachingDownloads: '%s'\n", pageID)
	cache := caching_http_client.NewCache()
	cachingClient := caching_http_client.New(cache)
	client := &notionapi.Client{
		DebugLog: flgVerbose,
		//Logger:     os.Stdout,
		AuthToken:  getToken(),
		HTTPClient: cachingClient,
	}

	pageID = notionapi.ToNoDashID(pageID)
	page1, err := client.DownloadPage(pageID)
	must(err)
	html := pageToHTML(page1)
	md := pageToMarkdown(page1)

	nRequests := cache.RequestsFromCache
	cache.RequestsFromCache = 0
	cache.RequestsNotFromCache = 0

	// this should satisfy downloads using a cache
	page2, err := client.DownloadPage(pageID)
	must(err)

	// verify we made the same amount of requests
	panicIf(nRequests != cache.RequestsNotFromCache, "nRequests: %d, cache.RequestsNotFromCache: %d", nRequests, cache.RequestsNotFromCache)

	html2 := pageToHTML(page2)
	md_2 := pageToMarkdown(page2)

	if !bytes.Equal(html, html2) {
		logf("html != html2!\n")
		return
	}

	if !bytes.Equal(md, md_2) {
		logf("md != md_2!\n")
		return
	}

	//logf("json:\n%s\n", string(d))
}
