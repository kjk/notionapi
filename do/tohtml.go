package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
	"github.com/kjk/notionapi/tohtml2"
)

func htmlPath(pageID string, n int) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.%d.page.html", pageID, n)
	return filepath.Join(cacheDir, name)
}

func toHTML(pageID string) {

	client := makeNotionClient()
	page, err := downloadPage(client, pageID)
	must(err)
	if page == nil {
		return
	}

	notionapi.PanicOnFailures = true

	{
		c := tohtml.NewConverter(page)
		c.AddHeaderAnchor = true
		html := c.ToHTML()

		html = makeFullHTML(html)
		path := htmlPath(pageID, 1)
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
	{
		c := tohtml2.NewConverter(page)
		html, _ := c.ToHTML()
		path := htmlPath(pageID, 2)
		err = ioutil.WriteFile(path, html, 0644)
		must(err)
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
}
