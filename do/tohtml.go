package main

import (
	"io/ioutil"
	"path/filepath"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
)

func toHTML(pageID string) {
	client := makeNotionClient()

	page, err := downloadPageCached(client, pageID)
	must(err)
	if page == nil {
		return
	}

	r := tohtml.NewConverter(page)
	r.AddIDAttribute = true
	r.AddHeaderAnchor = true
	notionapi.PanicOnFailures = true
	html := r.ToHTML()

	html = makeFullHTML(html)
	path := htmlFilePathForPageID(pageID)
	err = ioutil.WriteFile(path, html, 0644)
	must(err)
	log("%s : HTML version of the page\n", path)
	if !flgNoOpen {
		path, err := filepath.Abs(path)
		must(err)
		uri := "file://" + path
		log("Opening browser with %s\n", uri)
		openBrowser(uri)
	}
}
