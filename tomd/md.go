package tomd

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kjk/notionapi"
)

func log(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func maybePanic(format string, args ...interface{}) {
	notionapi.MaybePanic(format, args...)
}

func markdownFileName(title, pageID string) string {
	s := notionapi.SafeName(title)
	return s + "-" + notionapi.ToDashID(pageID) + ".md"
}

// MarkdownFileNameForPage returns file name for markdown file
func MarkdownFileNameForPage(page *notionapi.Page) string {
	return markdownFileName(page.Root.Title, page.ID)
}

// BlockRenderFunc is a function for rendering a particular block
type BlockRenderFunc func(block *notionapi.Block) bool

// MarkdownRenderer converts a Page to HTML
type MarkdownRenderer struct {
	Page *notionapi.Page

	// Buf is where HTML is being written to
	Buf *bytes.Buffer

	// allows over-riding rendering of specific blocks
	// return false for default rendering
	RenderBlockOverride BlockRenderFunc

	// RewriteURL allows re-writing URLs e.g. to convert inter-notion URLs
	// to destination URLs
	RewriteURL func(url string) string

	// data provided by they caller, useful when providing
	// RenderBlockOverride
	Data interface{}

	// we need this to properly render ordered and numbered lists
	CurrBlocks   []*notionapi.Block
	CurrBlockIdx int

	Indent string
	ListNo int

	bufs []*bytes.Buffer
}

// NewMarkdownRenderer returns customizable Markdown renderer
func NewMarkdownRenderer(page *notionapi.Page) *MarkdownRenderer {
	return &MarkdownRenderer{
		Page: page,
	}
}

// PushNewBuffer creates a new buffer and sets Buf to it
func (r *MarkdownRenderer) PushNewBuffer() {
	r.bufs = append(r.bufs, r.Buf)
	r.Buf = &bytes.Buffer{}
}

// PopBuffer pops a buffer
func (r *MarkdownRenderer) PopBuffer() *bytes.Buffer {
	res := r.Buf
	n := len(r.bufs)
	r.Buf = r.bufs[n-1]
	r.bufs = r.bufs[:n-1]
	return res
}

// Eol writes end-of-line to the buffer. Doesn't write multiple.
func (r *MarkdownRenderer) Eol() {
	d := r.Buf.Bytes()
	n := len(d)
	if n > 0 && d[n-1] != '\n' {
		r.Buf.WriteByte('\n')
	}
}

// Newline writes a newline to the buffer. It'll suppress multiple newlines.
func (r *MarkdownRenderer) Newline() {
	d := r.Buf.Bytes()
	n := 0
	idx := len(d) - 1
	for idx >= 0 && d[idx] == '\n' {
		n++
		idx--
	}
	switch n {
	case 0:
		r.Buf.WriteString("\n\n")
	case 1:
		r.Buf.WriteByte('\n')
	}

}

func (r *MarkdownRenderer) incIndent() {
	r.Indent += "    "
}

func (r *MarkdownRenderer) decIndent() {
	r.Indent = r.Indent[:len(r.Indent)-4]
}

// WriteString writes a string to the buffer
func (r *MarkdownRenderer) WriteString(s string) {
	r.Buf.WriteString(s)
}

// PrevBlock is a block preceding current block
func (r *MarkdownRenderer) PrevBlock() *notionapi.Block {
	if r.CurrBlockIdx == 0 {
		return nil
	}
	return r.CurrBlocks[r.CurrBlockIdx-1]
}

// NextBlock is a block preceding current block
func (r *MarkdownRenderer) NextBlock() *notionapi.Block {
	nextIdx := r.CurrBlockIdx + 1
	lastIdx := len(r.CurrBlocks) - 1
	if nextIdx > lastIdx {
		return nil
	}
	return r.CurrBlocks[nextIdx]
}

// IsPrevBlockOfType returns true if previous block is of a given type
func (r *MarkdownRenderer) IsPrevBlockOfType(t string) bool {
	b := r.PrevBlock()
	if b == nil {
		return false
	}
	return b.Type == t
}

// IsNextBlockOfType returns true if next block is of a given type
func (r *MarkdownRenderer) IsNextBlockOfType(t string) bool {
	b := r.NextBlock()
	if b == nil {
		return false
	}
	return b.Type == t
}

// FormatDate formats the data
func (r *MarkdownRenderer) FormatDate(d *notionapi.Date) string {
	// TODO: allow over-riding date formatting
	s := notionapi.FormatDate(d)
	return fmt.Sprintf(`<span class="notion-date">@%s</span>`, s)
}

func isWs(c byte) bool {
	switch c {
	case ' ', '\n':
		return true
	}
	return false
}

func getBeforeWhitespace(text string) (string, string) {
	var res []byte
	for len(text) > 0 {
		c := text[0]
		if !isWs(c) {
			break
		}
		res = append(res, c)
		text = text[1:]
	}
	return string(res), text
}

func getAfterWhitespace(text string) (string, string) {
	var res []byte
	for len(text) > 0 {
		lastIdx := len(text) - 1
		c := text[lastIdx]
		if !isWs(c) {
			break
		}
		res = append(res, c)
		text = text[:lastIdx]
	}
	return string(res), text
}

// TODO: should respect Unicode whitespace
// test page: https://www.notion.so/5fea966407204d9080a5b989360b205f
func shuffleWhitespace(text string) (string, string, string) {
	before, text := getBeforeWhitespace(text)
	after, text := getAfterWhitespace(text)
	return before, text, after
}

func makeUserName(user *notionapi.User) string {
	s := user.GivenName
	if len(s) > 0 {
		s += " "
	}
	s += user.FamilyName
	if len(s) > 0 {
		return s
	}
	return user.ID
}

func resolveUser(page *notionapi.Page, userID string) string {
	for _, u := range page.Users {
		//log("u: %s, %s, %s\n", u.ID, u.GivenName, u.FamilyName)
		if u.ID == userID {
			return makeUserName(u)
		}
	}
	return userID
}

// RenderInline renders inline block
func (r *MarkdownRenderer) RenderInline(b *notionapi.InlineBlock, isLast bool) {
	var style string
	if b.AttrFlags&notionapi.AttrBold != 0 {
		style = `**`
	}
	if b.AttrFlags&notionapi.AttrItalic != 0 {
		style = `*`
	}
	if b.AttrFlags&notionapi.AttrStrikeThrought != 0 {
		style = "~~"
	}
	if b.AttrFlags&notionapi.AttrCode != 0 {
		style = "`"
	}
	text := b.Text
	// TODO: colors
	var before, after string
	if b.Link != "" {
		uri := b.Link
		if r.RewriteURL != nil {
			uri = r.RewriteURL(uri)
		}
		before, text, after = shuffleWhitespace(text)
		// TOOD: if text has "[" or "]" in it, has to escape
		text = fmt.Sprintf(`[%s](%s)`, text, uri)
	}
	if b.UserID != "" {
		text = fmt.Sprintf(`@%s`, resolveUser(r.Page, b.UserID))
	}
	if b.Date != nil {
		text = r.FormatDate(b.Date)
	}
	s := style
	s += text
	if style != "" {
		// move whitespace from inside style to outside, to match Notion export
		before, text, after = shuffleWhitespace(text)
	}
	if isLast {
		text = strings.TrimRight(text, " ")
		after = ""
	}
	r.WriteString(before + style + text + style + after)
}

// RenderInlines renders inline blocks
func (r *MarkdownRenderer) RenderInlines(blocks []*notionapi.InlineBlock, trimEndSpace bool) {
	lastIdx := len(blocks) - 1
	for idx, block := range blocks {
		r.RenderInline(block, trimEndSpace && idx == lastIdx)
	}
}

// GetInlineContent is like RenderInlines but instead of writing to
// output buffer, we return it as string
func (r *MarkdownRenderer) GetInlineContent(blocks []*notionapi.InlineBlock, trimeEndSpace bool) string {
	r.PushNewBuffer()
	r.RenderInlines(blocks, trimeEndSpace)
	return r.PopBuffer().String()
}

// RenderCode renders BlockCode
func (r *MarkdownRenderer) RenderCode(block *notionapi.Block) {
	// TODO: implement me
	r.WriteString("RenderCode NYI\n")
}

// RenderPage renders BlockPage
func (r *MarkdownRenderer) RenderPage(block *notionapi.Block) {
	tp := block.GetPageType()
	if tp == notionapi.BlockPageTopLevel {
		title := r.GetInlineContent(block.TitleFull, false)
		r.WriteString("# " + title + "\n")
		r.RenderChildren(block)
		return
	}
	if tp == notionapi.BlockPageLink {
		return
	}
	// TODO: if block.Title has "[" or "]" in it, needs to escape
	fileName := markdownFileName(block.Title, block.ID)
	title := r.GetInlineContent(block.TitleFull, false)
	s := fmt.Sprintf("[%s](./%s)", title, fileName)
	r.WriteString(s)
	r.Eol()
}

// RenderText renders BlockText
func (r *MarkdownRenderer) RenderText(block *notionapi.Block) {
	r.RenderInlines(block.InlineContent, false)
	r.Newline()
}

// RenderToggle renders BlockToggle
func (r *MarkdownRenderer) RenderToggle(block *notionapi.Block) {
	r.incIndent()
	defer r.decIndent()

	r.WriteString("- ")
	r.RenderInlines(block.InlineContent, true)
	r.Eol()

	r.RenderChildren(block)
}

// RenderNumberedList renders BlockNumberedList
func (r *MarkdownRenderer) RenderNumberedList(block *notionapi.Block) {
	r.incIndent()
	defer r.decIndent()

	isPrevSame := r.IsPrevBlockOfType(notionapi.BlockNumberedList)
	if isPrevSame {
		r.ListNo++
	} else {
		r.ListNo = 1
	}
	r.WriteString(fmt.Sprintf("%d. ", r.ListNo))
	r.RenderInlines(block.InlineContent, false)
	r.Eol()

	r.RenderChildren(block)
}

// RenderBulletedList renders BlockBulletedList
func (r *MarkdownRenderer) RenderBulletedList(block *notionapi.Block) {
	r.incIndent()
	defer r.decIndent()

	r.WriteString("- ")
	r.RenderInlines(block.InlineContent, true)
	r.Eol()

	r.RenderChildren(block)
}

// RenderHeaderLevel renders BlockHeader, SubHeader and SubSubHeader
func (r *MarkdownRenderer) RenderHeaderLevel(block *notionapi.Block, level int) {
	s := ""
	switch level {
	case 1:
		s = "# "
	case 2:
		s = "## "
	case 3:
		s = "### "
	default:
		// TODO: shouldn't happen?
		for i := 1; i < level; i++ {
			s += "#"
		}
		s += " "
	}
	content := r.GetInlineContent(block.InlineContent, false)
	content = strings.TrimRight(content, " ")
	r.WriteString(s + content)
	r.Newline()
}

// RenderHeader renders BlockHeader
func (r *MarkdownRenderer) RenderHeader(block *notionapi.Block) {
	r.RenderHeaderLevel(block, 1)
}

// RenderSubHeader renders BlockSubHeader
func (r *MarkdownRenderer) RenderSubHeader(block *notionapi.Block) {
	r.RenderHeaderLevel(block, 2)
}

// RenderSubSubHeader renders BlocSubSubkHeader
func (r *MarkdownRenderer) RenderSubSubHeader(block *notionapi.Block) {
	r.RenderHeaderLevel(block, 3)
}

// RenderTodo renders BlockTodo
func (r *MarkdownRenderer) RenderTodo(block *notionapi.Block) {
	r.incIndent()
	defer r.decIndent()

	text := r.GetInlineContent(block.InlineContent, true)
	s := fmt.Sprintf("- [ ]  %s\n", text)
	r.WriteString(s)

	r.RenderChildren(block)
}

// RenderQuote renders BlockQuote
func (r *MarkdownRenderer) RenderQuote(block *notionapi.Block) {
	text := r.GetInlineContent(block.InlineContent, true)
	s := fmt.Sprintf("> %s\n", text)
	r.WriteString(s)
}

// RenderCallout renders BlockCallout
func (r *MarkdownRenderer) RenderCallout(block *notionapi.Block) {
	// TODO: implement me
	r.WriteString("RenderCallout NYI\n")
}

// RenderDivider renders BlockDivider
func (r *MarkdownRenderer) RenderDivider(block *notionapi.Block) {
	r.WriteString("---\n\n")
}

// RenderBookmark renders BlockBookmark
func (r *MarkdownRenderer) RenderBookmark(block *notionapi.Block) {
	title := notionapi.GetInlineText(block.InlineContent)
	uri := block.Link
	if title == "" && uri == "" {
		return
	}
	r.WriteString(fmt.Sprintf("[%s](%s)\n", title, uri))
}

// RenderTweet renders BlockTweet
func (r *MarkdownRenderer) RenderTweet(block *notionapi.Block) {
	r.WriteString("RenderTweet NYI\n")
}

// RenderGist renders BlockGist
func (r *MarkdownRenderer) RenderGist(block *notionapi.Block) {
	r.WriteString("RenderGist NYI\n")
}

func localFileNameFromURL(fileID, uri string) string {
	parts := strings.Split(uri, "/")
	fileName := parts[len(parts)-1]
	parts = strings.SplitN(fileName, ".", 2)
	ext := ""
	if len(parts) == 2 {
		fileName = parts[0]
		ext = "." + parts[1]
	}
	return fileName + "-" + fileID + ext
}

// for video, image, pdf
func getEmbeddedFileNameAndURL(block *notionapi.Block) (string, string) {
	source := block.Source
	if source == "" {
		return "", ""
	}
	if len(block.FileIDs) == 0 {
		return source, source
	}

	// assumes locally downloaded file
	fileID := block.FileIDs[0]
	parts := strings.Split(source, "/")
	fileName := parts[len(parts)-1]
	parts = strings.SplitN(fileName, ".", 2)
	ext := ""
	if len(parts) == 2 {
		fileName = parts[0]
		ext = "." + parts[1]
	}
	uri := fileName + "-" + fileID + ext
	// TODO: did I mean to use uri, uri and not fileName, uri?
	// TODO: use localFileNameFromURL
	return uri, uri
}

// RenderVideo renders BlockTweet
func (r *MarkdownRenderer) RenderVideo(block *notionapi.Block) {
	name, uri := getEmbeddedFileNameAndURL(block)
	s := fmt.Sprintf("[%s](%s)\n", name, uri)
	r.WriteString(s)
}

// RenderFile renders BlockFile
func (r *MarkdownRenderer) RenderFile(block *notionapi.Block) {
	fileID := block.FileIDs[0]
	localFileName := localFileNameFromURL(fileID, block.Source)
	name := block.Title
	s := fmt.Sprintf("[%s](%s)\n", name, localFileName)
	r.WriteString(s)
}

// RenderPDF renders BlockPDF
func (r *MarkdownRenderer) RenderPDF(block *notionapi.Block) {
	name, uri := getEmbeddedFileNameAndURL(block)
	s := fmt.Sprintf("[%s](%s)\n", name, uri)
	r.WriteString(s)
}

// RenderEmbed renders BlockEmbed
func (r *MarkdownRenderer) RenderEmbed(block *notionapi.Block) {
	uri := block.Source
	s := fmt.Sprintf("[%s](%s)\n", uri, uri)
	r.WriteString(s)
}

// RenderImage renders BlockImage
func (r *MarkdownRenderer) RenderImage(block *notionapi.Block) {
	// TODO: not sure if always has FileIDs
	if len(block.FileIDs) == 0 {
		r.WriteString("RenderImage when len(FileIDs) == 0 NYI\n")
	}
	source := block.Source // also present in block.Format.DisplaySource
	// source looks like: "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/e5470cfd-08f0-4fb8-8ec2-452ca1a3f05e/Schermafbeelding2018-06-19om09.52.45.png"
	fileID := block.FileIDs[0]
	parts := strings.Split(source, "/")
	fileName := parts[len(parts)-1]
	parts = strings.SplitN(fileName, ".", 2)
	ext := ""
	if len(parts) == 2 {
		fileName = parts[0]
		ext = "." + parts[1]
	}
	s := fmt.Sprintf("![](%s-%s%s)\n", fileName, fileID, ext)
	r.WriteString(s)
}

// RenderColumnList renders BlockColumnList
// it's children are BlockColumn
func (r *MarkdownRenderer) RenderColumnList(block *notionapi.Block) {
	r.RenderChildren(block)
}

// RenderColumn renders BlockColumn
// it's parent is BlockColumnList
func (r *MarkdownRenderer) RenderColumn(block *notionapi.Block) {
	r.RenderChildren(block)
}

// RenderCollectionView renders BlockCollectionView
func (r *MarkdownRenderer) RenderCollectionView(block *notionapi.Block) {
	if len(block.CollectionViews) == 0 {
		return
	}
	r.WriteString("RendeRenderCollectionViewrCode NYI\n")

	r.RenderChildren(block)
}

// DefaultRenderFunc returns a defult rendering function for a type of
// a given block
func (r *MarkdownRenderer) DefaultRenderFunc(blockType string) func(*notionapi.Block) {
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
	default:
		maybePanic("DefaultRenderFunc: unsupported block type '%s' in %s\n", blockType, r.Page.NotionURL())
	}
	return nil
}

func isEmptyBlock(block *notionapi.Block) bool {
	if block.Type == notionapi.BlockText {
		return len(block.InlineContent) == 0
	}
	if block.Type == notionapi.BlockPage {
		return block.GetPageType() == notionapi.BlockPageLink
	}
	return false
}

func skipChildren(block *notionapi.Block) bool {
	if len(block.Content) == 0 {
		// optimization for blocks that don't have children
		// even if they can have them
		return true
	}
	if block.Type == notionapi.BlockPage {
		tp := block.GetPageType()
		if tp == notionapi.BlockPageTopLevel {
			return false
		}
		// those are just links to pages but can have Content
		// list. We don't want to render those
		return true
	}
	return false
}

func (r *MarkdownRenderer) RenderChildren(block *notionapi.Block) {
	if skipChildren(block) {
		return
	}
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
}

func (r *MarkdownRenderer) AddNewlineBeforeBlock(block *notionapi.Block) {
	// TODO: hacky
	addNl := true
	// I think this should depend on previous block i.e.
	// don't add newline if previous block the same as this one
	switch block.Type {
	case notionapi.BlockNumberedList,
		notionapi.BlockBulletedList,
		notionapi.BlockToggle,
		notionapi.BlockTodo:
		addNl = false
	}
	if addNl {
		r.Newline()
	}
	if !isEmptyBlock(block) {
		if len(r.Indent) > 0 {
			r.WriteString(r.Indent)
		}
	}
}

// RenderBlock renders a block to html
func (r *MarkdownRenderer) RenderBlock(block *notionapi.Block) {
	if block == nil {
		// a missing block, can happen if we don't have access to a referenced block
		return
	}

	if r.RenderBlockOverride != nil && r.RenderBlockOverride(block) {
		return
	}

	def := r.DefaultRenderFunc(block.Type)
	if def != nil {
		r.AddNewlineBeforeBlock(block)
		def(block)
	}
}

func (r *MarkdownRenderer) ToMarkdown() []byte {
	r.PushNewBuffer()

	r.RenderBlock(r.Page.Root)
	buf := r.PopBuffer()
	// a bit of a hack to account for adding newlines before and after each block
	// which adds empty lines at top and bottom
	d := buf.Bytes()
	d = bytes.TrimSpace(d)
	return d
}

// ToMarkdown converts a page to Markdown
func ToMarkdown(page *notionapi.Page) []byte {
	r := NewMarkdownRenderer(page)
	return r.ToMarkdown()
}
