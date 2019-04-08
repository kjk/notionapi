package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjk/notionapi"
)

func openLogFileForPageID(pageID string) (io.WriteCloser, error) {
	name := fmt.Sprintf("%s.go.log.txt", pageID)
	path := filepath.Join("log", name)
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		return nil, err
	}
	return f, nil
}

func genHTMLTitle(f io.Writer, pageBlock *notionapi.Block) {
	title := ""
	if len(pageBlock.InlineContent) > 0 {
		title = pageBlock.InlineContent[0].Text
		title = template.HTMLEscapeString(title)
	}

	s := fmt.Sprintf(`  <div class="title">%s</div>%s`, title, "\n")
	io.WriteString(f, s)
}

func genInlineBlockHTML(f io.Writer, b *notionapi.InlineBlock) error {
	var start, close string
	if b.AttrFlags&notionapi.AttrBold != 0 {
		start += "<b>"
		close += "</b>"
	}
	if b.AttrFlags&notionapi.AttrItalic != 0 {
		start += "<i>"
		close += "</i>"
	}
	if b.AttrFlags&notionapi.AttrStrikeThrought != 0 {
		start += "<strike>"
		close += "</strike>"
	}
	if b.AttrFlags&notionapi.AttrCode != 0 {
		start += "<code>"
		close += "</code>"
	}
	skipText := false
	if b.Link != "" {
		start += fmt.Sprintf(`<a href="%s">%s</a>`, b.Link, b.Text)
		skipText = true
	}
	if b.UserID != "" {
		start += fmt.Sprintf(`<span class="user">@%s</span>`, b.UserID)
		skipText = true
	}
	if b.Date != nil {
		// TODO: serialize date properly
		start += fmt.Sprintf(`<span class="date">@TODO: date</span>`)
		skipText = true
	}
	if !skipText {
		start += b.Text
	}
	_, err := io.WriteString(f, start+close)
	if err != nil {
		return err
	}
	return nil
}

func genInlineBlocksHTML(f io.Writer, blocks []*notionapi.InlineBlock) error {
	for _, block := range blocks {
		err := genInlineBlockHTML(f, block)
		if err != nil {
			return err
		}
	}
	return nil
}

func genBlockSurroudedHTML(f io.Writer, block *notionapi.Block, start, close string, level int) {
	io.WriteString(f, start+"\n")
	genInlineBlocksHTML(f, block.InlineContent)
	genBlocksHTML(f, block, level+1)
	io.WriteString(f, close+"\n")
}

func genBlockHTML(f io.Writer, block *notionapi.Block, level int) {
	levelCls := ""
	if level > 0 {
		levelCls = fmt.Sprintf(" lvl%d", level)
	}

	switch block.Type {
	case notionapi.BlockText:
		start := fmt.Sprintf(`<div class="text%s">`, levelCls)
		close := `</div>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.BlockHeader:
		start := fmt.Sprintf(`<h1 class="hdr%s">`, levelCls)
		close := `</h1>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.BlockSubHeader:
		start := fmt.Sprintf(`<h2 class="hdr%s">`, levelCls)
		close := `</h2>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.BlockToggle:
		start := fmt.Sprintf(`<div class="toggle%s">`, levelCls)
		close := `</div>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.BlockBulletedList:
		start := fmt.Sprintf(`<div class="bullet-list%s">`, levelCls)
		close := `</div>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.BlockNumberedList:
		start := fmt.Sprintf(`<div class="numbered-list%s">`, levelCls)
		close := `</div>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.BlockQuote:
		start := fmt.Sprintf(`<quote class="%s">`, levelCls)
		close := `</quote>`
		genBlockSurroudedHTML(f, block, start, close, level)
	case notionapi.BlockDivider:
		fmt.Fprintf(f, `<hr class="%s"/>`+"\n", levelCls)
	case notionapi.BlockPage:
		id := strings.TrimSpace(block.ID)
		cls := "page"
		if block.IsLinkToPage() {
			cls = "page-link"
		}
		title := template.HTMLEscapeString(block.Title)
		url, _ := notionapi.NormalizeID(id)
		url += ".html"
		html := fmt.Sprintf(`<div class="%s%s"><a href="%s">%s</a></div>`, cls, levelCls, url, title)
		fmt.Fprintf(f, "%s\n", html)
	case notionapi.BlockCode:
		code := template.HTMLEscapeString(block.Code)
		fmt.Fprintf(f, `<div class="%s">Lang for code: %s</div>
<pre class="%s">
%s
</pre>`, levelCls, block.CodeLanguage, levelCls, code)
	case notionapi.BlockBookmark:
		fmt.Fprintf(f, `<div class="bookmark %s">Bookmark to %s</div>`+"\n", levelCls, block.Link)
	case notionapi.BlockGist:
		fmt.Fprintf(f, `<div class="gist %s">Gist for %s</div>`+"\n", levelCls, block.Source)
	case notionapi.BlockImage:
		link := block.ImageURL
		fmt.Fprintf(f, `<img class="%s" src="%s" />`+"\n", levelCls, link)
	case notionapi.BlockColumnList:
		// TODO: implement me
	case notionapi.BlockCollectionView:
		// TODO: implement me
	default:
		fmt.Printf("Unsupported block type '%s', id: %s\n", block.Type, block.ID)
		panic(fmt.Sprintf("Unsupported block type '%s'", block.Type))
	}
}

func genBlocksHTML(f io.Writer, parent *notionapi.Block, level int) {
	blocks := parent.Content
	for i, block := range blocks {
		if block == nil {
			id := parent.ContentIDs[i]
			fmt.Printf("No block at index %d with id=%s. Parent block %s of type %s\n", i, id, parent.ID, parent.Type)
		}
		genBlockHTML(f, block, level)
	}
}

func genHTML(pageID string, page *notionapi.Page) []byte {
	f := &bytes.Buffer{}
	title := page.Root.Title
	title = template.HTMLEscapeString(title)
	fmt.Fprintf(f, `<!doctype html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="/main.css" rel="stylesheet">
		<title>%s</title>
	</head>
	<body>`, title)

	root := page.Root
	genHTMLTitle(f, root)
	genBlocksHTML(f, root, 0)
	fmt.Fprintf(f, "</body>\n</html>\n")
	return f.Bytes()
}

func downloadPageCached(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	var page notionapi.Page
	cachedPath := filepath.Join("cache", pageID+".json")
	if useCache {
		d, err := ioutil.ReadFile(cachedPath)
		if err == nil {
			err = json.Unmarshal(d, &page)
			if err == nil {
				fmt.Printf("Got data for pageID %s from cache file %s\n", pageID, cachedPath)
				return &page, nil
			}
			// not a fatal error, just a warning
			fmt.Printf("json.Unmarshal() on '%s' failed with %s\n", cachedPath, err)
		}
	}
	res, err := client.DownloadPage(pageID)
	if err != nil {
		return nil, err
	}
	d, err := json.MarshalIndent(res, "", "  ")
	if err == nil {
		err = ioutil.WriteFile(cachedPath, d, 0644)
		if err != nil {
			// not a fatal error, just a warning
			fmt.Printf("ioutil.WriteFile(%s) failed with %s\n", cachedPath, err)
		}
	} else {
		// not a fatal error, just a warning
		fmt.Printf("json.Marshal() on pageID '%s' failed with %s\n", pageID, err)
	}
	return res, nil
}

func toHTML(client *notionapi.Client, pageID, path string) (*notionapi.Page, error) {
	fmt.Printf("toHTML: pageID=%s, path=%s\n", pageID, path)
	client.Logger, _ = openLogFileForPageID(pageID)
	if client.Logger != nil {
		defer func() {
			lf := client.Logger.(*os.File)
			lf.Close()
		}()
	}
	page, err := downloadPageCached(client, pageID)
	if err != nil {
		fmt.Printf("downloadPageCached('%s') failed with %s\n", pageID, err)
		return nil, err
	}
	d := genHTML(pageID, page)
	err = ioutil.WriteFile(path, d, 0644)
	return page, err
}

func findSubPageIDs(blocks []*notionapi.Block) []string {
	var res []string
	for _, block := range blocks {
		if block.Type == notionapi.BlockPage {
			res = append(res, block.ID)
		}
	}
	return res
}

var (
	flgRecursive bool
	flgNoCache   bool
	useCache     bool // !flgNoCache, reads better
	toVisit      []string
)

func usageAndExit() {
	cmd := filepath.Base(os.Args[0])
	fmt.Printf(`Usage:
%s [-recursive] [-no-cache] urlOrId ...
urlOrId can be a Notion page URL like:
https://www.notion.so/kjkpublic/Test-page-text-4c6a54c68b3e4ea2af9cfaabcc88d58d
or a notion page id:
4c6a54c68b3e4ea2af9cfaabcc88d58d
Pages must be publicly visible.
`, cmd)
	os.Exit(1)
}

func parseCmdFlags() {
	flag.BoolVar(&flgRecursive, "recursive", false, "if true, recursively download page")
	flag.BoolVar(&flgNoCache, "no-cache", false, "if true, use cached responses, if available from previous runs")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Printf("Must provide Notion public page url or id\n")
		usageAndExit()
	}

	useCache = !flgNoCache

	// handle:
	// https://www.notion.so/kjkpublic/Test-page-c969c9455d7c4dd79c7f860f3ace6429
	// or
	// 300db9dc27c84958a08b8d0c37f4cfe5
	for _, arg := range args {
		parts := strings.Split(arg, "-")
		n := len(parts)
		id := parts[0]
		if n > 1 {
			id = parts[n-1]
		}
		id, _ = notionapi.NormalizeID(id)
		if len(id) != 36 {
			fmt.Printf("Id '%s' extracted from '%s' doesn't look like a valid Notion page id\n", id, arg)
			usageAndExit()
		}
		toVisit = append(toVisit, id)
	}
}

func copyFile(dst, src string) error {
	d, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, d, 0644)
}

func copyCSS() {
	src := filepath.Join("cmd", "tohtml", "main.css")
	dst := filepath.Join("www", "main.css")
	err := copyFile(dst, src)
	if err != nil {
		panic(err.Error())
	}
}

func main() {
	parseCmdFlags()

	os.MkdirAll("log", 0755)
	os.MkdirAll("cache", 0755)
	os.MkdirAll("www", 0755)

	client := &notionapi.Client{
		DebugLog: true,
	}
	seen := map[string]struct{}{}
	firstPage := true
	for len(toVisit) > 0 {
		pageID := toVisit[0]
		toVisit = toVisit[1:]
		id, ok := notionapi.NormalizeID(pageID)
		if !ok {
			log.Fatalf("%s is not a valid Notion page id\n", pageID)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		name := id + ".html"
		if firstPage {
			name = "index.html"
		}
		path := filepath.Join("www", name)
		page, err := toHTML(client, id, path)
		if err != nil {
			fmt.Printf("toHTML('%s') failed with %s\n", id, err)
		}
		if flgRecursive {
			subPages := findSubPageIDs(page.Root.Content)
			toVisit = append(toVisit, subPages...)
		}
		firstPage = false
	}
	copyCSS()
}
