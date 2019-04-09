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
func (h *HTMLRenderer) log(format string, args ...interface{}) {
	if h.Log != nil {
		h.Log(format, args...)
	}
}

func (h *HTMLRenderer) maybePanic(format string, args ...interface{}) {
	if h.Log != nil {
		h.Log(format, args...)
	}
	if h.PanicOnFailures {
		panic(fmt.Sprintf(format, args...))
	}
}

// PushNewBuffer creates a new buffer and sets Buf to it
func (h *HTMLRenderer) PushNewBuffer() {
	h.bufs = append(h.bufs, h.Buf)
	h.Buf = &bytes.Buffer{}
}

// PopBuffer pops a buffer
func (h *HTMLRenderer) PopBuffer() *bytes.Buffer {
	res := h.Buf
	n := len(h.bufs)
	h.Buf = h.bufs[n-1]
	h.bufs = h.bufs[:n-1]
	return res
}

// Newline writes a newline to the buffer. It'll suppress multiple newlines.
func (h *HTMLRenderer) Newline() {
	d := h.Buf.Bytes()
	n := len(d)
	if n > 0 && d[n-1] != '\n' {
		h.Buf.WriteByte('\n')
	}
}

// WriteIndent writes 2 * Level spaces
func (h *HTMLRenderer) WriteIndent() {
	for n := 0; n < h.Level; n++ {
		h.Buf.WriteString("  ")
	}
}

func (h *HTMLRenderer) maybeGetID(block *notionapi.Block) string {
	if h.AppendID {
		return ""
	}
	return block.ID
}

// WriteElementWithContent is a helper class that writes HTML
// with optional class, optional id and optional content
func (h *HTMLRenderer) WriteElementWithContent(block *notionapi.Block, el string, class string, content string, entering bool) {
	if !entering {
		h.Buf.WriteString("</" + el + ">")
		h.Newline()
		return
	}
	s := "<" + el
	if class != "" {
		s += ` class="` + class + `"`
	}
	id := h.maybeGetID(block)
	if id != "" {
		s += ` id="` + id + `"`
	}
	s += ">"
	h.Buf.WriteString(s)
	h.Newline()
	h.Buf.WriteString(content)
	h.Newline()
	h.RenderInlines(block.InlineContent)
	h.Newline()
}

// WriteElement is a helper class that writes HTML with
// optional class, optional id and optional content
func (h *HTMLRenderer) WriteElement(block *notionapi.Block, el string, class string, entering bool) {
	h.WriteElementWithContent(block, el, class, "", entering)
}

// RenderInline renders inline block
func (h *HTMLRenderer) RenderInline(b *notionapi.InlineBlock) {
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
	h.Buf.WriteString(start + close)
}

// RenderInlines renders inline blocks
func (h *HTMLRenderer) RenderInlines(blocks []*notionapi.InlineBlock) {
	h.Level++
	for _, block := range blocks {
		h.RenderInline(block)
	}
	h.Level--
}

func (h *HTMLRenderer) renderCode(block *notionapi.Block, entering bool) bool {
	if !entering {
		h.Buf.WriteString("</code></pre>")
		h.Newline()
		return true
	}
	cls := "notion-code"
	lang := strings.ToLower(strings.TrimSpace(block.CodeLanguage))
	if lang != "" {
		cls += " notion-lang-" + lang
	}
	code := template.HTMLEscapeString(block.Code)
	s := fmt.Sprintf(`<pre class="%s"><code>%s`, cls, code)
	h.Buf.WriteString(s)
	return true
}

func (h *HTMLRenderer) renderPage(block *notionapi.Block, entering bool) bool {
	h.WriteElement(block, "div", "notion-page", entering)
	return true
}

func (h *HTMLRenderer) renderText(block *notionapi.Block, entering bool) bool {
	h.WriteElement(block, "p", "notion-text", entering)
	return true
}

func (h *HTMLRenderer) renderNumberedList(block *notionapi.Block, entering bool) bool {
	h.WriteElement(block, "ol", "notion-numbered-list", entering)
	return true
}

func (h *HTMLRenderer) renderBulletedList(block *notionapi.Block, entering bool) bool {
	h.WriteElement(block, "li", "notion-bulleted-list", entering)
	return true
}

func (h *HTMLRenderer) renderHeaderLevel(block *notionapi.Block, level int, entering bool) bool {
	el := fmt.Sprintf("h%d", level)
	cls := fmt.Sprintf("notion-header-%d", level)
	h.WriteElement(block, el, cls, entering)
	return true
}

func (h *HTMLRenderer) renderHeader(block *notionapi.Block, entering bool) bool {
	return h.renderHeaderLevel(block, 1, entering)
}

func (h *HTMLRenderer) renderSubHeader(block *notionapi.Block, entering bool) bool {
	return h.renderHeaderLevel(block, 2, entering)
}

func (h *HTMLRenderer) renderSubSubHeader(block *notionapi.Block, entering bool) bool {
	return h.renderHeaderLevel(block, 3, entering)
}

func (h *HTMLRenderer) renderTodo(block *notionapi.Block, entering bool) bool {
	cls := "notion-todo"
	if block.IsChecked {
		cls = "notion-todo-checked"
	}

	h.WriteElement(block, "div", cls, entering)
	return true
}

func (h *HTMLRenderer) renderToggle(block *notionapi.Block, entering bool) bool {
	h.maybePanic("NYI")
	return true
}

func (h *HTMLRenderer) renderQuote(block *notionapi.Block, entering bool) bool {
	h.maybePanic("NYI")
	return true
}

func (h *HTMLRenderer) renderDivider(block *notionapi.Block, entering bool) bool {
	if !entering {
		return true
	}
	h.Buf.WriteString(`<hr class="notion-divider">`)
	h.Newline()
	return true
}
func (h *HTMLRenderer) renderBookmark(block *notionapi.Block, entering bool) bool {
	content := fmt.Sprintf(`<a href="%s">%s</a>`, block.Link, block.Link)
	cls := "notion-bookmark"
	// TODO: don't render inlines (which seems to be title of the bookmarked page)
	// TODO: support caption
	// TODO: maybe support comments
	h.WriteElementWithContent(block, "div", cls, content, entering)
	return true
}

func (h *HTMLRenderer) renderGist(block *notionapi.Block, entering bool) bool {
	h.maybePanic("NYI")
	return true
}

func (h *HTMLRenderer) renderImage(block *notionapi.Block, entering bool) bool {
	h.maybePanic("NYI")
	return true
}

func (h *HTMLRenderer) renderColumnList(block *notionapi.Block, entering bool) bool {
	h.maybePanic("NYI")
	return true
}

func (h *HTMLRenderer) renderCollectionView(block *notionapi.Block, entering bool) bool {
	h.maybePanic("NYI")
	return true
}

func (h *HTMLRenderer) renderEmbed(block *notionapi.Block, entering bool) bool {
	h.maybePanic("NYI")
	return true
}

// DefaultRenderFunc returns a defult rendering function for a type of
// a given block
func (h *HTMLRenderer) DefaultRenderFunc(blockType string) BlockRenderFunc {
	switch blockType {
	case notionapi.BlockPage:
		return h.renderPage
	case notionapi.BlockText:
		return h.renderText
	case notionapi.BlockNumberedList:
		return h.renderBulletedList
	case notionapi.BlockBulletedList:
		return h.renderBulletedList
	case notionapi.BlockHeader:
		return h.renderHeader
	case notionapi.BlockSubHeader:
		return h.renderSubHeader
	case notionapi.BlockSubSubHeader:
		return h.renderSubSubHeader
	case notionapi.BlockTodo:
		return h.renderTodo
	case notionapi.BlockToggle:
		return h.renderToggle
	case notionapi.BlockQuote:
		return h.renderQuote
	case notionapi.BlockDivider:
		return h.renderDivider
	case notionapi.BlockCode:
		return h.renderCode
	case notionapi.BlockBookmark:
		return h.renderBookmark
	case notionapi.BlockGist:
		return h.renderGist
	case notionapi.BlockImage:
		return h.renderImage
	case notionapi.BlockColumnList:
		return h.renderColumnList
	case notionapi.BlockCollectionView:
		return h.renderCollectionView
	case notionapi.BlockEmbed:
		return h.renderEmbed
	default:
		h.maybePanic("DefaultRenderFunc: unsupported block type '%s'\n", blockType)
	}
	return nil
}

func (h *HTMLRenderer) blockHasChildren(blockType string) bool {
	switch blockType {
	case notionapi.BlockPage, notionapi.BlockNumberedList,
		notionapi.BlockBulletedList:
		return true
	case notionapi.BlockText:
		return false
	default:
		h.maybePanic("unrecognized block type '%s'", blockType)
	}
	return false
}

// RenderBlock renders a block to html
func (h *HTMLRenderer) RenderBlock(block *notionapi.Block) {
	if block == nil {
		// a missing block
		return
	}
	def := h.DefaultRenderFunc(block.Type)
	handled := false
	if h.RenderBlockOverride != nil {
		handled = h.RenderBlockOverride(block, true)
	}
	if !handled && def != nil {
		def(block, true)
	}

	// TODO: probably need to handle notionapi.BlockNumberedList
	// and notionapi.BlockBulletedList in a special way
	h.Level++
	for _, child := range block.Content {
		h.RenderBlock(child)
	}
	h.Level--

	/// TODO: not sure if this is needed
	/*
		if !h.blockHasChildren(block.Type) {
			if len(block.Content) != 0 {
				h.maybePanic("block has children but blockHasChildren() says it doesn't have children")
			}
		}
	*/

	handled = false
	if h.RenderBlockOverride != nil {
		handled = h.RenderBlockOverride(block, false)
	}
	if !handled && def != nil {
		def(block, false)
	}
}

// ToHTML renders a page to html
func (h *HTMLRenderer) ToHTML() []byte {
	h.Level = 0
	h.PushNewBuffer()

	h.RenderBlock(h.Page.Root)
	buf := h.PopBuffer()
	return buf.Bytes()
}

// ToHTML converts a page to HTML
func ToHTML(page *notionapi.Page) []byte {
	r := NewHTMLRenderer(page)
	return r.ToHTML()
}
