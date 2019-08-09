package main

import (
	"bytes"
	"fmt"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
	"github.com/kjk/notionapi/tohtml2"
	"github.com/kjk/notionapi/tomarkdown"
)

func pageToHTML1(page *notionapi.Page) []byte {
	converter := tohtml.NewConverter(page)
	d := converter.ToHTML()
	return d
}

func pageToHTML2(page *notionapi.Page) []byte {
	converter := tohtml2.NewConverter(page)
	d := converter.ToHTML()
	return d
}

func pageToMarkdown(page *notionapi.Page) []byte {
	converter := tomarkdown.NewConverter(page)
	d := converter.ToMarkdown()
	return d
}

func testPageJSONMarshal(pageID string) {
	// Test that we marshal Page object correctly:
	// - download page
	// - format as html
	// - marshal and unmarshal from json
	// - format as html
	// - compare html is identical
	cache := notionapi.NewHTTPCache()
	cachingClient := notionapi.NewCachingHTTPClient(cache)
	client := &notionapi.Client{
		DebugLog: true,
		//Logger:     os.Stdout,
		AuthToken:  getToken(),
		HTTPClient: cachingClient,
	}

	pageID = notionapi.ToNoDashID(pageID)
	page1, err := client.DownloadPage(pageID)
	must(err)
	html1 := pageToHTML1(page1)
	html2 := pageToHTML2(page1)
	md := pageToMarkdown(page1)

	nRequests := cache.RequestsFromCache
	cache.RequestsFromCache = 0
	cache.RequestsNotFromCache = 0

	// this should satisfy downloads using a cache
	page2, err := client.DownloadPage(pageID)
	must(err)

	// verify we made the same amount of requests
	panicIf(nRequests != cache.RequestsNotFromCache, "nRequests: %d, cache.RequestsNotFromCache: %d", nRequests, cache.RequestsNotFromCache)

	html1_2 := pageToHTML1(page2)
	html2_2 := pageToHTML2(page2)
	md_2 := pageToMarkdown(page2)

	if !bytes.Equal(html1, html1_2) {
		fmt.Printf("html1 != html1_2!\n")
		return
	}

	if !bytes.Equal(html2, html2_2) {
		fmt.Printf("html2 != html2_2!\n")
		return
	}

	if !bytes.Equal(md, md_2) {
		fmt.Printf("md != md_2!\n")
		return
	}

	fmt.Printf("testPageJSONMarshal() of %s ok!\n", pageID)

	//fmt.Printf("json:\n%s\n", string(d))
}
