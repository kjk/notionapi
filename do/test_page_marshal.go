package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml2"
)

func pageToHTML(page *notionapi.Page) []byte {
	converter := tohtml2.NewConverter(page)
	d := converter.ToHTML()
	return d
}

func testPageMarshal() {
	// Test that we marshal Page object correctly:
	// - download page
	// - format as html
	// - marshal and unmarshal from json
	// - format as html
	// - compare html is identical
	client := &notionapi.Client{
		DebugLog: true,
	}

	pageID := "c969c9455d7c4dd79c7f860f3ace6429"
	page1, err := client.DownloadPage(pageID)
	must(err)
	html1 := pageToHTML(page1)

	d, err := json.MarshalIndent(page1, "", "  ")
	must(err)
	var page2 notionapi.Page
	err = json.Unmarshal(d, &page2)
	must(err)
	html2 := pageToHTML(&page2)
	if bytes.Equal(html1, html2) {
		fmt.Printf("testPageMarshal() ok!\n")
		return
	}
	fmt.Printf("testPageMarshal() failed. html1 != html2. len(html1): %d, len(html2): %d!\n", len(html1), len(html2))

	//fmt.Printf("json:\n%s\n", string(d))
}
