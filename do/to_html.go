package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
)

func htmlPath(pageID string, n int) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.%d.page.html", pageID, n)
	return filepath.Join(cacheDir, name)
}

func toHTML(pageID string) {

	client := makeNotionClient()
	page, err := downloadPage(client, pageID)
	if err != nil {
		logf("toHTML: downloadPage() failed with '%s'\n", err)
		return
	}
	if page == nil {
		logf("toHTML: page is nil\n")
		return
	}

	notionapi.PanicOnFailures = true

	c := tohtml.NewConverter(page)
	c.FullHTML = true
	html, _ := c.ToHTML()
	path := htmlPath(pageID, 2)
	err = ioutil.WriteFile(path, html, 0644)
	must(err)
	must(err)
	logf("%s : HTML version of the page\n", path)
	if !flgNoOpen {
		path, err := filepath.Abs(path)
		must(err)
		uri := "file://" + path
		logf("Opening browser with %s\n", uri)
		openBrowser(uri)
	}
}
