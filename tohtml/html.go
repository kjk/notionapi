package tohtml

import (
	"bytes"
	"fmt"
	"html/template"
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
	// Level is current depth of the tree. Useuful for pretty-printing indentation
	Level int
	// if true, adds id=${NotionID} attribute to HTML nodes
	AppendID bool
	// allows over-riding rendering of specific blocks
	// return false for default rendering
	RenderBlockOverride BlockRenderFunc
	// data provided by they caller, useful when providing
	// RenderBlockOverride
	Data interface{}

	// mostly for debugging, if set we'll log to it when encountering
	// structure we can't handle
	Log func(format string, args ...interface{})
	// mostly for debugging. If true will panic when encounters
	// structure it cannot handle (e.g. when Notion adds another
	// type of block)
	PanicOnFailures bool

	bufs []*bytes.Buffer
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

// WriteIndent writes 2 * Level spaces
func (r *HTMLRenderer) WriteIndent() {
	for n := 0; n < r.Level; n++ {
		r.Buf.WriteString("  ")
	}
}

func (r *HTMLRenderer) maybeGetID(block *notionapi.Block) string {
	if r.AppendID {
		return ""
	}
	return block.ID
}

// WriteElementWithContent is a helper class that writes HTML
// with optional class, optional id and optional content
func (r *HTMLRenderer) WriteElementWithContent(block *notionapi.Block, el string, class string, content string, entering bool) {
	if !entering {
		r.Buf.WriteString("</" + el + ">")
		r.Newline()
		return
	}
	s := "<" + el
	if class != "" {
		s += ` class="` + class + `"`
	}
	id := r.maybeGetID(block)
	if id != "" {
		s += ` id="` + id + `"`
	}
	s += ">"
	r.Buf.WriteString(s)
	r.Newline()
	r.Buf.WriteString(content)
	r.Newline()
	r.RenderInlines(block.InlineContent)
	r.Newline()
}

// WriteElement is a helper class that writes HTML with
// optional class, optional id and optional content
func (r *HTMLRenderer) WriteElement(block *notionapi.Block, el string, class string, entering bool) {
	r.WriteElementWithContent(block, el, class, "", entering)
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
	r.Buf.WriteString(start + close)
}

// RenderInlines renders inline blocks
func (r *HTMLRenderer) RenderInlines(blocks []*notionapi.InlineBlock) {
	r.Level++
	for _, block := range blocks {
		r.RenderInline(block)
	}
	r.Level--
}

func (r *HTMLRenderer) RenderCode(block *notionapi.Block, entering bool) bool {
	if !entering {
		r.Buf.WriteString("</code></pre>")
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
	r.Buf.WriteString(s)
	return true
}

func (r *HTMLRenderer) RenderPage(block *notionapi.Block, entering bool) bool {
	r.WriteElement(block, "div", "notion-page", entering)
	return true
}

func (r *HTMLRenderer) RenderText(block *notionapi.Block, entering bool) bool {
	r.WriteElement(block, "p", "notion-text", entering)
	return true
}

func (r *HTMLRenderer) RenderNumberedList(block *notionapi.Block, entering bool) bool {
	r.WriteElement(block, "ol", "notion-numbered-list", entering)
	return true
}

func (r *HTMLRenderer) RenderBulletedList(block *notionapi.Block, entering bool) bool {
	r.WriteElement(block, "li", "notion-bulleted-list", entering)
	return true
}

func (r *HTMLRenderer) RenderHeaderLevel(block *notionapi.Block, level int, entering bool) bool {
	el := fmt.Sprintf("h%d", level)
	cls := fmt.Sprintf("notion-header-%d", level)
	r.WriteElement(block, el, cls, entering)
	return true
}

func (r *HTMLRenderer) RenderHeader(block *notionapi.Block, entering bool) bool {
	return r.RenderHeaderLevel(block, 1, entering)
}

func (r *HTMLRenderer) RenderSubHeader(block *notionapi.Block, entering bool) bool {
	return r.RenderHeaderLevel(block, 2, entering)
}

func (r *HTMLRenderer) RenderSubSubHeader(block *notionapi.Block, entering bool) bool {
	return r.RenderHeaderLevel(block, 3, entering)
}

func (r *HTMLRenderer) RenderTodo(block *notionapi.Block, entering bool) bool {
	cls := "notion-todo"
	if block.IsChecked {
		cls = "notion-todo-checked"
	}

	r.WriteElement(block, "div", cls, entering)
	return true
}

func (r *HTMLRenderer) RenderToggle(block *notionapi.Block, entering bool) bool {
	r.maybePanic("NYI")
	return true
}

func (r *HTMLRenderer) RenderQuote(block *notionapi.Block, entering bool) bool {
	cls := "notion-quote"
	r.WriteElement(block, "quote", cls, entering)
	return true
}

func (r *HTMLRenderer) RenderDivider(block *notionapi.Block, entering bool) bool {
	if !entering {
		return true
	}
	r.Buf.WriteString(`<hr class="notion-divider">`)
	r.Newline()
	return true
}
func (r *HTMLRenderer) RenderBookmark(block *notionapi.Block, entering bool) bool {
	content := fmt.Sprintf(`<a href="%s">%s</a>`, block.Link, block.Link)
	cls := "notion-bookmark"
	// TODO: don't render inlines (which seems to be title of the bookmarked page)
	// TODO: support caption
	// TODO: maybe support comments
	r.WriteElementWithContent(block, "div", cls, content, entering)
	return true
}

func (r *HTMLRenderer) RenderGist(block *notionapi.Block, entering bool) bool {
	content := fmt.Sprintf(`Embedded gist <a href="%s">%s</a>`, block.Source, block.Source)
	cls := "notion-gist"
	// TODO: don't render inlines (which seems to be title of the bookmarked page)
	// TODO: support caption
	// TODO: maybe support comments
	r.WriteElementWithContent(block, "div", cls, content, entering)
	return true
}

func (r *HTMLRenderer) RenderImage(block *notionapi.Block, entering bool) bool {
	r.maybePanic("NYI")
	return true
}

func (r *HTMLRenderer) RenderColumnList(block *notionapi.Block, entering bool) bool {
	r.maybePanic("NYI")
	return true
}

func (r *HTMLRenderer) RenderCollectionView(block *notionapi.Block, entering bool) bool {
	r.maybePanic("NYI")
	return true
}

func (r *HTMLRenderer) RenderEmbed(block *notionapi.Block, entering bool) bool {
	r.maybePanic("NYI")
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
		return r.RenderBulletedList
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
	case notionapi.BlockGist:
		return r.RenderGist
	case notionapi.BlockImage:
		return r.RenderImage
	case notionapi.BlockColumnList:
		return r.RenderColumnList
	case notionapi.BlockCollectionView:
		return r.RenderCollectionView
	case notionapi.BlockEmbed:
		return r.RenderEmbed
	default:
		r.maybePanic("DefaultRenderFunc: unsupported block type '%s'\n", blockType)
	}
	return nil
}

func (r *HTMLRenderer) blockHasChildren(blockType string) bool {
	switch blockType {
	case notionapi.BlockPage, notionapi.BlockNumberedList,
		notionapi.BlockBulletedList:
		return true
	case notionapi.BlockText:
		return false
	default:
		r.maybePanic("unrecognized block type '%s'", blockType)
	}
	return false
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

	// TODO: probably need to handle notionapi.BlockNumberedList
	// and notionapi.BlockBulletedList in a special way
	r.Level++
	for _, child := range block.Content {
		r.RenderBlock(child)
	}
	r.Level--

	/// TODO: not sure if this is needed
	/*
		if !r.blockHasChildren(block.Type) {
			if len(block.Content) != 0 {
				r.maybePanic("block has children but blockHasChildren() says it doesn't have children")
			}
		}
	*/

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
