package tomarkdown

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kjk/notionapi"
)

func maybePanic(format string, args ...interface{}) {
	notionapi.MaybePanic(format, args...)
}

func markdownFileName(title, pageID string) string {
	s := notionapi.SafeName(title)
	return s + "-" + notionapi.ToDashID(pageID) + ".md"
}

// MarkdownFileNameForPage returns file name for markdown file
func MarkdownFileNameForPage(page *notionapi.Page) string {
	rootPage := page.Root()
	return markdownFileName(rootPage.Title, page.ID)
}

// BlockRenderFunc is a function for rendering a particular block
type BlockRenderFunc func(block *notionapi.Block) bool

// MarkdownRenderer converts a Page to HTML
type Converter struct {
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

// NewConverter returns customizable Markdown renderer
func NewConverter(page *notionapi.Page) *Converter {
	return &Converter{
		Page: page,
	}
}

// PushNewBuffer creates a new buffer and sets Buf to it
func (c *Converter) PushNewBuffer() {
	c.bufs = append(c.bufs, c.Buf)
	c.Buf = &bytes.Buffer{}
}

// PopBuffer pops a buffer
func (c *Converter) PopBuffer() *bytes.Buffer {
	res := c.Buf
	n := len(c.bufs)
	c.Buf = c.bufs[n-1]
	c.bufs = c.bufs[:n-1]
	return res
}

// Eol writes end-of-line to the buffer. Doesn't write multiple.
func (c *Converter) Eol() {
	d := c.Buf.Bytes()
	n := len(d)
	if n > 0 && d[n-1] != '\n' {
		c.Buf.WriteByte('\n')
	}
}

func isWs(c byte) bool {
	switch c {
	case ' ':
		return true
	}
	return false
}

func bufTrimWhitespaceRight(buf *bytes.Buffer) {
	d := buf.Bytes()
	n := len(d) - 1
	nRemoved := 0
	for (n-nRemoved) >= 0 && isWs(d[n-nRemoved]) {
		nRemoved++
	}
	if nRemoved > 0 {
		buf.Truncate(len(d) - nRemoved)
	}
}

func (c *Converter) trimWhitespaceRight() {
	bufTrimWhitespaceRight(c.Buf)
}

// Newline writes a newline to the buffer. It'll suppress multiple newlines.
func (c *Converter) Newline() {
	d := c.Buf.Bytes()
	n := 0
	idx := len(d) - 1
	for idx >= 0 && d[idx] == '\n' {
		n++
		idx--
	}
	switch n {
	case 0:
		c.Buf.WriteString("\n\n")
	case 1:
		c.Buf.WriteByte('\n')
	}

}

func (c *Converter) incIndent() {
	c.Indent += "    "
}

func (c *Converter) decIndent() {
	c.Indent = c.Indent[:len(c.Indent)-4]
}

// WriteString writes a string to the buffer
func (c *Converter) WriteString(s string) {
	c.Buf.WriteString(s)
}

// PrevBlock is a block preceding current block
func (c *Converter) PrevBlock() *notionapi.Block {
	if c.CurrBlockIdx == 0 {
		return nil
	}
	return c.CurrBlocks[c.CurrBlockIdx-1]
}

// NextBlock is a block preceding current block
func (c *Converter) NextBlock() *notionapi.Block {
	nextIdx := c.CurrBlockIdx + 1
	lastIdx := len(c.CurrBlocks) - 1
	if nextIdx > lastIdx {
		return nil
	}
	return c.CurrBlocks[nextIdx]
}

// IsPrevBlockOfType returns true if previous block is of a given type
func (c *Converter) IsPrevBlockOfType(t string) bool {
	b := c.PrevBlock()
	if b == nil {
		return false
	}
	return b.Type == t
}

// IsNextBlockOfType returns true if next block is of a given type
func (c *Converter) IsNextBlockOfType(t string) bool {
	b := c.NextBlock()
	if b == nil {
		return false
	}
	return b.Type == t
}

// FormatDate formats the date
func (c *Converter) FormatDate(d *notionapi.Date) string {
	// TODO: allow over-riding date formatting
	s := notionapi.FormatDate(d)
	return s
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

/*
func (c *Converter) bufferEndsWith(s string) bool {
	d := c.Buf.Bytes()
	if len(d) < len(s) {
		return false
	}
	n := len(d)
	ns := len(s)
	for i := 0; i < ns; i++ {
		c := s[i]
		if d[n-ns-i] != c {
			return false
		}
	}
	return true
}
*/

// InlineToString renders inline block
func (c *Converter) InlineToString(b *notionapi.TextSpan) string {
	text := b.Text
	var start, end, before, after string
	for _, attr := range b.Attrs {
		switch notionapi.AttrGetType(attr) {
		case notionapi.AttrBold:
			start += "**"
			end = "**" + end
		case notionapi.AttrItalic:
			start += "*"
			end = "*" + end
		case notionapi.AttrStrikeThrought:
			start += "~~"
			end = "~~" + end
		case notionapi.AttrCode:
			start += "`"
			end = "`" + end
		case notionapi.AttrPage:
			pageID := notionapi.AttrGetPageID(attr)
			// TODO: find the page
			// TODO: needs to download info when recursively scanning
			// for pages
			pageTitle := ""
			uri := "https://www.notion.so/" + pageID
			if c.RewriteURL != nil {
				uri = c.RewriteURL(uri)
			}
			start += fmt.Sprintf(`[%s](%s)`, pageTitle, uri)
		case notionapi.AttrLink:
			uri := notionapi.AttrGetLink(attr)
			if c.RewriteURL != nil {
				uri = c.RewriteURL(uri)
			}
			before, text, after = shuffleWhitespace(text)
			// TOOD: if text has "[" or "]" in it, has to escape
			text = fmt.Sprintf(`%s[%s](%s)%s`, before, text, uri, after)
		case notionapi.AttrUser:
			userID := notionapi.AttrGetUserID(attr)
			text = fmt.Sprintf(`@%s`, notionapi.GetUserNameByID(c.Page, userID))
		case notionapi.AttrDate:
			date := notionapi.AttrGetDate(attr)
			text = c.FormatDate(date)
		}
	}
	// move whitespace from inside style to outside, to match Notion export
	before, text, after = shuffleWhitespace(text)
	start = before + start
	end = end + after
	return start + text + end
}

func (c *Converter) RenderInline(b *notionapi.TextSpan) {
	s := c.InlineToString(b)
	c.Printf(s)
}

// RenderInlines renders inline blocks
func (c *Converter) RenderInlines(blocks []*notionapi.TextSpan, trimEndSpace bool) {
	n := c.Buf.Len()
	for _, block := range blocks {
		c.RenderInline(block)
	}

	if trimEndSpace && c.Buf.Len() > n {
		c.trimWhitespaceRight()
	}
}

// GetInlineContent is like RenderInlines but instead of writing to
// output buffer, we return it as string
func (c *Converter) GetInlineContent(blocks []*notionapi.TextSpan, trimeEndSpace bool) string {
	c.PushNewBuffer()
	c.RenderInlines(blocks, trimeEndSpace)
	return c.PopBuffer().String()
}

// RenderCode renders BlockCode
func (c *Converter) RenderCode(block *notionapi.Block) {
	code := block.Code
	ind := "    "
	parts := strings.Split(code, "\n")
	for _, part := range parts {
		c.Printf(ind + part + "\n")
	}
}

func (c *Converter) renderRootPage(block *notionapi.Block) {
	title := c.GetInlineContent(block.InlineContent, false)
	c.Printf("# " + title)
	c.Newline()
	c.RenderChildren(block)
}

func escapeMarkdownLinkText(s string) string {
	// TODO: shouldn't escape those that are already escaped
	s = strings.Replace(s, `[`, `\[`, -1)
	s = strings.Replace(s, `]`, `\]`, -1)
	return s
}

// RenderPage renders BlockPage
func (c *Converter) RenderPage(block *notionapi.Block) {
	if c.Page.IsRoot(block) {
		c.renderRootPage(block)
		return
	}
	title := c.GetInlineContent(block.InlineContent, false)
	uri := ""
	if c.RewriteURL != nil {
		uri = c.RewriteURL("https://notion.so/" + block.ID)
	} else {
		uri = "./" + markdownFileName(block.Title, block.ID)
	}
	title = escapeMarkdownLinkText(title)
	c.Printf("[%s](%s)", title, uri)
	c.Eol()
}

// RenderText renders BlockText
func (c *Converter) RenderText(block *notionapi.Block) {
	c.RenderInlines(block.InlineContent, false)
	c.Newline()
	c.RenderChildren(block)
}

// RenderToggle renders BlockToggle
func (c *Converter) RenderToggle(block *notionapi.Block) {
	c.Printf("- ")
	c.RenderInlines(block.InlineContent, true)
	c.Eol()

	c.incIndent()
	defer c.decIndent()

	c.RenderChildren(block)
}

// RenderNumberedList renders BlockNumberedList
func (c *Converter) RenderNumberedList(block *notionapi.Block) {
	c.incIndent()
	defer c.decIndent()

	isPrevSame := c.IsPrevBlockOfType(notionapi.BlockNumberedList)
	if isPrevSame {
		c.ListNo++
	} else {
		c.ListNo = 1
	}
	c.WriteString(fmt.Sprintf("%d. ", c.ListNo))
	c.RenderInlines(block.InlineContent, false)
	c.Eol()

	c.RenderChildren(block)
}

// RenderBulletedList renders BlockBulletedList
func (c *Converter) RenderBulletedList(block *notionapi.Block) {
	c.incIndent()
	defer c.decIndent()

	c.WriteString("- ")
	c.RenderInlines(block.InlineContent, true)
	c.Eol()

	c.RenderChildren(block)
}

// RenderHeaderLevel renders BlockHeader, SubHeader and SubSubHeader
func (c *Converter) RenderHeaderLevel(block *notionapi.Block, level int) {
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
	content := c.GetInlineContent(block.InlineContent, false)
	content = strings.TrimRight(content, " ")
	c.WriteString(s + content)
	c.Newline()
}

// RenderHeader renders BlockHeader
func (c *Converter) RenderHeader(block *notionapi.Block) {
	c.RenderHeaderLevel(block, 1)
}

// RenderSubHeader renders BlockSubHeader
func (c *Converter) RenderSubHeader(block *notionapi.Block) {
	c.RenderHeaderLevel(block, 2)
}

// RenderSubSubHeader renders BlocSubSubkHeader
func (c *Converter) RenderSubSubHeader(block *notionapi.Block) {
	c.RenderHeaderLevel(block, 3)
}

func (c *Converter) Printf(format string, args ...interface{}) {
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	c.Buf.WriteString(s)
}

// RenderTodo renders BlockTodo
func (c *Converter) RenderTodo(block *notionapi.Block) {
	text := c.GetInlineContent(block.InlineContent, true)

	if block.IsChecked {
		c.Printf("- [x]  %s\n", text)
	} else {
		c.Printf("- [ ]  %s\n", text)
	}

	c.incIndent()
	defer c.decIndent()
	c.RenderChildren(block)
}

// RenderQuote renders BlockQuote
func (c *Converter) RenderQuote(block *notionapi.Block) {
	text := c.GetInlineContent(block.InlineContent, true)
	lines := strings.Split(strings.Replace(text, "\r\n", "\n", -1), "\n")
	for _, line := range lines {
		s := fmt.Sprintf("> %s\n", line)
		c.WriteString(s)
	}
}

// RenderCallout renders BlockCallout
func (c *Converter) RenderCallout(block *notionapi.Block) {
	// TODO: implement me
	c.WriteString("RenderCallout NYI\n")
}

// RenderDivider renders BlockDivider
func (c *Converter) RenderDivider(block *notionapi.Block) {
	c.Printf("---\n\n")
}

// RenderBookmark renders BlockBookmark
func (c *Converter) RenderBookmark(block *notionapi.Block) {
	title := notionapi.TextSpansToString(block.InlineContent)
	uri := block.Link
	if title == "" && uri == "" {
		return
	}
	c.Printf("[%s](%s)\n", title, uri)
	c.renderCaption(block)
}

// RenderTweet renders BlockTweet
func (c *Converter) RenderTweet(block *notionapi.Block) {
	c.RenderEmbed(block)
}

// RenderGist renders BlockGist
func (c *Converter) RenderGist(block *notionapi.Block) {
	c.RenderEmbed(block)
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

func (c *Converter) renderCaption(block *notionapi.Block) {
	caption := block.GetCaption()
	if caption == nil {
		return
	}
	c.Newline()
	c.RenderInlines(caption, false)
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

// RenderAudio renders BlockAudio
func (c *Converter) RenderAudio(block *notionapi.Block) {
	name, uri := getEmbeddedFileNameAndURL(block)
	c.Printf("[%s](%s)\n", name, uri)
	c.renderCaption(block)
}

// RenderVideo renders BlockTweet
func (c *Converter) RenderVideo(block *notionapi.Block) {
	name, uri := getEmbeddedFileNameAndURL(block)
	c.Printf("[%s](%s)\n", name, uri)
	c.renderCaption(block)
}

// RenderFile renders BlockFile
func (c *Converter) RenderFile(block *notionapi.Block) {
	fileID := block.FileIDs[0]
	localFileName := localFileNameFromURL(fileID, block.Source)
	name := block.Title
	c.Printf("[%s](%s)\n", name, localFileName)
	c.renderCaption(block)
}

// RenderDrive renders BlockDrive
func (c *Converter) RenderDrive(block *notionapi.Block) {
	docURL, _ := block.PropAsString("format.drive_properties.url")
	title, _ := block.PropAsString("format.drive_properties.title")
	c.Printf("[%s](%s)\n", title, docURL)
	c.renderCaption(block)
}

// RenderPDF renders BlockPDF
func (c *Converter) RenderPDF(block *notionapi.Block) {
	name, uri := getEmbeddedFileNameAndURL(block)
	c.Printf("[%s](%s)\n", name, uri)
	c.renderCaption(block)
}

// RenderEmbed renders BlockEmbed
func (c *Converter) RenderEmbed(block *notionapi.Block) {
	uri := block.Source
	c.Printf("[%s](%s)\n", uri, uri)
	c.renderCaption(block)
}

// RenderFigma renders BlockFigma
func (c *Converter) RenderFigma(block *notionapi.Block) {
	c.RenderEmbed(block)
}

// RenderImage renders BlockImage
func (c *Converter) RenderImage(block *notionapi.Block) {
	// TODO: not sure if always has FileIDs
	if len(block.FileIDs) == 0 {
		c.WriteString("RenderImage when len(FileIDs) == 0 NYI\n")
	}
	source := block.Source // also present in block.Format.DisplaySource
	// source looks like: "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/e5470cfd-08f0-4fb8-8ec2-452ca1a3f05e/Schermafbeelding2018-06-19om09.52.45.png"
	var fileID string
	if len(block.FileIDs) > 0 {
		fileID = block.FileIDs[0]
	}
	parts := strings.Split(source, "/")
	fileName := parts[len(parts)-1]
	parts = strings.SplitN(fileName, ".", 2)
	ext := ""
	if len(parts) == 2 {
		fileName = parts[0]
		ext = "." + parts[1]
	}
	c.Printf("![](%s-%s%s)\n", fileName, fileID, ext)
	c.renderCaption(block)
}

// RenderColumnList renders BlockColumnList
// it's children are BlockColumn
func (c *Converter) RenderColumnList(block *notionapi.Block) {
	c.RenderChildren(block)
}

// RenderColumn renders BlockColumn
// it's parent is BlockColumnList
func (c *Converter) RenderColumn(block *notionapi.Block) {
	c.RenderChildren(block)
}

// RenderCollectionView renders BlockCollectionView
func (c *Converter) RenderCollectionView(block *notionapi.Block) {
	if len(block.TableViews) == 0 {
		return
	}
	c.WriteString("RendeRenderCollectionViewrCode NYI\n")

	c.RenderChildren(block)
}

// DefaultRenderFunc returns a defult rendering function for a type of
// a given block
func (c *Converter) DefaultRenderFunc(blockType string) func(*notionapi.Block) {
	switch blockType {
	case notionapi.BlockPage:
		return c.RenderPage
	case notionapi.BlockText:
		return c.RenderText
	case notionapi.BlockEquation:
		// TODO: NYI
	case notionapi.BlockNumberedList:
		return c.RenderNumberedList
	case notionapi.BlockBulletedList:
		return c.RenderBulletedList
	case notionapi.BlockHeader:
		return c.RenderHeader
	case notionapi.BlockSubHeader:
		return c.RenderSubHeader
	case notionapi.BlockSubSubHeader:
		return c.RenderSubSubHeader
	case notionapi.BlockTodo:
		return c.RenderTodo
	case notionapi.BlockToggle:
		return c.RenderToggle
	case notionapi.BlockQuote:
		return c.RenderQuote
	case notionapi.BlockDivider:
		return c.RenderDivider
	case notionapi.BlockCode:
		return c.RenderCode
	case notionapi.BlockBookmark:
		return c.RenderBookmark
	case notionapi.BlockImage:
		return c.RenderImage
	case notionapi.BlockColumnList:
		return c.RenderColumnList
	case notionapi.BlockColumn:
		return c.RenderColumn
	case notionapi.BlockCollectionView:
		return c.RenderCollectionView
	case notionapi.BlockCollectionViewPage:
		// TODO: NYI
	case notionapi.BlockEmbed:
		return c.RenderEmbed
	case notionapi.BlockGist:
		return c.RenderGist
	case notionapi.BlockMaps:
		return c.RenderEmbed
	case notionapi.BlockCodepen:
		return c.RenderEmbed
	case notionapi.BlockTweet:
		return c.RenderEmbed
	case notionapi.BlockVideo:
		return c.RenderVideo
	case notionapi.BlockAudio:
		return c.RenderAudio
	case notionapi.BlockFile:
		return c.RenderFile
	case notionapi.BlockDrive:
		return c.RenderDrive
	case notionapi.BlockFigma:
		return c.RenderFigma
	case notionapi.BlockPDF:
		return c.RenderPDF
	case notionapi.BlockCallout:
		return c.RenderCallout
	case notionapi.BlockTableOfContents:
		// TODO: NYI
	case notionapi.BlockBreadcrumb:
		// TODO: NYI
	case notionapi.BlockFactory:
		return nil
	default:
		maybePanic("DefaultRenderFunc: unsupported block type '%s' in %s\n", blockType, c.Page.NotionURL())
	}
	return nil
}

func isEmptyBlock(block *notionapi.Block) bool {
	if block.Type == notionapi.BlockText {
		return len(block.InlineContent) == 0
	}
	if block.Type == notionapi.BlockPage {
		return !block.IsSubPage()
	}
	return false
}

func (c *Converter) skipChildren(block *notionapi.Block) bool {
	if len(block.Content) == 0 {
		// optimization for blocks that don't have children
		// even if they can have them
		return true
	}
	if block.Type == notionapi.BlockPage {
		// we don't want to render content of links to pages
		return !c.Page.IsRoot(block)
	}
	return false
}

func (c *Converter) RenderChildren(block *notionapi.Block) {
	if c.skipChildren(block) {
		return
	}
	currIdx := c.CurrBlockIdx
	currBlocks := c.CurrBlocks
	c.CurrBlocks = block.Content
	for i, child := range block.Content {
		child.Parent = block
		c.CurrBlockIdx = i
		c.RenderBlock(child)
	}
	c.CurrBlockIdx = currIdx
	c.CurrBlocks = currBlocks
}

func (c *Converter) AddNewlineBeforeBlock(block *notionapi.Block) {
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
		c.Newline()
	}
	if !isEmptyBlock(block) {
		if len(c.Indent) > 0 {
			c.WriteString(c.Indent)
		}
	}
}

// RenderBlock renders a block to html
func (c *Converter) RenderBlock(block *notionapi.Block) {
	if block == nil {
		// a missing block, can happen if we don't have access to a referenced block
		return
	}

	if c.RenderBlockOverride != nil && c.RenderBlockOverride(block) {
		return
	}

	def := c.DefaultRenderFunc(block.Type)
	if def != nil {
		c.AddNewlineBeforeBlock(block)
		def(block)
	}
}

func (c *Converter) ToMarkdown() []byte {
	c.PushNewBuffer()

	c.RenderBlock(c.Page.Root())
	buf := c.PopBuffer()
	// a bit of a hack to account for adding newlines before and after each block
	// which adds empty lines at top and bottom
	d := buf.Bytes()
	d = bytes.TrimSpace(d)
	return d
}

// ToMarkdown converts a page to Markdown
func ToMarkdown(page *notionapi.Page) []byte {
	r := NewConverter(page)
	return r.ToMarkdown()
}
