package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

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
		start := fmt.Sprintf(`<div class="bullet%s">`, levelCls)
		close := `</div>`
		err = genBlockSurroudedHTML(f, block, start, close, level)
	default:
		fmt.Printf("Unsupported block type '%s'\n", block.Type)
		return fmt.Errorf("Unsupported block type '%s'", block.Type)
	}
	return err
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
	path := path.Join("www", pageID+".html")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, `<!doctype html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<link href="/main.css" rel="stylesheet">
	</head>
	<body>`)

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

func toHTML(pageID string) error {
	fmt.Printf("toHTML: pageID=%s\n", pageID)
	lf, _ := openLogFileForPageID(pageID)
	if lf != nil {
		defer lf.Close()
	}
	pageInfo, err := notion.GetPageInfo(pageID)
	if err != nil {
		fmt.Printf("GetPageInfo('%s') failed with %s\n", pageID, err)
		return err
	}
	return genHTML(pageID, pageInfo)
}

// https://www.notion.so/kjkpublic/Test-page-c969c9455d7c4dd79c7f860f3ace6429
// https://www.notion.so/kjkpublic/Test-page-text-4c6a54c68b3e4ea2af9cfaabcc88d58d
// https://www.notion.so/kjkpublic/Test-page-text-not-simple-f97ffca91f8949b48004999df34ab1f7
func main() {
	ids := []string{
		//"f97ffca91f8949b48004999df34ab1f7", // text not simple
		//"6682351e44bb4f9ca0e149b703265bdb", // header
		//"fd9338a719a24f02993fcfbcf3d00bb0", // todo list
		//"484919a1647144c29234447ce408ff6b", // toggle and bullet list
		"c969c9455d7c4dd79c7f860f3ace6429" //
	}
	for _, pageID := range ids {
		err := toHTML(pageID)
		if err != nil {
			fmt.Printf("toHTML() failed with %s\n", err)
		}
	}
}

/*
TODO:
* handle bullet list
*/
