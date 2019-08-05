package tohtml2

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

func maybePanicIfErr(err error, format string, args ...interface{}) {
	if err != nil {
		notionapi.MaybePanic(format, args...)
	}
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

func fileNameFromPageCoverURL(uri string) string {
	parts := strings.Split(uri, "/")
	lastIdx := len(parts) - 1
	return parts[lastIdx]
}

func filePathFromPageCoverURL(uri string, block *notionapi.Block) string {
	// TODO: not sure about this heuristic. Maybe turn it into a whitelist:
	// if starts with notion.so or aws, then download and convert to local
	// otherwise leave alone
	if strings.HasPrefix(uri, "https://cdn.dutchcowboys.nl/uploads") {
		return uri
	}
	if strings.HasPrefix(uri, "https://images.unsplash.com") {
		return uri
	}
	if strings.HasPrefix(uri, "https://www.notion.so/images/") {
		return uri
	}
	if strings.HasPrefix(uri, "/images/page-cover/") {
		return "https://www.notion.so" + uri
	}
	fileName := fileNameFromPageCoverURL(uri)
	// TODO: probably need to build mulitple dirs
	dir := safeName(block.Title)
	return path.Join(dir, fileName)
}

func filePathForPage(block *notionapi.Block) string {
	name := safeName(block.Title) + ".html"
	for block.Parent != nil {
		block = block.Parent
		if block.Type != notionapi.BlockPage {
			continue
		}
		name = safeName(block.Title) + "/" + name
	}
	return name
}

func urlBaseName(uri string) string {
	parts := strings.Split(uri, "/")
	return parts[len(parts)-1]
}

func getTitleColDownloadedURL(row *notionapi.Block, block *notionapi.Block, collection *notionapi.Collection) string {
	title := ""
	titleSpans := row.GetTitle()
	if len(titleSpans) == 0 {
		log("title is empty)")
	} else {
		title = titleSpans[0].Text
	}
	if title == "" {
		title = "Untitled"
	}
	name := safeName(title) + ".html"
	colName := collection.GetName()
	if colName == "" {
		colName = "Untitled Database"
	}
	name = safeName(colName) + "/" + name
	for block.Parent != nil {
		block = block.Parent
		if block.Type != notionapi.BlockPage {
			continue
		}
		name = safeName(block.Title) + "/" + name
	}
	return name
}

func getDownloadedFileName(uri string, block *notionapi.Block) string {
	name := urlBaseName(uri)
	switch block.Type {
	case notionapi.BlockFile:
	// do nothing
	default:
		name = safeName(block.Title) + "/" + name
	}
	var parents []string
	tmp := block
	for tmp.Parent != nil {
		tmp = tmp.Parent
		if tmp.Type != notionapi.BlockPage {
			continue
		}
		parents = append(parents, safeName(tmp.Title))
	}

	for _, s := range parents {
		name = s + "/" + name
	}

	for strings.Contains(name, "//") {
		name = strings.Replace(name, "//", "/", -1)
	}
	return name
}

func getFileOrSourceURL(block *notionapi.Block) string {
	if len(block.FileIDs) > 0 {
		return getDownloadedFileName(block.Source, block)
	}
	return block.Source
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
type Converter struct {
	// Buf is where HTML is being written to
	Buf  *bytes.Buffer
	Page *notionapi.Page

	// tracks current number of numbered lists
	ListNo int

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

	// if true, generates stand-alone HTML with inline CSS
	// otherwise it's just the inner part going inside the body
	FullHTML bool

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

// NewConverter returns customizable HTML renderer
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

var (
	// true only when debugging
	doNewline = false
)

// Newline writes a newline to the buffer. It'll suppress multiple newlines.
func (c *Converter) Newline() {
	if !doNewline {
		return
	}
	d := c.Buf.Bytes()
	n := len(d)
	if n > 0 && d[n-1] != '\n' {
		c.Buf.WriteByte('\n')
	}
}

func (c *Converter) Printf(format string, args ...interface{}) {
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	if doNewline {
		if strings.HasPrefix(s, `<`) {
			c.Newline()
		}
	}
	c.Buf.WriteString(s)
}

// A writes <a></a> element to output
func (c *Converter) A(uri, text, cls string) {
	// TODO: Notion seems to encode url but it's probably not correct
	// (it encodes "&" as "&amp;")
	// at best should only encoede as url
	uri = escapeHTML(uri)
	text = escapeHTML(text)
	if cls != "" {
		cls = fmt.Sprintf(` class="%s"`, cls)
	}
	if uri == "" {
		c.Printf(`<a%s>%s</a>`, cls, text)
		return
	}
	c.Printf(`<a%s href="%s">%s</a>`, cls, uri, text)
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

// FormatDate formats the data
func (c *Converter) FormatDate(d *notionapi.Date) string {
	// TODO: allow over-riding date formatting
	s := notionapi.FormatDate(d)
	return fmt.Sprintf(`<time>@%s</time>`, s)
}

// RenderInline renders inline block
func (c *Converter) RenderInline(b *notionapi.TextSpan) {
	var start, close string
	text := b.Text
	for i := range b.Attrs {
		attr := b.Attrs[len(b.Attrs)-i-1]
		switch notionapi.AttrGetType(attr) {
		case notionapi.AttrHighlight:
			// TODO: possibly needs to change b.Highlight
			hl := notionapi.AttrGetHighlight(attr)
			start += fmt.Sprintf(`<mark class="highlight-%s">`, hl)
			close = `</mark>` + close
		case notionapi.AttrBold:
			start += `<strong>`
			close = `</strong>` + close
		case notionapi.AttrItalic:
			start += `<em>`
			close = `</em>` + close
		case notionapi.AttrStrikeThrought:
			start += `<del>`
			close = `</del>` + close
		case notionapi.AttrCode:
			start += `<code>`
			close = `</code>` + close
		case notionapi.AttrPage:
			pageID := notionapi.AttrGetPageID(attr)
			// TODO: find the page
			// TODO: needs to download info when recursively scanning
			// for pages
			pageTitle := ""
			urlPrefix := ""
			if pageTitle != "" {
				urlPrefix = safeName(pageTitle) + "-"
			}
			uri := "https://www.notion.so/" + urlPrefix + pageID
			if c.RewriteURL != nil {
				uri = c.RewriteURL(uri)
			}
			start += fmt.Sprintf(`<a href="%s">%s</a>`, uri, escapeHTML(pageTitle))
		case notionapi.AttrLink:
			uri := notionapi.AttrGetLink(attr)
			if c.RewriteURL != nil {
				uri = c.RewriteURL(uri)
			}
			if uri == "" {
				start += `<a>`
			} else {
				// TODO: notion escapes url but it seems to be wrong
				uri = escapeHTML(uri)
				start += fmt.Sprintf(`<a href="%s">`, uri)
			}
			close = `</a>` + close
		case notionapi.AttrUser:
			userID := notionapi.AttrGetUserID(attr)
			userName := notionapi.ResolveUser(c.Page, userID)
			start += fmt.Sprintf(`<span class="user">@%s</span>`, userName)
			text = ""
		case notionapi.AttrDate:
			date := notionapi.AttrGetDate(attr)
			start += c.FormatDate(date)
			text = ""
		}
	}
	c.Printf(start + escapeHTML(text) + close)
}

// RenderInlines renders inline blocks
func (c *Converter) RenderInlines(blocks []*notionapi.TextSpan) {
	for _, block := range blocks {
		c.RenderInline(block)
	}
}

// GetInlineContent is like RenderInlines but instead of writing to
// output buffer, we return it as string
func (c *Converter) GetInlineContent(blocks []*notionapi.TextSpan) string {
	if len(blocks) == 0 {
		return ""
	}
	c.PushNewBuffer()
	for _, block := range blocks {
		c.RenderInline(block)
	}
	return c.PopBuffer().String()
}

// RenderCode renders BlockCode
func (c *Converter) RenderCode(block *notionapi.Block) {
	cls := "code"
	c.Printf(`<pre id="%s" class="%s">`, block.ID, cls)
	{
		code := escapeHTML(block.Code)
		c.Printf(`<code>%s</code>`, code)
	}
	c.Printf("</pre>")
}

func escapeHTML(s string) string {
	s = html.EscapeString(s)
	// don't get why this is needed but it happens in
	// https://www.notion.so/Blendle-s-Employee-Handbook-3b617da409454a52bc3a920ba8832bf7
	s = strings.Replace(s, "&#39;", "&#x27;", -1)
	s = strings.Replace(s, "&#34;", "&quot;", -1)
	//s = strings.Replace(s, "&#x27;", "'", -1)
	return s
}

func (c *Converter) renderHeader(block *notionapi.Block) {
	c.Printf(`<header>`)
	{
		formatPage := block.FormatPage()
		// formatPage == nil happened in bf5d1c1f793a443ca4085cc99186d32f
		if formatPage != nil && formatPage.PageCover != "" {
			position := (1 - formatPage.PageCoverPosition) * 100
			coverURL := filePathFromPageCoverURL(formatPage.PageCover, block)
			// TODO: Notion incorrectly escapes them
			coverURL = escapeHTML(coverURL)
			c.Printf(`<img class="page-cover-image" src="%s" style="object-position:center %v%%">`, coverURL, position)
		}
		if formatPage != nil && formatPage.PageIcon != "" {
			// TODO: "undefined" is bad in Notion export
			clsCover := "undefined"
			if formatPage.PageCover != "" {
				clsCover = "page-header-icon-with-cover"
			}
			c.Printf(`<div class="page-header-icon %s">`, clsCover)
			if len(block.FileIDs) > 0 {
				fileName := getDownloadedFileName(formatPage.PageIcon, block)
				c.Printf(`<img class="icon" src="%s">`, fileName)
			} else {
				c.Printf(`<span class="icon">%s</span>`, formatPage.PageIcon)
			}
			c.Printf(`</div>`)
		}

		c.Printf(`<h1 class="page-title">`)
		{
			c.RenderInlines(block.InlineContent)
		}
		c.Printf(`</h1>`)
	}
	c.Printf(`</header>`)
}

// RenderCollectionViewPage renders BlockCollectionViewPage
func (c *Converter) RenderCollectionViewPage(block *notionapi.Block) {
	// TODO: grab collection by id
	// use "icon" for img url, "name" for src link
	/*
		<figure id="9c051067-c117-4b1e-a61a-70735c0494eb" class="link-to-page">
		<a href="Notion Pok dex/Type Chart.html"
		><img
		class="icon"
		src="Notion Pok dex/Type Chart/sticker375x360.u2.png"
		/>Type Chart</a
		>
	*/
	c.renderLinkToPage(block)
}

func (c *Converter) renderLinkToPage(block *notionapi.Block) {
	uri := filePathForPage(block)
	cls := appendClass(getBlockColorClass(block), "link-to-page")
	c.Printf(`<figure id="%s" class="%s">`, block.ID, cls)
	{
		c.Printf(`<a href="%s">`, uri)
		f := block.FormatPage()
		if f != nil && f.PageIcon != "" {
			if len(block.FileIDs) > 0 {
				fileName := getDownloadedFileName(f.PageIcon, block)
				c.Printf(`<img class="icon" src="%s">`, fileName)
			} else {
				c.Printf(`<span class="icon">%s</span>`, f.PageIcon)
			}
		}
		// TODO: possibly r.RenderInlines(block.InlineContent)
		c.Printf(escapeHTML(block.Title))
		c.Printf(`</a>`)
	}
	c.Printf(`</figure>`)
}

// RenderPage renders BlockPage
func (c *Converter) RenderPage(block *notionapi.Block) {
	tp := block.GetPageType()
	if tp == notionapi.BlockPageTopLevel {
		if c.FullHTML {
			c.Printf(`<html>`)
			{
				c.Printf(`<head>`)
				{
					c.Printf(`<meta http-equiv="Content-Type" content="text/html; charset=utf-8">`)
					c.Printf(`<title>%s</title>`, escapeHTML(block.Title))
					c.Printf("<style>%s\t\n</style>", cssNotion)
				}
				c.Printf(`</head>`)
			}
			c.Printf(`<body>`)
		}

		// TODO: sans could be mono, depending on format
		clsFont := "sans"
		fp := block.FormatPage()
		if fp != nil {
			if fp.PageFont != "" {
				clsFont = fp.PageFont
			}
		}
		c.Printf(`<article id="%s" class="page %s">`, block.ID, clsFont)
		c.renderHeader(block)
		{
			c.Printf(`<div class="page-body">`)
			c.RenderChildren(block)
			c.Printf(`</div>`)
		}
		c.Printf(`</article>`)

		if c.FullHTML {
			c.Printf(`</body></html>`)
		}
		return
	}

	if block.Parent != nil && block.Parent.Type == notionapi.BlockToggle {
		// TODO: seem like a bug in Notion exporter
		// page: https://www.notion.so/Soft-shizzle-13aa42a5a95d4357aa830c3e7ff35ae1
		return
	}

	// TODO: fixes some pages, breaks some other pages
	if false && block.Parent != nil && block.Parent.Type == notionapi.BlockPage {
		// TODO: seem like a bug in Notion exporter
		// page: https://www.notion.so/b1b31f6d3405466c988676f996ce03ad
		return
	}
	c.renderLinkToPage(block)
}

func appendClass(s, cls string) string {
	if len(s) == 0 {
		return cls
	}
	return s + " " + cls
}

func getBlockColorClass(block *notionapi.Block) string {
	// TODO: inefficient, use FromatStringValue("block_color") instead
	var col string
	if block.FormatText() != nil {
		col = block.FormatText().BlockColor
	} else if block.FormatPage() != nil {
		col = block.FormatPage().BlockColor
	} else if block.FormatToggle() != nil {
		col = block.FormatToggle().BlockColor
	} else if block.FormatHeader() != nil {
		col = block.FormatHeader().BlockColor
	} else if block.FormatNumberedList() != nil {
		col = block.FormatNumberedList().BlockColor
	} else if block.FormatBulletedList() != nil {
		col = block.FormatBulletedList().BlockColor
	}
	if col == "" {
		return ""
	}
	return "block-color-" + col
}

// RenderText renders BlockText
func (c *Converter) RenderText(block *notionapi.Block) {
	cls := getBlockColorClass(block)
	c.Printf(`<p id="%s" class="%s">`, block.ID, cls)
	c.RenderInlines(block.InlineContent)
	c.Printf(`</p>`)
	c.RenderChildren(block)
}

// RenderNumberedList renders BlockNumberedList
func (c *Converter) RenderNumberedList(block *notionapi.Block) {
	isPrevSame := c.IsPrevBlockOfType(notionapi.BlockNumberedList)
	if isPrevSame {
		c.ListNo++
	} else {
		c.ListNo = 1
	}

	cls := getBlockColorClass(block)
	cls = appendClass(cls, "numbered-list")
	c.Printf(`<ol id="%s" class="%s" start="%d">`, block.ID, cls, c.ListNo)
	{
		c.Printf(`<li>`)
		{
			c.RenderInlines(block.InlineContent)
			c.RenderChildren(block)
		}
		c.Printf(`</li>`)
	}
	c.Printf(`</ol>`)
}

// RenderBulletedList renders BlockBulletedList
func (c *Converter) RenderBulletedList(block *notionapi.Block) {
	cls := getBlockColorClass(block)
	cls = appendClass(cls, "bulleted-list")
	c.Printf(`<ul id="%s" class="%s">`, block.ID, cls)
	{
		c.Printf(`<li>`)
		{
			c.RenderInlines(block.InlineContent)
			c.RenderChildren(block)
		}
		c.Printf(`</li>`)
	}
	c.Printf(`</ul>`)
}

// RenderHeaderLevel renders BlockHeader, SubHeader and SubSubHeader
func (c *Converter) RenderHeaderLevel(block *notionapi.Block, level int) {
	cls := getBlockColorClass(block)
	c.Printf(`<h%d id="%s" class="%s">`, level, block.ID, cls)
	id := block.ID
	if c.AddHeaderAnchor {
		c.Printf(`<a class="notion-header-anchor" href="#%s" aria-hidden="true"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 8 8"><path d="M5.88.03c-.18.01-.36.03-.53.09-.27.1-.53.25-.75.47a.5.5 0 1 0 .69.69c.11-.11.24-.17.38-.22.35-.12.78-.07 1.06.22.39.39.39 1.04 0 1.44l-1.5 1.5c-.44.44-.8.48-1.06.47-.26-.01-.41-.13-.41-.13a.5.5 0 1 0-.5.88s.34.22.84.25c.5.03 1.2-.16 1.81-.78l1.5-1.5c.78-.78.78-2.04 0-2.81-.28-.28-.61-.45-.97-.53-.18-.04-.38-.04-.56-.03zm-2 2.31c-.5-.02-1.19.15-1.78.75l-1.5 1.5c-.78.78-.78 2.04 0 2.81.56.56 1.36.72 2.06.47.27-.1.53-.25.75-.47a.5.5 0 1 0-.69-.69c-.11.11-.24.17-.38.22-.35.12-.78.07-1.06-.22-.39-.39-.39-1.04 0-1.44l1.5-1.5c.4-.4.75-.45 1.03-.44.28.01.47.09.47.09a.5.5 0 1 0 .44-.88s-.34-.2-.84-.22z"></path></svg></a>`, id)
	}
	c.RenderInlines(block.InlineContent)
	c.Printf(`</h%d>`, level)
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

// RenderTodo renders BlockTodo
func (c *Converter) RenderTodo(block *notionapi.Block) {
	c.Printf(`<ul id="%s" class="to-do-list">`, block.ID)
	{
		c.Printf(`<li>`)
		{
			cls := "checkbox-off"
			if block.IsChecked {
				cls = "checkbox-on"
			}
			c.Printf(`<div class="checkbox %s"></div>`, cls)

			cls = "to-do-children-unchecked"
			if block.IsChecked {
				cls = "to-do-children-checked"
			}
			c.Printf(`<span class="%s">`, cls)
			c.RenderInlines(block.InlineContent)
			c.Printf(`</span>`)

			c.RenderChildren(block)
		}
		c.Printf(`</li>`)
	}
	c.Printf(`</ul>`)
}

// RenderToggle renders BlockToggle
func (c *Converter) RenderToggle(block *notionapi.Block) {
	cls := getBlockColorClass(block)
	cls = appendClass(cls, "toggle")
	c.Printf(`<ul id="%s" class="%s">`, block.ID, cls)
	{
		c.Printf(`<li>`)
		{
			c.Printf(`<details open="">`)
			{
				c.Printf(`<summary>`)
				c.RenderInlines(block.InlineContent)
				c.Printf(`</summary>`)
				c.RenderChildren(block)
			}
			c.Printf(`</details>`)
		}
		c.Printf(`</li>`)
	}
	c.Printf(`</ul>`)
}

// RenderQuote renders BlockQuote
func (c *Converter) RenderQuote(block *notionapi.Block) {
	c.Printf(`<blockquote id="%s" class="">`, block.ID)
	{
		c.RenderInlines(block.InlineContent)
		// TODO: do they have children?
		c.RenderChildren(block)
	}
	c.Printf(`</blockquote>`)
}

// RenderCallout renders BlockCallout
func (c *Converter) RenderCallout(block *notionapi.Block) {
	fmt.Printf(`<div class="notion-callout">`)
	fmt.Printf(`</div>`)
}

func isHeaderBlock(block *notionapi.Block) bool {
	switch block.Type {
	case notionapi.BlockHeader, notionapi.BlockSubHeader, notionapi.BlockSubSubHeader:
		return true
	}
	return false
}

func getHeaderBlocks(blocks []*notionapi.Block) []*notionapi.Block {
	var res []*notionapi.Block
	for _, b := range blocks {
		if isHeaderBlock(b) {
			res = append(res, b)
			continue
		}
		if len(b.Content) == 0 {
			continue
		}
		sub := getHeaderBlocks(b.Content)
		res = append(res, sub...)
	}
	return res
}

// RenderTableOfContents renders BlockTableOfContents
func (c *Converter) RenderTableOfContents(block *notionapi.Block) {
	// TODO: block-color-gray comes from "format": { "block_color" }
	c.Printf(`<nav id="%s" class="block-color-gray table_of_contents">`, block.ID)
	blocks := getHeaderBlocks(c.Page.Root.Content)
	for _, b := range blocks {
		s := c.GetInlineContent(b.InlineContent)
		// TODO: "indent-0" might probably be differnt
		c.Printf(`<div class="table_of_contents-item table_of_contents-indent-0">`)
		{
			c.Printf(`<a class="table_of_contents-link" href="#%s">%s</a>`, b.ID, s)
		}
		c.Printf(`</div>`)
	}
	c.Printf(`</nav>`)
}

// RenderDivider renders BlockDivider
func (c *Converter) RenderDivider(block *notionapi.Block) {
	c.Printf(`<hr id="%s">`, block.ID)
}

func (c *Converter) RenderCaption(block *notionapi.Block) {
	caption := block.GetCaption()
	if caption == nil {
		return
	}
	c.Printf(`<figcaption>`)
	c.RenderInlines(caption)
	c.Printf(`</figcaption>`)
}

// RenderBookmark renders BlockBookmark
func (c *Converter) RenderBookmark(block *notionapi.Block) {
	c.Printf(`<figure id="%s">`, block.ID)
	{
		c.Printf(`<div class="bookmark source">`)
		{
			uri := block.Link
			text := block.Title
			c.A(uri, text, "")
			c.Printf(`<br>`)
			c.A(uri, uri, "bookmark-href")
		}
		c.Printf(`</div>`)
		c.RenderCaption(block)
	}
	c.Printf(`</figure>`)
}

// RenderVideo renders BlockVideo
func (c *Converter) RenderVideo(block *notionapi.Block) {
	c.Printf(`<figure id="%s">`, block.ID)
	{
		c.Printf(`<div class="source">`)
		{
			source := block.Source
			fileName := source
			if len(block.FileIDs) > 0 {
				fileName = getDownloadedFileName(source, block)
			}
			if source == "" {
				c.Printf(`<a></a>`)
			} else {
				c.A(fileName, source, "")
			}
		}
		c.Printf(`</div>`)
		c.RenderCaption(block)
	}
	c.Printf(`</figure>`)
}

// RenderTweet renders BlockTweet
func (c *Converter) RenderTweet(block *notionapi.Block) {
	c.Printf(`<figure id="%s">`, block.ID)
	{
		c.Printf(`<div class="source">`)
		{
			uri := block.Source
			c.A(uri, uri, "")
		}
		c.Printf(`</div>`)
		c.RenderCaption(block)
	}
	c.Printf(`</figure>`)
}

// RenderGist renders BlockGist
func (c *Converter) RenderGist(block *notionapi.Block) {
	c.Printf(`<figure id="%s">`, block.ID)
	{
		c.Printf(`<div class="source">`)
		uri := block.Source
		c.A(uri, uri, "")
		c.Printf(`</div>`)
		c.RenderCaption(block)
	}
	c.Printf(`</figure>`)
}

// RenderEmbed renders BlockEmbed
func (c *Converter) RenderEmbed(block *notionapi.Block) {
	c.Printf(`<figure id="%s">`, block.ID)
	{
		c.Printf(`<div class="source">`)
		{
			uri := getFileOrSourceURL(block)
			text := block.Source
			c.A(uri, text, "")
		}
		c.Printf(`</div>`)
		c.RenderCaption(block)
	}
	c.Printf(`</figure>`)
}

// RenderFile renders BlockFile, BlockDrive
func (c *Converter) RenderFile(block *notionapi.Block) {
	c.Printf(`<figure id="%s">`, block.ID)
	{
		c.Printf(`<div class="source">`)
		{
			uri := getDownloadedFileName(block.Source, block)
			c.A(uri, block.Source, "")
		}
		c.Printf(`</div>`)
		c.RenderCaption(block)
	}
	c.Printf(`</figure>`)
}

// RenderPDF renders BlockPDF
func (c *Converter) RenderPDF(block *notionapi.Block) {
	c.Printf(`<figure id="%s">`, block.ID)
	{
		c.Printf(`<div class="source">`)
		uri := getDownloadedFileName(block.Source, block)
		c.A(uri, block.Source, "")
		c.Printf(`</div>`)
		c.RenderCaption(block)
	}
	c.Printf(`</figure>`)
}

func getImageStyle(block *notionapi.Block) string {
	f := block.FormatImage()
	if f == nil || f.BlockWidth == 0 {
		return ""
	}
	return fmt.Sprintf(`style="width:%dpx" `, int(f.BlockWidth))
}

// RenderImage renders BlockImage
func (c *Converter) RenderImage(block *notionapi.Block) {
	c.Printf(`<figure id="%s" class="image">`, block.ID)
	{
		uri := getFileOrSourceURL(block)
		style := getImageStyle(block)
		c.Printf(`<a href="%s">`, uri)
		c.Printf(`<img %s src="%s">`, style, uri)
		c.Printf(`</a>`)

		c.RenderCaption(block)
	}
	c.Printf(`</figure>`)
}

// RenderColumnList renders BlockColumnList
// it's children are BlockColumn
func (c *Converter) RenderColumnList(block *notionapi.Block) {
	nColumns := len(block.Content)
	if nColumns == 0 {
		maybePanic("has no columns")
		return
	}
	c.Printf(`<div id="%s" class="column-list">`, block.ID)
	c.RenderChildren(block)
	c.Printf(`</div>`)
}

// RenderColumn renders BlockColumn
// it's parent is BlockColumnList
func (c *Converter) RenderColumn(block *notionapi.Block) {
	var colRatio float64 = 50
	fc := block.FormatColumn()
	if fc != nil {
		colRatio = fc.ColumnRatio * 100
	}
	c.Printf(`<div id="%s" style="width:%v%%" class="column">`, block.ID, colRatio)
	c.RenderChildren(block)
	c.Printf("</div>")
}

// RenderCollectionView renders BlockCollectionView
func (c *Converter) RenderCollectionView(block *notionapi.Block) {
	pageID := ""
	if c.Page != nil {
		pageID = notionapi.ToNoDashID(c.Page.ID)
	}

	if len(block.CollectionViews) == 0 {
		log("missing block.CollectionViews for block %s %s in page %s\n", block.ID, block.Type, pageID)
		return
	}
	viewInfo := block.CollectionViews[0]
	view := viewInfo.CollectionView
	collection := viewInfo.Collection
	if view.Format == nil {
		log("missing view.Format for block %s %s in page %s\n", block.ID, block.Type, pageID)
		return
	}

	columns := view.Format.TableProperties
	c.Printf(`<div id="%s" class="collection-content">`, block.ID)
	{
		name := collection.GetName()
		c.Printf(`<h4 class="collection-title">%s</h4>`, name)
		c.Printf(`<table class="collection-content">`)
		{
			c.Printf(`<thead>`)
			{
				c.Printf(`<tr>`)
				for _, col := range columns {
					colName := col.Property
					colInfo := viewInfo.Collection.CollectionSchema[colName]
					name := ""
					if colInfo != nil {
						name = colInfo.Name
						name = escapeHTML(name)
					}
					c.Printf(`<th>%s</th>`, name)
				}
				c.Printf(`</tr>`)
			}
			c.Printf(`</thead>`)

			c.Printf(`<tbody>`)
			{
				for _, row := range viewInfo.CollectionRows {
					c.Printf(`<tr id="%s">`, row.ID)
					props := row.Properties
					for _, col := range columns {
						colName := col.Property
						v := props[colName]
						inlineContent, err := notionapi.ParseTextSpans(v)
						maybePanicIfErr(err, "ParseTextSpans of '%v' failed with %s\n", v, err)
						colVal := c.GetInlineContent(inlineContent)
						colInfo := viewInfo.Collection.CollectionSchema[colName]
						if colInfo.Type == "title" {
							uri := getTitleColDownloadedURL(row, block, viewInfo.Collection)
							if colVal == "" {
								colVal = "Untitled"
							}
							colVal = fmt.Sprintf(`<a href="%s">%s</a>`, uri, colVal)
						} else if colInfo.Type == "multi_select" {
							vals := strings.Split(colVal, ",")
							s := ""
							for i := range vals {
								// TODO: Notion prints in reverse order
								idx := len(vals) - 1 - i
								v := escapeHTML(vals[idx])
								if v == "" {
									continue
								}
								s += fmt.Sprintf(`<span class="selected-value">%s</span>`, v)
							}
							colVal = s
						}
						colNameCls := escapeHTML(colName)
						c.Printf(`<td class="cell-%s">%s</td>`, colNameCls, colVal)
					}
					c.Printf("</tr>\n")
				}
			}
			c.Printf(`</tbody>`)
		}
		c.Printf(`</table>`)
	}
	c.Printf(`</div>`)
}

// DefaultRenderFunc returns a defult rendering function for a type of
// a given block
func (c *Converter) DefaultRenderFunc(blockType string) func(*notionapi.Block) {
	switch blockType {
	case notionapi.BlockPage:
		return c.RenderPage
	case notionapi.BlockText:
		return c.RenderText
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
		return c.RenderCollectionViewPage
	case notionapi.BlockEmbed:
		return c.RenderEmbed
	case notionapi.BlockGist:
		return c.RenderGist
	case notionapi.BlockTweet:
		return c.RenderTweet
	case notionapi.BlockVideo:
		return c.RenderVideo
	case notionapi.BlockFile, notionapi.BlockDrive:
		return c.RenderFile
	case notionapi.BlockPDF:
		return c.RenderPDF
	case notionapi.BlockCallout:
		return c.RenderCallout
	case notionapi.BlockTableOfContents:
		return c.RenderTableOfContents
	default:
		maybePanic("DefaultRenderFunc: unsupported block type '%s' in %s\n", blockType, c.Page.NotionURL())
	}
	return nil
}

func needsIndent(block *notionapi.Block) bool {
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

func (c *Converter) RenderChildren(block *notionapi.Block) {
	if len(block.Content) == 0 {
		return
	}

	doIndent := needsIndent(block)
	// provides indentation for children
	if doIndent {
		c.Printf(`<div class="indented">`)
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

	if doIndent {
		c.Printf(`</div>`)
	}
}

// RenderBlock renders a block to html
func (c *Converter) RenderBlock(block *notionapi.Block) {
	if block == nil {
		// a missing block is possible
		return
	}
	if c.RenderBlockOverride != nil {
		handled := c.RenderBlockOverride(block)
		if handled {
			return
		}
	}
	def := c.DefaultRenderFunc(block.Type)
	if def != nil {
		def(block)
	}
}

// ToHTML renders a page to html
func (c *Converter) ToHTML() []byte {
	c.PushNewBuffer()
	c.RenderBlock(c.Page.Root)
	buf := c.PopBuffer()
	return buf.Bytes()
}

// ToHTML converts a page to HTML
func ToHTML(page *notionapi.Page) []byte {
	r := NewConverter(page)
	return r.ToHTML()
}
