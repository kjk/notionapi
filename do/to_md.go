package main

import (
	"fmt"
	"path/filepath"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tomarkdown"
	"github.com/kjk/u"
)

func mdPath(pageID string, n int) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.%d.page.html", pageID, n)
	return filepath.Join(cacheDir, name)
}

func toMd(pageID string) *notionapi.Page {
	client := makeNotionClient()
	page, err := downloadPage(client, pageID)
	if err != nil {
		logf("toHTML: downloadPage() failed with '%s'\n", err)
		return nil
	}
	if page == nil {
		logf("toHTML: page is nil\n")
		return nil
	}

	notionapi.PanicOnFailures = true

	c := tomarkdown.NewConverter(page)
	md := c.ToMarkdown()
	path := htmlPath(pageID, 2)
	u.WriteFileMust(path, md)
	logf("%s : MarkDown version of the page\n", path)
	if !flgNoOpen {
		path, err := filepath.Abs(path)
		must(err)
		uri := "file://" + path
		logf("Opening browser with %s\n", uri)
		u.OpenBrowser(uri)
	}
	return page
}
