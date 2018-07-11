package main

import (
	"fmt"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kjk/notion"
)

func openLogFileForPageID(pageID string) (io.WriteCloser, error) {
	name := fmt.Sprintf("%s.go.log.txt", pageID)
	path := filepath.Join("log", name)
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		return nil, err
	}
	notion.Logger = f
	return f, nil
}

func genHTMLTitle(f io.Writer, pageBlock *notion.Block) error {
	title := ""
	if len(pageBlock.InlineContent) > 0 {
		title = pageBlock.InlineContent[0].Text
		title = template.HTMLEscapeString(title)
	}

	s := fmt.Sprintf(`  <div class="title">%s</div>%s`, title, "\n")
	_, err := io.WriteString(f, s)
	return err
}

func genInlineBlockHTML(f io.Writer, b *notion.InlineBlock) error {
	var start, close string
	if b.AttrFlags&notion.AttrBold != 0 {
		start += "<b>"
		close += "</b>"
	}
	if b.AttrFlags&notion.AttrItalic != 0 {
		start += "<i>"
		close += "</i>"
	}
	if b.AttrFlags&notion.AttrStrikeThrought != 0 {
		start += "<strike>"
		close += "</strike>"
	}
	if b.AttrFlags&notion.AttrCode != 0 {
		start += "<code>"
		close += "</code>"
	}
	skipText := false
	for _, attrRaw := range b.Attrs {
		switch attr := attrRaw.(type) {
		case *notion.AttrLink:
			start += fmt.Sprintf(`<a href="%s">%s</a>`, attr.Link, b.Text)
			skipText = true
		case *notion.AttrUser:
			start += fmt.Sprintf(`<span class="user">@%s</span>`, attr.UserID)
			skipText = true
		case *notion.AttrDate:
			// TODO: serialize date properly
			start += fmt.Sprintf(`<span class="date">@TODO: date</span>`)
			skipText = true
		}
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

func genInlineBlocksHTML(f io.Writer, blocks []*notion.InlineBlock) error {
	for _, block := range blocks {
		err := genInlineBlockHTML(f, block)
		if err != nil {
			return err
		}
	}
	return nil
}

func genBlockSurroudedHTML(f io.Writer, block *notion.Block, start, close string, level int) error {
	_, err := io.WriteString(f, start+"\n")
	if err != nil {
		return err
	}

	err = genInlineBlocksHTML(f, block.InlineContent)
	if err != nil {
		return err
	}

	err = genBlocksHTML(f, block.Content, level+1)
	if err != nil {
		return err
	}

	_, err = io.WriteString(f, close+"\n")
	if err != nil {
		return err
	}
	return nil
}

func genBlockHTML(f io.Writer, block *notion.Block, level int) error {
	var err error
	levelCls := ""
	if level > 0 {
		levelCls = fmt.Sprintf(" lvl%d", level)
	}

	switch block.Type {
	case notion.TypeText:
		start := fmt.Sprintf(`<div class="text%s">`, levelCls)
		close := `</div>`
		err = genBlockSurroudedHTML(f, block, start, close, level)
	case notion.TypeHeader:
		start := fmt.Sprintf(`<h1 class="hdr%s">`, levelCls)
		close := `</h1>`
		err = genBlockSurroudedHTML(f, block, start, close, level)
	case notion.TypeSubHeader:
		start := fmt.Sprintf(`<h2 class="hdr%s">`, levelCls)
		close := `</h2>`
		err = genBlockSurroudedHTML(f, block, start, close, level)
	case notion.TypeTodo:
		// TODO: add checked
		clsChecked := ""
		if block.IsChecked {
			clsChecked = " is_checked"
		}
		start := fmt.Sprintf(`<div class="todo%s%s">`, levelCls, clsChecked)
		close := `</div>`
		err = genBlockSurroudedHTML(f, block, start, close, level)
	case notion.TypeToggle:
		start := fmt.Sprintf(`<div class="toggle%s">`, levelCls)
		close := `</div>`
		err = genBlockSurroudedHTML(f, block, start, close, level)
	case notion.TypeBulletedList:
		start := fmt.Sprintf(`<div class="bullet_list%s">`, levelCls)
		close := `</div>`
		err = genBlockSurroudedHTML(f, block, start, close, level)
	case notion.TypeNumberedList:
		start := fmt.Sprintf(`<div class="numbered_list%s">`, levelCls)
		close := `</div>`
		err = genBlockSurroudedHTML(f, block, start, close, level)
	case notion.TypeQuote:
		start := fmt.Sprintf(`<quote class="%s">`, levelCls)
		close := `</quote>`
		err = genBlockSurroudedHTML(f, block, start, close, level)
	case notion.TypeDivider:
		_, err = fmt.Fprintf(f, `<hr class="%s"/>`+"\n", levelCls)
	case notion.TypePage:
		id := strings.TrimSpace(block.ID)
		cls := "page"
		if block.IsLinkToPage() {
			cls = "page_link"
		}
		title := template.HTMLEscapeString(block.Title)
		url := normalizeID(id) + ".html"
		html := fmt.Sprintf(`<div class="%s%s"><a href="%s">%s</a></div>`, cls, levelCls, url, title)
		_, err = fmt.Fprintf(f, "%s\n", html)
	case notion.TypeCode:
		start := fmt.Sprintf(`<pre class="%s">`, levelCls)
		close := `</pre>`
		err = genBlockSurroudedHTML(f, block, start, close, level)
	case notion.TypeBookmark:
		_, err = fmt.Fprintf(f, `<div class="bookmark %s">Bookmark to %s</div>`+"\n", levelCls, block.Link)
	case notion.TypeGist:
		_, err = fmt.Fprintf(f, `<div class="gist %s">Gist for %s</div>`+"\n", levelCls, block.Source)
	case notion.TypeImage:
		link := block.Source
		_, err = fmt.Fprintf(f, `<img class="%s" src="%s" />`+"\n", levelCls, link)
	case notion.TypeCollectionView:
		// TODO: implement me
	default:
		fmt.Printf("Unsupported block type '%s', id: %s\n", block.Type, block.ID)
		return fmt.Errorf("Unsupported block type '%s'", block.Type)
	}
	return err
}

// convert 2131b10c-ebf6-4938-a127-7089ff02dbe4 to 2131b10cebf64938a1277089ff02dbe4
func normalizeID(s string) string {
	return strings.Replace(s, "-", "", -1)
}

func genBlocksHTML(f io.Writer, blocks []*notion.Block, level int) error {
	for _, block := range blocks {
		err := genBlockHTML(f, block, level)
		if err != nil {
			return err
		}
	}
	return nil
}

func genHTML(pageID string, pageInfo *notion.PageInfo) error {
	path := path.Join("www", normalizeID(pageID)+".html")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	title := pageInfo.Page.Title
	title = template.HTMLEscapeString(title)
	_, err = fmt.Fprintf(f, `<!doctype html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="/main.css" rel="stylesheet">
		<title>%s</title>
	</head>
	<body>`, title)

	if err != nil {
		return err
	}

	page := pageInfo.Page
	err = genHTMLTitle(f, page)
	if err != nil {
		return err
	}

	err = genBlocksHTML(f, page.Content, 0)

	_, err = fmt.Fprintf(f, "</body>\n</html>\n")
	if err != nil {
		return err
	}
	return nil
}

func toHTML(pageID string) (*notion.PageInfo, error) {
	fmt.Printf("toHTML: pageID=%s\n", pageID)
	lf, _ := openLogFileForPageID(pageID)
	if lf != nil {
		defer lf.Close()
	}
	pageInfo, err := notion.GetPageInfo(pageID)
	if err != nil {
		fmt.Printf("GetPageInfo('%s') failed with %s\n", pageID, err)
		return nil, err
	}
	return pageInfo, genHTML(pageID, pageInfo)
}

func findSubPageIDs(blocks []*notion.Block) []string {
	var res []string
	for _, block := range blocks {
		if block.Type == notion.TypePage {
			res = append(res, block.ID)
		}
	}
	return res
}

var (
	recursive = false
)

// https://www.notion.so/kjkpublic/Test-page-c969c9455d7c4dd79c7f860f3ace6429
// https://www.notion.so/kjkpublic/Test-page-text-4c6a54c68b3e4ea2af9cfaabcc88d58d
// https://www.notion.so/kjkpublic/Test-page-text-not-simple-f97ffca91f8949b48004999df34ab1f7
// https://www.notion.so/kjkpublic/blog-300db9dc27c84958a08b8d0c37f4cfe5
func main() {
	os.MkdirAll("log", 755)
	os.MkdirAll("cache", 755)
	toVisit := []string{
		//"f97ffca91f8949b48004999df34ab1f7", // text not simple
		//"6682351e44bb4f9ca0e149b703265bdb", // header
		//"fd9338a719a24f02993fcfbcf3d00bb0", // todo list
		//"484919a1647144c29234447ce408ff6b", // toggle and bullet list
		//"c969c9455d7c4dd79c7f860f3ace6429",
		"300db9dc27c84958a08b8d0c37f4cfe5", // large page (my blog)
		//"0367c2db381a4f8b9ce360f388a6b2e3", // index page for test pages
	}
	seen := map[string]struct{}{}
	for len(toVisit) > 0 {
		pageID := toVisit[0]
		toVisit = toVisit[1:]
		id := normalizeID(pageID)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		pageInfo, err := toHTML(id)
		if err != nil {
			fmt.Printf("toHTML('%s') failed with %s\n", id, err)
		}
		if recursive {
			subPages := findSubPageIDs(pageInfo.Page.Content)
			toVisit = append(toVisit, subPages...)
		}
	}
}
