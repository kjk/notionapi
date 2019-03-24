package tohtml

import (
	"bytes"
	"fmt"

	"github.com/kjk/notionapi"
)

// BlockRenderFunc is a function for rendering a particular
type BlockRenderFunc func(block *notionapi.Block, entering bool) bool

// ToHTML converts a Page to HTML
type ToHTML struct {
	// Buf is where HTML is being written to
	Buf  *bytes.Buffer
	Page *notionapi.Page
	// Level is current depth of the tree. Useuful for pretty-printing indentation
	Level int
	// if true, adds id=${NotionID} attribute to HTML nodes
	AppendID bool
	// if set, will be called to log unexpected things
	Log func(format string, args ...interface{})
	// allows over-riding rendering of specific blocks
	// return false for default rendering
	RenderBlockOverride BlockRenderFunc
	// data provided by they caller, useful when providing
	// RenderBlockOverride
	Data interface{}

	bufs []*bytes.Buffer
}

func (h *ToHTML) lg(format string, args ...interface{}) {
	if h.Log != nil {
		h.Log(format, args...)
	}
}

// PushNewBuffer creates a new buffer and sets Buf to it
func (h *ToHTML) PushNewBuffer() {
	h.bufs = append(h.bufs, h.Buf)
	h.Buf = &bytes.Buffer{}
}

// PopBuffer pops a buffer
func (h *ToHTML) PopBuffer() *bytes.Buffer {
	res := h.Buf
	n := len(h.bufs)
	h.Buf = h.bufs[n-1]
	h.bufs = h.bufs[:n-1]
	return res
}

// Newline writes a newline to the buffer as long as it doesn't already
// ends with newline
func (h *ToHTML) Newline() {
	d := h.Buf.Bytes()
	n := len(d)
	if n > 0 && d[n-1] != '\n' {
		h.Buf.WriteByte('\n')
	}
}

// WriteIndent writes 2 * Level spaces
func (h *ToHTML) WriteIndent() {
	for n := 0; n < h.Level; n++ {
		h.Buf.WriteString("  ")
	}
}

// WriteElement is a helper class that writes HTML with a given
// class (optional) and id (also optional)
func (h *ToHTML) WriteElement(el string, class string, id string, entering bool) {
	if !entering {
		h.Buf.WriteString("</" + el + ">")
		return
	}
	s := "<" + el
	if class != "" {
		s += ` class="` + class + `"`
	}
	if id != "" {
		s += ` id="` + id + `"`
	}
	s += ">"
	h.Buf.WriteString(s)
}

func (h *ToHTML) maybeGetID(block *notionapi.Block) string {
	if h.AppendID {
		return ""
	}
	return block.ID
}

// RenderInline renders inline block
func (h *ToHTML) RenderInline(b *notionapi.InlineBlock) {
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
	// TODO: allow over-riding rendering of links/user ids/dates etc.
	if b.Link != "" {
		link := b.Link
		start += fmt.Sprintf(`<a class="notion-link" href="%s">%s</a>`, link, b.Text)
		skipText = true
	}
	if b.UserID != "" {
		start += fmt.Sprintf(`<span class="notion-user">@%s</span>`, b.UserID)
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
func (h *ToHTML) RenderInlines(blocks []*notionapi.InlineBlock) {
	h.Level++
	for _, block := range blocks {
		h.RenderInline(block)
	}
	h.Level--
}

func (h *ToHTML) renderPage(block *notionapi.Block, entering bool) bool {
	h.WriteElement("div", "notion-page", h.maybeGetID(block), entering)
	h.Newline()
	return true
}

func (h *ToHTML) renderText(block *notionapi.Block, entering bool) bool {
	h.WriteElement("p", "notion-text", h.maybeGetID(block), entering)
	h.Newline()
	if entering {
		h.RenderInlines(block.InlineContent)
	}
	return true
}

func (h *ToHTML) renderNumberedList(block *notionapi.Block, entering bool) bool {
	h.WriteElement("ol", "notion-numbered-list", h.maybeGetID(block), entering)
	h.Newline()
	return true
}

func (h *ToHTML) renderBulletedList(block *notionapi.Block, entering bool) bool {
	h.WriteElement("li", "notion-bulleted-list", h.maybeGetID(block), entering)
	h.Newline()
	return true
}

func (h *ToHTML) renderHeaderLevel(block *notionapi.Block, level int, entering bool) bool {
	el := fmt.Sprintf("h%d", level)
	cls := fmt.Sprintf("notion-header-%d", level)
	h.WriteElement(el, cls, h.maybeGetID(block), entering)
	h.Newline()
	if entering {
		h.RenderInlines(block.InlineContent)
	}
	return true
}

func (h *ToHTML) renderHeader(block *notionapi.Block, entering bool) bool {
	return h.renderHeaderLevel(block, 1, entering)
}

func (h *ToHTML) renderSubHeader(block *notionapi.Block, entering bool) bool {
	return h.renderHeaderLevel(block, 2, entering)
}

func (h *ToHTML) renderSubSubHeader(block *notionapi.Block, entering bool) bool {
	return h.renderHeaderLevel(block, 3, entering)
}

func (h *ToHTML) renderTodo(block *notionapi.Block, entering bool) bool {
	return true
}

func (h *ToHTML) renderToggle(block *notionapi.Block, entering bool) bool {
	return true
}

func (h *ToHTML) renderQuote(block *notionapi.Block, entering bool) bool {
	return true
}

func (h *ToHTML) renderDivider(block *notionapi.Block, entering bool) bool {
	return true
}

func (h *ToHTML) renderCode(block *notionapi.Block, entering bool) bool {
	return true
}

func (h *ToHTML) renderBookmark(block *notionapi.Block, entering bool) bool {
	return true
}

func (h *ToHTML) renderGist(block *notionapi.Block, entering bool) bool {
	return true
}

func (h *ToHTML) renderImage(block *notionapi.Block, entering bool) bool {
	return true
}

func (h *ToHTML) renderColumnList(block *notionapi.Block, entering bool) bool {
	return true
}

func (h *ToHTML) renderCollectionView(block *notionapi.Block, entering bool) bool {
	return true
}

func (h *ToHTML) renderEmbed(block *notionapi.Block, entering bool) bool {
	return true
}

// DefaultRenderFunc returns a defult rendering function for a type of
// a given block
func (h *ToHTML) DefaultRenderFunc(blockType string) BlockRenderFunc {
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
		h.lg("DefaultRenderFunc: unsupported block type '%s'\n", blockType)
	}
	return nil
}

func blockHasChildren(blockType string, lg func(format string, args ...interface{})) bool {
	switch blockType {
	case notionapi.BlockPage, notionapi.BlockNumberedList,
		notionapi.BlockBulletedList:
		return true
	case notionapi.BlockText:
		return false
	default:
		if lg != nil {
			lg("unrecognized block type '%s'", blockType)
		}
	}
	return false
}

// RenderBlock renders a block to html
func (h *ToHTML) RenderBlock(block *notionapi.Block) {
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

	if !blockHasChildren(block.Type, h.Log) {
		//panicIf(len(block.Content) != 0)
	}

	handled = false
	if h.RenderBlockOverride != nil {
		handled = h.RenderBlockOverride(block, false)
	}
	if !handled && def != nil {
		def(block, false)
	}
}

// RenderPage renders a page to html
func (h *ToHTML) RenderPage(page *notionapi.Page) string {
	h.Page = page
	h.Level = 0
	h.PushNewBuffer()

	h.RenderBlock(page.Root)
	buf := h.PopBuffer()
	return buf.String()
}
