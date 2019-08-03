package tohtml

import (
	"bytes"
	"fmt"
	"html"

	"path"
	"strings"

	"github.com/kjk/notionapi"
)

func maybePanic(format string, args ...interface{}) {
	notionapi.MaybePanic(format, args...)
}

func isSafeChar(r rune) bool {
	if r >= '0' && r <= '9' {
		return true
	}
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	return false
}

// safeName returns a file-system safe name
func safeName(s string) string {
	var res string
	for _, r := range s {
		if !isSafeChar(r) {
			res += " "
		} else {
			res += string(r)
		}
	}
	// replace multi-dash with single dash
	for strings.Contains(res, "  ") {
		res = strings.Replace(res, "  ", " ", -1)
	}
	res = strings.TrimLeft(res, " ")
	res = strings.TrimRight(res, " ")
	return res
}

func htmlFileName(title string) string {
	s := safeName(title)
	return s + ".html"
}

// HTMLFileNameForPage returns file name for html file
func HTMLFileNameForPage(page *notionapi.Page) string {
	return htmlFileName(page.Root.Title)
}
func log(format string, args ...interface{}) {
	notionapi.Log(format, args...)
}

// BlockRenderFunc is a function for rendering a particular block
type BlockRenderFunc func(block *notionapi.Block) bool

// HTMLRenderer converts a Page to HTML
type HTMLRenderer struct {
	// Buf is where HTML is being written to
	Buf  *bytes.Buffer
	Page *notionapi.Page

	// if true, adds id=${NotionID} attribute to HTML nodes
	AddIDAttribute bool

	// if true, adds <a href="#{$NotionID}">svg(anchor-icon)</a>
	// to h1/h2/h3
	AddHeaderAnchor bool

	// allows over-riding rendering of specific blocks
	// return false for default rendering
	RenderBlockOverride BlockRenderFunc

	// RewriteURL allows re-writing URLs e.g. to convert inter-notion URLs
	// to destination URLs
	RewriteURL func(url string) string

	// data provided by they caller, useful when providing
	// RenderBlockOverride
	Data interface{}

	// Level is current depth of the tree. Useuful for pretty-printing indentation
	Level int

	// we need this to properly render ordered and numbered lists
	CurrBlocks   []*notionapi.Block
	CurrBlockIdx int

	bufs []*bytes.Buffer
}

var (
	selfClosingTags = map[string]bool{
		"img": true,
	}
)

func isSelfClosing(tag string) bool {
	return selfClosingTags[tag]
}

// NewHTMLRenderer returns customizable HTML renderer
func NewHTMLRenderer(page *notionapi.Page) *HTMLRenderer {
	return &HTMLRenderer{
		Page: page,
	}
}

// PushNewBuffer creates a new buffer and sets Buf to it
func (r *HTMLRenderer) PushNewBuffer() {
	r.bufs = append(r.bufs, r.Buf)
	r.Buf = &bytes.Buffer{}
}

// PopBuffer pops a buffer
func (r *HTMLRenderer) PopBuffer() *bytes.Buffer {
	res := r.Buf
	n := len(r.bufs)
	r.Buf = r.bufs[n-1]
	r.bufs = r.bufs[:n-1]
	return res
}

// Newline writes a newline to the buffer. It'll suppress multiple newlines.
func (r *HTMLRenderer) Newline() {
	d := r.Buf.Bytes()
	n := len(d)
	if n > 0 && d[n-1] != '\n' {
		r.Buf.WriteByte('\n')
	}
}

// WriteString writes a string to the buffer
func (r *HTMLRenderer) WriteString(s string) {
	r.Buf.WriteString(s)
}

// WriteIndent writes 2 * Level spaces
func (r *HTMLRenderer) WriteIndent() {
	if r.Level < 0 {
		panic("r.Level is < 0")
	}
	for n := 0; n < r.Level; n++ {
		r.WriteString("  ")
	}
}

func (r *HTMLRenderer) maybeGetID(block *notionapi.Block) string {
	if r.AddIDAttribute {
		return notionapi.ToNoDashID(block.ID)
	}
	return ""
}

// WriteElement is a helper class that writes HTML with
// attributes and optional content
func (r *HTMLRenderer) WriteElement(block *notionapi.Block, tag string, attrs []string, content string, entering bool) {
	if !entering {
		if !isSelfClosing(tag) {
			r.WriteIndent()
			r.WriteString("</" + tag + ">")
			r.Newline()
		}
		return
	}

	s := "<" + tag
	nAttrs := len(attrs) / 2
	for i := 0; i < nAttrs; i++ {
		a := attrs[i*2]
		// TODO: quote value if necessary
		v := attrs[i*2+1]
		s += fmt.Sprintf(` %s="%s"`, a, v)
	}
	id := r.maybeGetID(block)
	if id != "" {
		s += ` id="` + id + `"`
	}
	s += ">"
	r.WriteIndent()
	r.WriteString(s)
	r.Newline()
	if len(content) > 0 {
		r.WriteIndent()
		r.WriteString(content)
		r.Newline()
	} else {
		r.RenderInlines(block.InlineContent)
	}
	r.Newline()
}

// PrevBlock is a block preceding current block
func (r *HTMLRenderer) PrevBlock() *notionapi.Block {
	if r.CurrBlockIdx == 0 {
		return nil
	}
	return r.CurrBlocks[r.CurrBlockIdx-1]
}

// NextBlock is a block preceding current block
func (r *HTMLRenderer) NextBlock() *notionapi.Block {
	nextIdx := r.CurrBlockIdx + 1
	lastIdx := len(r.CurrBlocks) - 1
	if nextIdx > lastIdx {
		return nil
	}
	return r.CurrBlocks[nextIdx]
}

// IsPrevBlockOfType returns true if previous block is of a given type
func (r *HTMLRenderer) IsPrevBlockOfType(t string) bool {
	b := r.PrevBlock()
	if b == nil {
		return false
	}
	return b.Type == t
}

// IsNextBlockOfType returns true if next block is of a given type
func (r *HTMLRenderer) IsNextBlockOfType(t string) bool {
	b := r.NextBlock()
	if b == nil {
		return false
	}
	return b.Type == t
}

// FormatDate formats the data
func (r *HTMLRenderer) FormatDate(d *notionapi.Date) string {
	// TODO: allow over-riding date formatting
	s := notionapi.FormatDate(d)
	return fmt.Sprintf(`<span class="notion-date">@%s</span>`, s)
}

// RenderInline renders inline block
func (r *HTMLRenderer) RenderInline(b *notionapi.InlineBlock) {
	var start, close string
	if b.AttrFlags&notionapi.AttrBold != 0 {
		start += `<b>`
		close += `</b>`
	}
	if b.AttrFlags&notionapi.AttrItalic != 0 {
		start += `<i>`
		close += `</i>`
	}
	if b.AttrFlags&notionapi.AttrStrikeThrought != 0 {
		start += `<strike>`
		close += `</strike>`
	}
	if b.AttrFlags&notionapi.AttrCode != 0 {
		start += `<code class="notion-code-inline">`
		close += `</code>`
	}
	skipText := false
	// TODO: colors
	if b.Link != "" {
		uri := b.Link
		if r.RewriteURL != nil {
			uri = r.RewriteURL(uri)
		}
		text := html.EscapeString(b.Text)
		s := fmt.Sprintf(`<a class="notion-link" href="%s">%s</a>`, uri, text)
		start += s
		skipText = true
	}
	if b.UserID != "" {
		start += fmt.Sprintf(`<span class="notion-user">@TODO: user with id%s</span>`, b.UserID)
		skipText = true
	}
	if b.Date != nil {
		start += r.FormatDate(b.Date)
		skipText = true
	}
	if !skipText {
		start += html.EscapeString(b.Text)
	}
	r.WriteString(start + close)
}

// RenderInlines renders inline blocks
func (r *HTMLRenderer) RenderInlines(blocks []*notionapi.InlineBlock) {
	r.Level++
	r.WriteIndent()
	for _, block := range blocks {
		r.RenderInline(block)
	}

	r.Level--
}

// GetInlineContent is like RenderInlines but instead of writing to
// output buffer, we return it as string
func (r *HTMLRenderer) GetInlineContent(blocks []*notionapi.InlineBlock) string {
	if len(blocks) == 0 {
		return ""
	}
	r.PushNewBuffer()
	for _, block := range blocks {
		r.RenderInline(block)
	}
	return r.PopBuffer().String()
}

// RenderCode renders BlockCode
func (r *HTMLRenderer) RenderCode(block *notionapi.Block) {
	cls := "notion-code"
	lang := strings.ToLower(strings.TrimSpace(block.CodeLanguage))
	if lang != "" {
		cls += " notion-lang-" + lang
	}
	code := html.EscapeString(block.Code)
	s := fmt.Sprintf(`<pre class="%s"><code>%s`, cls, code)
	r.WriteString(s)

	r.WriteString("</code></pre>")
	r.Newline()
}

// RenderPage renders BlockPage
func (r *HTMLRenderer) RenderPage(block *notionapi.Block) {
	tp := block.GetPageType()
	if tp == notionapi.BlockPageTopLevel {
		title := html.EscapeString(block.Title)
		content := fmt.Sprintf(`<div class="notion-page-content">%s</div>`, title)
		attrs := []string{"class", "notion-page"}
		r.WriteElement(block, "div", attrs, content, true)
		r.RenderChildren(block)
		r.WriteElement(block, "div", attrs, content, false)
		return
	}

	cls := "notion-page-link"
	if tp == notionapi.BlockPageSubPage {
		cls = "notion-sub-page"
	}
	id := notionapi.ToNoDashID(block.ID)
	uri := "https://notion.so/" + id
	title := html.EscapeString(block.Title)
	s := fmt.Sprintf(`<div class="%s"><a href="%s">%s</a></div>`, cls, uri, title)
	r.WriteIndent()
	r.WriteString(s)
	r.Newline()
}

// RenderText renders BlockText
func (r *HTMLRenderer) RenderText(block *notionapi.Block) {
	attrs := []string{"class", "notion-text"}
	r.WriteElement(block, "div", attrs, "", true)
	r.RenderChildren(block)
	r.WriteElement(block, "div", attrs, "", false)
}

// RenderNumberedList renders BlockNumberedList
func (r *HTMLRenderer) RenderNumberedList(block *notionapi.Block) {
	isPrevSame := r.IsPrevBlockOfType(notionapi.BlockNumberedList)
	if !isPrevSame {
		r.WriteIndent()
		r.WriteString(`<ol class="notion-numbered-list">`)
	}
	attrs := []string{"class", "notion-numbered-list"}
	r.WriteElement(block, "li", attrs, "", true)

	r.RenderChildren(block)

	r.WriteIndent()
	r.WriteString(`</li>`)
	isNextSame := r.IsNextBlockOfType(notionapi.BlockNumberedList)
	if !isNextSame {
		r.WriteIndent()
		r.WriteString(`</ol>`)
	}
	r.Newline()
}

// RenderBulletedList renders BlockBulletedList
func (r *HTMLRenderer) RenderBulletedList(block *notionapi.Block) {

	isPrevSame := r.IsPrevBlockOfType(notionapi.BlockBulletedList)
	if !isPrevSame {
		r.WriteIndent()
		r.WriteString(`<ul class="notion-bulleted-list">`)
		r.Newline()
		r.Level++
	}
	attrs := []string{"class", "notion-bulleted-list"}
	r.WriteElement(block, "li", attrs, "", true)

	r.RenderChildren(block)

	r.WriteIndent()
	r.WriteString(`</li>`)
	isNextSame := r.IsNextBlockOfType(notionapi.BlockBulletedList)
	if !isNextSame {
		r.Level--
		r.Newline()
		r.WriteIndent()
		r.WriteString(`</ul>`)
	}
	r.Newline()
}

// RenderHeaderLevel renders BlockHeader, SubHeader and SubSubHeader
func (r *HTMLRenderer) RenderHeaderLevel(block *notionapi.Block, level int) {
	el := fmt.Sprintf("h%d", level)
	cls := fmt.Sprintf("notion-header-%d", level)
	attrs := []string{"class", cls}
	content := r.GetInlineContent(block.InlineContent)
	id := r.maybeGetID(block)
	if r.AddHeaderAnchor {
		content += fmt.Sprintf(` <a class="notion-header-anchor" href="#%s" aria-hidden="true"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 8 8"><path d="M5.88.03c-.18.01-.36.03-.53.09-.27.1-.53.25-.75.47a.5.5 0 1 0 .69.69c.11-.11.24-.17.38-.22.35-.12.78-.07 1.06.22.39.39.39 1.04 0 1.44l-1.5 1.5c-.44.44-.8.48-1.06.47-.26-.01-.41-.13-.41-.13a.5.5 0 1 0-.5.88s.34.22.84.25c.5.03 1.2-.16 1.81-.78l1.5-1.5c.78-.78.78-2.04 0-2.81-.28-.28-.61-.45-.97-.53-.18-.04-.38-.04-.56-.03zm-2 2.31c-.5-.02-1.19.15-1.78.75l-1.5 1.5c-.78.78-.78 2.04 0 2.81.56.56 1.36.72 2.06.47.27-.1.53-.25.75-.47a.5.5 0 1 0-.69-.69c-.11.11-.24.17-.38.22-.35.12-.78.07-1.06-.22-.39-.39-.39-1.04 0-1.44l1.5-1.5c.4-.4.75-.45 1.03-.44.28.01.47.09.47.09a.5.5 0 1 0 .44-.88s-.34-.2-.84-.22z"></path></svg></a>`, id)
	}
	r.WriteElement(block, el, attrs, content, true)
	r.WriteElement(block, el, attrs, "", false)
}

// RenderHeader renders BlockHeader
func (r *HTMLRenderer) RenderHeader(block *notionapi.Block) {
	r.RenderHeaderLevel(block, 1)
}

// RenderSubHeader renders BlockSubHeader
func (r *HTMLRenderer) RenderSubHeader(block *notionapi.Block) {
	r.RenderHeaderLevel(block, 2)
}

// RenderSubSubHeader renders BlocSubSubkHeader
func (r *HTMLRenderer) RenderSubSubHeader(block *notionapi.Block) {
	r.RenderHeaderLevel(block, 3)
}

// RenderTodo renders BlockTodo
func (r *HTMLRenderer) RenderTodo(block *notionapi.Block) {
	cls := "notion-todo"
	if block.IsChecked {
		cls = "notion-todo-checked"
	}
	attrs := []string{"class", cls}
	r.WriteElement(block, "div", attrs, "", true)

	r.RenderChildren(block)

	r.WriteElement(block, "div", attrs, "", false)
}

// RenderToggle renders BlockToggle
func (r *HTMLRenderer) RenderToggle(block *notionapi.Block) {
	s := `<details class="notion-toggle"`
	id := r.maybeGetID(block)
	if id != "" {
		s += fmt.Sprintf(` id="%s"`, id)
	}
	r.WriteString(s + `>`)
	r.Newline()

	// we don't want id on summary but on <details> above
	prevAddID := r.AddIDAttribute
	r.AddIDAttribute = false
	r.WriteElement(block, "summary", nil, "", true)
	r.WriteString(`</summary>`)
	r.AddIDAttribute = prevAddID

	r.Newline()

	r.RenderChildren(block)

	r.WriteString("</details>\n")
}

// RenderQuote renders BlockQuote
func (r *HTMLRenderer) RenderQuote(block *notionapi.Block) {
	cls := "notion-quote"
	attrs := []string{"class", cls}
	r.WriteElement(block, "quote", attrs, "", true)

	r.RenderChildren(block)

	r.WriteElement(block, "quote", attrs, "", false)
}

// RenderCallout renders BlockCallout
func (r *HTMLRenderer) RenderCallout(block *notionapi.Block) {
	cls := "notion-callout"
	attrs := []string{"class", cls}
	r.WriteElement(block, "div", attrs, "", true)

	r.RenderChildren(block)
	r.WriteElement(block, "div", attrs, "", false)
}

// RenderTableOfContents renders BlockTableOfContents
func (r *HTMLRenderer) RenderTableOfContents(block *notionapi.Block) {
	// TODO: implement me
}

// RenderDivider renders BlockDivider
func (r *HTMLRenderer) RenderDivider(block *notionapi.Block) {
	r.WriteString(`<hr class="notion-divider">` + "\n")
}

// RenderBookmark renders BlockBookmark
func (r *HTMLRenderer) RenderBookmark(block *notionapi.Block) {
	content := fmt.Sprintf(`<a href="%s">%s</a>`, block.Link, block.Link)
	cls := "notion-bookmark"
	// TODO: don't render inlines (which seems to be title of the bookmarked page)
	// TODO: support caption
	// TODO: maybe support comments
	attrs := []string{"class", cls}
	r.WriteElement(block, "div", attrs, content, true)
	r.WriteElement(block, "div", attrs, content, false)
}

// RenderVideo renders BlockTweet
func (r *HTMLRenderer) RenderVideo(block *notionapi.Block) {
	f := block.FormatVideo
	ws := fmt.Sprintf("%d", f.BlockWidth)
	uri := f.DisplaySource
	if uri == "" {
		// TODO: not sure if this is needed
		uri = block.Source
	}
	// TODO: get more info from format
	attrs := []string{
		"class", "notion-video",
		"width", ws,
		"src", uri,
		"frameborder", "0",
		"allow", "encrypted-media",
		"allowfullscreen", "true",
	}
	// TODO: can it be that f.BlockWidth is 0 and we need to
	// calculate it from f.BlockHeight
	h := f.BlockHeight
	if h == 0 {
		h = int64(float64(f.BlockWidth) * f.BlockAspectRatio)
	}
	if h > 0 {
		hs := fmt.Sprintf("%d", h)
		attrs = append(attrs, "height", hs)
	}

	r.WriteElement(block, "iframe", attrs, "", true)
	r.WriteElement(block, "iframe", attrs, "", false)
}

// RenderTweet renders BlockTweet
func (r *HTMLRenderer) RenderTweet(block *notionapi.Block) {
	uri := block.Source
	content := fmt.Sprintf(`Embedded tweet <a href="%s">%s</a>`, uri, uri)
	cls := "notion-embed"
	// TODO: don't render inlines (which seems to be title of the bookmarked page)
	// TODO: support caption
	// TODO: maybe support comments
	attrs := []string{"class", cls}
	r.WriteElement(block, "div", attrs, content, true)
	r.WriteElement(block, "div", attrs, content, false)
}

// RenderGist renders BlockGist
func (r *HTMLRenderer) RenderGist(block *notionapi.Block) {
	uri := block.Source + ".js"
	cls := "notion-embed-gist"
	attrs := []string{"src", uri, "class", cls}
	// TODO: support caption
	// TODO: maybe support comments
	r.WriteElement(block, "script", attrs, "", true)
	r.WriteElement(block, "script", attrs, "", false)
}

// RenderEmbed renders BlockEmbed
func (r *HTMLRenderer) RenderEmbed(block *notionapi.Block) {
	// TODO: best effort at making the URL readable
	uri := block.FormatEmbed.DisplaySource
	title := block.Title
	if title == "" {
		title = path.Base(uri)
	}
	title = html.EscapeString(title)
	content := fmt.Sprintf(`Oembed: <a href="%s">%s</a>`, uri, title)
	cls := "notion-embed"
	attrs := []string{"class", cls}
	r.WriteElement(block, "div", attrs, content, true)
	r.WriteElement(block, "div", attrs, content, false)
}

// RenderFile renders BlockFile
func (r *HTMLRenderer) RenderFile(block *notionapi.Block) {
	// TODO: best effort at making the URL readable
	uri := block.Source
	title := block.Title
	if title == "" {
		title = path.Base(uri)
	}
	title = html.EscapeString(title)
	content := fmt.Sprintf(`Embedded file: <a href="%s">%s</a>`, uri, title)
	cls := "notion-embed"
	attrs := []string{"class", cls}
	r.WriteElement(block, "div", attrs, content, true)
	r.WriteElement(block, "div", attrs, content, false)
}

// RenderPDF renders BlockPDF
func (r *HTMLRenderer) RenderPDF(block *notionapi.Block) {
	// TODO: best effort at making the URL readable
	uri := block.Source
	title := block.Title
	if title == "" {
		title = path.Base(uri)
	}
	title = html.EscapeString(title)
	content := fmt.Sprintf(`Embedded PDF: <a href="%s">%s</a>`, uri, title)
	cls := "notion-embed"
	attrs := []string{"class", cls}
	r.WriteElement(block, "div", attrs, content, true)
	r.WriteElement(block, "div", attrs, content, false)
}

// RenderImage renders BlockImage
func (r *HTMLRenderer) RenderImage(block *notionapi.Block) {
	link := block.ImageURL
	attrs := []string{"class", "notion-image", "src", link}
	r.WriteElement(block, "img", attrs, "", true)
	r.WriteElement(block, "img", attrs, "", false)
}

// RenderColumnList renders BlockColumnList
// it's children are BlockColumn
func (r *HTMLRenderer) RenderColumnList(block *notionapi.Block) {
	nColumns := len(block.Content)
	if nColumns == 0 {
		maybePanic("has no columns")
		return
	}
	attrs := []string{"class", "notion-column-list"}
	r.WriteElement(block, "div", attrs, "", true)
	r.RenderChildren(block)
	r.WriteElement(block, "div", attrs, "", false)
}

// RenderColumn renders BlockColumn
// it's parent is BlockColumnList
func (r *HTMLRenderer) RenderColumn(block *notionapi.Block) {
	// TODO: get column ration from col.FormatColumn.ColumnRation, which is float 0...1
	attrs := []string{"class", "notion-column"}
	r.WriteElement(block, "div", attrs, "", true)
	r.RenderChildren(block)
	r.WriteElement(block, "div", attrs, "", false)
}

// RenderCollectionView renders BlockCollectionView
func (r *HTMLRenderer) RenderCollectionView(block *notionapi.Block) {
	viewInfo := block.CollectionViews[0]
	view := viewInfo.CollectionView
	if view.Format == nil {
		id := ""
		if r.Page != nil {
			id = notionapi.ToNoDashID(r.Page.ID)
		}
		log("missing view.Format for block %s %s in page %s\n", block.ID, block.Type, id)
		return
	}
	columns := view.Format.TableProperties

	r.Newline()
	r.WriteIndent()
	r.WriteString("\n" + `<table class="notion-collection-view">` + "\n")

	// generate header row
	r.Level++
	r.WriteIndent()
	r.WriteString("<thead>\n")

	r.Level++
	r.WriteIndent()
	r.WriteString("<tr>\n")

	for _, col := range columns {
		colName := col.Property
		colInfo := viewInfo.Collection.CollectionSchema[colName]
		if colInfo != nil {
			name := colInfo.Name
			r.Level++
			r.WriteIndent()
			r.WriteString(`<th>` + html.EscapeString(name) + "</th>\n")
			r.Level--
		} else {
			r.Level++
			r.WriteIndent()
			r.WriteString(`<th>&nbsp;` + "</th>\n")
			r.Level--
		}
	}
	r.WriteIndent()
	r.WriteString("</tr>\n")

	r.Level--
	r.WriteIndent()
	r.WriteString("</thead>\n\n")

	r.WriteIndent()
	r.WriteString("<tbody>\n")

	for _, row := range viewInfo.CollectionRows {
		r.Level++
		r.WriteIndent()
		r.WriteString("<tr>\n")

		props := row.Properties
		for _, col := range columns {
			colName := col.Property
			v := props[colName]
			//fmt.Printf("inline: '%s'\n", fmt.Sprintf("%v", v))
			inlineContent, err := notionapi.ParseInlineBlocks(v)
			if err != nil {
				maybePanic("ParseInlineBlocks of '%v' failed with %s\n", v, err)
			}
			//pretty.Print(inlineContent)
			colVal := r.GetInlineContent(inlineContent)
			//fmt.Printf("colVal: '%s'\n", colVal)
			r.Level++
			r.WriteIndent()
			//colInfo := viewInfo.Collection.CollectionSchema[colName]
			// TODO: format colVal according to colInfo
			r.WriteString(`<td>` + colVal + `</td>`)
			r.Newline()
			r.Level--
		}
		r.WriteIndent()
		r.WriteString("</tr>\n")
		r.Level--
	}

	r.WriteIndent()
	r.WriteString("</tbody>\n")

	r.Level--
	r.WriteIndent()
	r.WriteString("</table>\n")
}

// DefaultRenderFunc returns a defult rendering function for a type of
// a given block
func (r *HTMLRenderer) DefaultRenderFunc(blockType string) func(*notionapi.Block) {
	switch blockType {
	case notionapi.BlockPage:
		return r.RenderPage
	case notionapi.BlockText:
		return r.RenderText
	case notionapi.BlockNumberedList:
		return r.RenderNumberedList
	case notionapi.BlockBulletedList:
		return r.RenderBulletedList
	case notionapi.BlockHeader:
		return r.RenderHeader
	case notionapi.BlockSubHeader:
		return r.RenderSubHeader
	case notionapi.BlockSubSubHeader:
		return r.RenderSubSubHeader
	case notionapi.BlockTodo:
		return r.RenderTodo
	case notionapi.BlockToggle:
		return r.RenderToggle
	case notionapi.BlockQuote:
		return r.RenderQuote
	case notionapi.BlockDivider:
		return r.RenderDivider
	case notionapi.BlockCode:
		return r.RenderCode
	case notionapi.BlockBookmark:
		return r.RenderBookmark
	case notionapi.BlockImage:
		return r.RenderImage
	case notionapi.BlockColumnList:
		return r.RenderColumnList
	case notionapi.BlockColumn:
		return r.RenderColumn
	case notionapi.BlockCollectionView:
		return r.RenderCollectionView
	case notionapi.BlockEmbed:
		return r.RenderEmbed
	case notionapi.BlockGist:
		return r.RenderGist
	case notionapi.BlockTweet:
		return r.RenderTweet
	case notionapi.BlockVideo:
		return r.RenderVideo
	case notionapi.BlockFile:
		return r.RenderFile
	case notionapi.BlockPDF:
		return r.RenderPDF
	case notionapi.BlockCallout:
		return r.RenderCallout
	case notionapi.BlockTableOfContents:
		return r.RenderTableOfContents
	default:
		maybePanic("DefaultRenderFunc: unsupported block type '%s' in %s\n", blockType, r.Page.NotionURL())
	}
	return nil
}

func needsWrapper(block *notionapi.Block) bool {
	if len(block.Content) == 0 {
		return false
	}
	switch block.Type {
	// TODO: maybe more block types need this
	case notionapi.BlockText:
		return true
	}
	return false
}

func (r *HTMLRenderer) RenderChildren(block *notionapi.Block) {
	if len(block.Content) == 0 {
		return
	}

	// .notion-wrap provides indentation for children
	if needsWrapper(block) {
		r.Newline()
		r.WriteIndent()
		r.WriteString(`<div class="notion-wrap">`)
		r.Newline()
	}

	r.Level++
	currIdx := r.CurrBlockIdx
	currBlocks := r.CurrBlocks
	r.CurrBlocks = block.Content
	for i, child := range block.Content {
		child.Parent = block
		r.CurrBlockIdx = i
		r.RenderBlock(child)
	}
	r.CurrBlockIdx = currIdx
	r.CurrBlocks = currBlocks
	r.Level--

	if needsWrapper(block) {
		r.Newline()
		r.WriteIndent()
		r.WriteString(`</div>`)
		r.Newline()
	}
}

// RenderBlock renders a block to html
func (r *HTMLRenderer) RenderBlock(block *notionapi.Block) {
	if block == nil {
		// a missing block
		return
	}
	if r.RenderBlockOverride != nil {
		handled := r.RenderBlockOverride(block)
		if handled {
			return
		}
	}
	def := r.DefaultRenderFunc(block.Type)
	if def != nil {
		def(block)
	}
}

// ToHTML renders a page to html
func (r *HTMLRenderer) ToHTML() []byte {
	r.Level = 0
	r.PushNewBuffer()

	r.RenderBlock(r.Page.Root)
	buf := r.PopBuffer()
	if r.Level != 0 {
		panic(fmt.Sprintf("r.Level is %d, should be 0", r.Level))
	}
	return buf.Bytes()
}

// ToHTML converts a page to HTML
func ToHTML(page *notionapi.Page) []byte {
	r := NewHTMLRenderer(page)
	return r.ToHTML()
}
