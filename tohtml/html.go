package tohtml

import (
	"bytes"
	"fmt"
	"html"
	"html/template"
	"path"
	"strings"

	"github.com/kjk/notionapi"
)

// BlockRenderFunc is a function for rendering a particular
type BlockRenderFunc func(block *notionapi.Block, entering bool) bool

// HTMLRenderer converts a Page to HTML
type HTMLRenderer struct {
	// Buf is where HTML is being written to
	Buf  *bytes.Buffer
	Page *notionapi.Page

	// if true, adds id=${NotionID} attribute to HTML nodes
	AppendID bool

	// mostly for debugging. If true will panic when encounters
	// structure it cannot handle (e.g. when Notion adds another
	// type of block)
	PanicOnFailures bool

	// allows over-riding rendering of specific blocks
	// return false for default rendering
	RenderBlockOverride BlockRenderFunc

	// data provided by they caller, useful when providing
	// RenderBlockOverride
	Data interface{}

	// mostly for debugging, if set we'll log to it when encountering
	// structure we can't handle
	Log func(format string, args ...interface{})

	// Level is current depth of the tree. Useuful for pretty-printing indentation
	Level int

	// we need this to properly render ordered and numbered lists
	CurrBlocks   []*notionapi.Block
	CurrBlockIdx int

	// keeps a nesting stack of numbered / bulleted list
	// we need this because they are not nested in data model
	ListStack []string

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

// TODO: not sure if I want to keep this or always use maybePanic
// (which also logs)
func (r *HTMLRenderer) log(format string, args ...interface{}) {
	if r.Log != nil {
		r.Log(format, args...)
	}
}

func (r *HTMLRenderer) maybePanic(format string, args ...interface{}) {
	if r.Log != nil {
		r.Log(format, args...)
	}
	if r.PanicOnFailures {
		panic(fmt.Sprintf(format, args...))
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
	for n := 0; n < r.Level; n++ {
		r.WriteString("  ")
	}
}

func (r *HTMLRenderer) maybeGetID(block *notionapi.Block) string {
	if r.AppendID {
		return ""
	}
	return block.ID
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
	}
	r.RenderInlines(block.InlineContent)
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
	lastIdx := len(r.CurrBlocks) - 1
	if r.CurrBlockIdx+1 > lastIdx {
		return nil
	}
	return r.CurrBlocks[r.CurrBlockIdx+1]
}

// RenderInline renders inline block
func (r *HTMLRenderer) RenderInline(b *notionapi.InlineBlock) {
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
	// TODO: allow over-riding rendering of links, user ids, dates etc.
	if b.Link != "" {
		link := b.Link
		start += fmt.Sprintf(`<a class="notion-link" href="%s">%s</a>`, link, b.Text)
		skipText = true
	}
	if b.UserID != "" {
		start += fmt.Sprintf(`<span class="notion-user">@TODO: user with id%s</span>`, b.UserID)
		skipText = true
	}
	if b.Date != nil {
		// TODO: serialize date properly
		start += fmt.Sprintf(`<span class="notion-date">@TODO: date</span>`)
		skipText = true
	}
	if !skipText {
		start += b.Text
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

// RenderCode renders BlockCode
func (r *HTMLRenderer) RenderCode(block *notionapi.Block, entering bool) bool {
	if !entering {
		r.WriteString("</code></pre>")
		r.Newline()
		return true
	}
	cls := "notion-code"
	lang := strings.ToLower(strings.TrimSpace(block.CodeLanguage))
	if lang != "" {
		cls += " notion-lang-" + lang
	}
	code := template.HTMLEscapeString(block.Code)
	s := fmt.Sprintf(`<pre class="%s"><code>%s`, cls, code)
	r.WriteString(s)
	return true
}

// RenderPage renders BlockPage
func (r *HTMLRenderer) RenderPage(block *notionapi.Block, entering bool) bool {
	tp := block.GetPageType()
	if tp == notionapi.BlockPageTopLevel {
		title := template.HTMLEscapeString(block.Title)
		content := fmt.Sprintf(`<div class="notion-page-content">%s</div>`, title)
		attrs := []string{"class", "notion-page"}
		r.WriteElement(block, "div", attrs, content, entering)
		return true
	}

	if !entering {
		return true
	}

	cls := "notion-page-link"
	if tp == notionapi.BlockPageSubPage {
		cls = "notion-sub-page"
	}
	id := notionapi.ToNoDashID(block.ID)
	uri := "https://notion.so/" + id
	title := template.HTMLEscapeString(block.Title)
	s := fmt.Sprintf(`<div class="%s"><a href="%s">%s</a></div>`, cls, uri, title)
	r.WriteIndent()
	r.WriteString(s)
	r.Newline()
	return true
}

// RenderText renders BlockText
func (r *HTMLRenderer) RenderText(block *notionapi.Block, entering bool) bool {
	attrs := []string{"class", "notion-text"}
	r.WriteElement(block, "div", attrs, "", entering)
	return true
}

// IsPrevBlockOfType returns true if previous block is of a given type
func (r *HTMLRenderer) IsPrevBlockOfType(t string) bool {
	prev := r.PrevBlock()
	if prev == nil {
		return false
	}
	return prev.Type == t
}

// IsNextBlockOfType returns true if next block is of a given type
func (r *HTMLRenderer) IsNextBlockOfType(t string) bool {
	prev := r.NextBlock()
	if prev == nil {
		return false
	}
	return prev.Type == t
}

// RenderNumberedList renders BlockNumberedList
func (r *HTMLRenderer) RenderNumberedList(block *notionapi.Block, entering bool) bool {
	if entering {
		isPrevSame := r.IsPrevBlockOfType(notionapi.BlockNumberedList)
		if !isPrevSame {
			r.WriteIndent()
			r.WriteString(`<ol class="notion-numbered-list">`)
		}
		attrs := []string{"class", "notion-numbered-list"}
		r.WriteElement(block, "li", attrs, "", entering)
	} else {
		r.WriteIndent()
		r.WriteString(`</li>`)
		isNextSame := r.IsNextBlockOfType(notionapi.BlockNumberedList)
		if !isNextSame {
			r.WriteIndent()
			r.WriteString(`</ol>`)
		}
		r.Newline()
	}
	return true
}

// RenderBulletedList renders BlockBulletedList
func (r *HTMLRenderer) RenderBulletedList(block *notionapi.Block, entering bool) bool {

	if entering {
		isPrevSame := r.IsPrevBlockOfType(notionapi.BlockBulletedList)
		if !isPrevSame {
			r.WriteIndent()
			r.WriteString(`<ul class="notion-bulleted-list">`)
		}
		attrs := []string{"class", "notion-bulleted-list"}
		r.WriteElement(block, "li", attrs, "", entering)
	} else {
		r.WriteIndent()
		r.WriteString(`</li>`)
		isNextSame := r.IsNextBlockOfType(notionapi.BlockBulletedList)
		if !isNextSame {
			r.WriteIndent()
			r.WriteString(`</ul>`)
		}
		r.Newline()
	}
	return true
}

// RenderHeaderLevel renders BlockHeader, SubHeader and SubSubHeader
func (r *HTMLRenderer) RenderHeaderLevel(block *notionapi.Block, level int, entering bool) bool {
	el := fmt.Sprintf("h%d", level)
	cls := fmt.Sprintf("notion-header-%d", level)
	attrs := []string{"class", cls}
	r.WriteElement(block, el, attrs, "", entering)
	return true
}

// RenderHeader renders BlockHeader
func (r *HTMLRenderer) RenderHeader(block *notionapi.Block, entering bool) bool {
	return r.RenderHeaderLevel(block, 1, entering)
}

// RenderSubHeader renders BlockSubHeader
func (r *HTMLRenderer) RenderSubHeader(block *notionapi.Block, entering bool) bool {
	return r.RenderHeaderLevel(block, 2, entering)
}

// RenderSubSubHeader renders BlocSubSubkHeader
func (r *HTMLRenderer) RenderSubSubHeader(block *notionapi.Block, entering bool) bool {
	return r.RenderHeaderLevel(block, 3, entering)
}

// RenderTodo renders BlockTodo
func (r *HTMLRenderer) RenderTodo(block *notionapi.Block, entering bool) bool {
	cls := "notion-todo"
	if block.IsChecked {
		cls = "notion-todo-checked"
	}
	attrs := []string{"class", cls}
	r.WriteElement(block, "div", attrs, "", entering)
	return true
}

// RenderToggle renders BlockToggle
func (r *HTMLRenderer) RenderToggle(block *notionapi.Block, entering bool) bool {
	if entering {
		attrs := []string{"class", "notion-toggle"}
		r.WriteElement(block, "div", attrs, "", entering)

		s := `<div class="notion-toggle-wrapper">`
		r.WriteString(s)
		r.Newline()
	} else {
		s := `</div>`
		r.WriteString(s)
		r.Newline()
		attrs := []string{"class", "notion-toggle"}
		r.WriteElement(block, "div", attrs, "", entering)
	}

	return true
}

// RenderQuote renders BlockQuote
func (r *HTMLRenderer) RenderQuote(block *notionapi.Block, entering bool) bool {
	cls := "notion-quote"
	attrs := []string{"class", cls}
	r.WriteElement(block, "quote", attrs, "", entering)
	return true
}

// RenderDivider renders BlockDivider
func (r *HTMLRenderer) RenderDivider(block *notionapi.Block, entering bool) bool {
	if !entering {
		return true
	}
	r.WriteString(`<hr class="notion-divider">`)
	r.Newline()
	return true
}

// RenderBookmark renders BlockBookmark
func (r *HTMLRenderer) RenderBookmark(block *notionapi.Block, entering bool) bool {
	content := fmt.Sprintf(`<a href="%s">%s</a>`, block.Link, block.Link)
	cls := "notion-bookmark"
	// TODO: don't render inlines (which seems to be title of the bookmarked page)
	// TODO: support caption
	// TODO: maybe support comments
	attrs := []string{"class", cls}
	r.WriteElement(block, "div", attrs, content, entering)
	return true
}

// RenderVideo renders BlockTweet
func (r *HTMLRenderer) RenderVideo(block *notionapi.Block, entering bool) bool {
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

	r.WriteElement(block, "iframe", attrs, "", entering)
	return true
}

// RenderTweet renders BlockTweet
func (r *HTMLRenderer) RenderTweet(block *notionapi.Block, entering bool) bool {
	uri := block.Source
	content := fmt.Sprintf(`Embedded tweet <a href="%s">%s</a>`, uri, uri)
	cls := "notion-embed"
	// TODO: don't render inlines (which seems to be title of the bookmarked page)
	// TODO: support caption
	// TODO: maybe support comments
	attrs := []string{"class", cls}
	r.WriteElement(block, "div", attrs, content, entering)
	return true
}

// RenderGist renders BlockGist
func (r *HTMLRenderer) RenderGist(block *notionapi.Block, entering bool) bool {
	uri := block.Source + ".js"
	cls := "notion-embed-gist"
	attrs := []string{"src", uri, "class", cls}
	// TODO: support caption
	// TODO: maybe support comments
	r.WriteElement(block, "script", attrs, "", entering)
	return true
}

// RenderEmbed renders BlockEmbed
func (r *HTMLRenderer) RenderEmbed(block *notionapi.Block, entering bool) bool {
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
	r.WriteElement(block, "div", attrs, content, entering)
	return true
}

// RenderFile renders BlockFile
func (r *HTMLRenderer) RenderFile(block *notionapi.Block, entering bool) bool {
	// TODO: best effort at making the URL readable
	uri := block.Source
	title := block.Title
	if title == "" {
		title = path.Base(uri)
	}
	title = html.EscapeString(title)
	content := fmt.Sprintf(`Embedded file: <a href="%s">%s</a>`, uri, title)
	fmt.Printf("File: '%s'\n", content)
	cls := "notion-embed"
	attrs := []string{"class", cls}
	r.WriteElement(block, "div", attrs, content, entering)
	return true
}

// RenderPDF renders BlockPDF
func (r *HTMLRenderer) RenderPDF(block *notionapi.Block, entering bool) bool {
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
	r.WriteElement(block, "div", attrs, content, entering)
	return true
}

// RenderImage renders BlockImage
func (r *HTMLRenderer) RenderImage(block *notionapi.Block, entering bool) bool {
	link := block.ImageURL
	attrs := []string{"class", "notion-image", "src", link}
	r.WriteElement(block, "img", attrs, "", entering)
	return true
}

// RenderColumnList renders BlockColumnList
// it's children are BlockColumn
func (r *HTMLRenderer) RenderColumnList(block *notionapi.Block, entering bool) bool {
	nColumns := len(block.Content)
	if nColumns == 0 {
		r.maybePanic("has no columns")
		return true
	}
	attrs := []string{"class", "notion-column-list"}
	r.WriteElement(block, "div", attrs, "", entering)
	return true
}

// RenderColumn renders BlockColumn
// it's parent is BlockColumnList
func (r *HTMLRenderer) RenderColumn(block *notionapi.Block, entering bool) bool {
	// TODO: get column ration from col.FormatColumn.ColumnRation, which is float 0...1
	attrs := []string{"class", "notion-column"}
	r.WriteElement(block, "div", attrs, "", entering)
	return true
}

// v is expected to be
// [
// 	[
// 		"foo"
// 	]
// ]
// and we want to return "foo"
// If not present or unexpected shape, return ""
// is still visible
// TODO: this mabye belongs to notionapi
func propsValueToText(v interface{}) string {
	if v == nil {
		return ""
	}

	// [ [ "foo" ]]
	a, ok := v.([]interface{})
	if !ok {
		return fmt.Sprintf("type1: %T", v)
	}
	// [ "foo" ]
	if len(a) == 0 {
		return ""
	}
	v = a[0]
	a, ok = v.([]interface{})
	if !ok {
		return fmt.Sprintf("type2: %T", v)
	}
	// "foo"
	if len(a) == 0 {
		return ""
	}
	v = a[0]
	str, ok := v.(string)
	if !ok {
		return fmt.Sprintf("type3: %T", v)
	}
	return str
}

// RenderCollectionView renders BlockCollectionView
// TODO: it renders all views, should render just one
// TODO: maybe add alternating background color for rows
func (r *HTMLRenderer) RenderCollectionView(block *notionapi.Block, entering bool) bool {
	viewInfo := block.CollectionViews[0]
	view := viewInfo.CollectionView
	columns := view.Format.TableProperties
	s := `<table class="notion-collection-view">`

	// generate header row
	s += `<thead><tr>`
	for _, col := range columns {
		colName := col.Property
		colInfo := viewInfo.Collection.CollectionSchema[colName]
		name := colInfo.Name
		s += `<th>` + html.EscapeString(name) + `</th>`
	}
	s += `</tr></thead>`
	s += `<tbody>`

	for _, row := range viewInfo.CollectionRows {
		s += `<tr>`
		props := row.Properties
		for _, col := range columns {
			colName := col.Property
			v := props[colName]
			colVal := propsValueToText(v)
			if colVal == "" {
				// use &nbsp; so that empty row still shows up
				// could also set a min-height to 1em or sth. like that
				s += `<td>&nbsp;</td>`
			} else {
				//colInfo := viewInfo.Collection.CollectionSchema[colName]
				// TODO: format colVal according to colInfo
				s += `<td>` + html.EscapeString(colVal) + `</td>`
			}
		}
		s += `</tr>`
	}
	s += `</tbody>`
	s += `</table>`
	r.WriteString(s)
	return true
}

// DefaultRenderFunc returns a defult rendering function for a type of
// a given block
func (r *HTMLRenderer) DefaultRenderFunc(blockType string) BlockRenderFunc {
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
	default:
		r.maybePanic("DefaultRenderFunc: unsupported block type '%s' in %s\n", blockType, r.Page.NotionURL())
	}
	return nil
}

// RenderBlock renders a block to html
func (r *HTMLRenderer) RenderBlock(block *notionapi.Block) {
	if block == nil {
		// a missing block
		return
	}
	def := r.DefaultRenderFunc(block.Type)
	handled := false
	if r.RenderBlockOverride != nil {
		handled = r.RenderBlockOverride(block, true)
	}
	if !handled && def != nil {
		def(block, true)
	}

	r.Level++
	for i, child := range block.Content {
		child.Parent = block
		r.CurrBlocks = block.Content
		r.CurrBlockIdx = i
		r.RenderBlock(child)
	}
	r.Level--

	handled = false
	if r.RenderBlockOverride != nil {
		handled = r.RenderBlockOverride(block, false)
	}
	if !handled && def != nil {
		def(block, false)
	}
}

// ToHTML renders a page to html
func (r *HTMLRenderer) ToHTML() []byte {
	r.Level = 0
	r.PushNewBuffer()

	r.RenderBlock(r.Page.Root)
	buf := r.PopBuffer()
	return buf.Bytes()
}

// ToHTML converts a page to HTML
func ToHTML(page *notionapi.Page) []byte {
	r := NewHTMLRenderer(page)
	return r.ToHTML()
}
