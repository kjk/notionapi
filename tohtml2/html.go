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

	// Level is current depth of the tree. Useuful for pretty-printing indentation
	Level int

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

var (
	// true only when debugging
	doNewline = false
)

// Newline writes a newline to the buffer. It'll suppress multiple newlines.
func (r *HTMLRenderer) Newline() {
	if !doNewline {
		return
	}
	d := r.Buf.Bytes()
	n := len(d)
	if n > 0 && d[n-1] != '\n' {
		r.Buf.WriteByte('\n')
	}
}

func (r *HTMLRenderer) Printf(format string, args ...interface{}) {
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	if doNewline {
		if strings.HasPrefix(s, `<`) {
			r.Newline()
		}
	}
	r.Buf.WriteString(s)
}

// A writes <a></a> element to output
func (r *HTMLRenderer) A(uri, text, cls string) {
	// TODO: Notion seems to encode url but it's probably not correct
	// (it encodes "&" as "&amp;")
	// at best should only encoede as url
	uri = escapeHTML(uri)
	text = escapeHTML(text)
	if cls != "" {
		cls = fmt.Sprintf(` class="%s"`, cls)
	}
	if uri == "" {
		r.Printf(`<a%s>%s</a>`, cls, text)
		return
	}
	r.Printf(`<a%s href="%s">%s</a>`, cls, uri, text)
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
func (r *HTMLRenderer) RenderInline(b *notionapi.TextSpan) {
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
			start += `<strike>`
			close = `</strike>` + close
		case notionapi.AttrCode:
			start += `<code>`
			close = `</code>` + close
		case notionapi.AttrLink:
			uri := notionapi.AttrGetLink(attr)
			if r.RewriteURL != nil {
				uri = r.RewriteURL(uri)
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
			start += fmt.Sprintf(`<span class="notion-user">@TODO: user with id%s</span>`, userID)
			text = ""
		case notionapi.AttrDate:
			date := notionapi.AttrGetDate(attr)
			start += r.FormatDate(date)
			text = ""
		}
	}
	r.Printf(start + escapeHTML(text) + close)
}

// RenderInlines renders inline blocks
func (r *HTMLRenderer) RenderInlines(blocks []*notionapi.TextSpan) {
	r.Level++
	for _, block := range blocks {
		r.RenderInline(block)
	}
	r.Level--
}

// GetInlineContent is like RenderInlines but instead of writing to
// output buffer, we return it as string
func (r *HTMLRenderer) GetInlineContent(blocks []*notionapi.TextSpan) string {
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
	r.Printf(s)

	r.Printf("</code></pre>")
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

func (r *HTMLRenderer) renderHeader(block *notionapi.Block) {
	r.Printf(`<header>`)
	{
		formatPage := block.FormatPage
		// formatPage == nil happened in bf5d1c1f793a443ca4085cc99186d32f
		if formatPage != nil && formatPage.PageCover != "" {
			position := (1 - formatPage.PageCoverPosition) * 100
			coverURL := filePathFromPageCoverURL(formatPage.PageCover, block)
			// TODO: Notion incorrectly escapes them
			coverURL = escapeHTML(coverURL)
			r.Printf(`<img class="page-cover-image" src="%s" style="object-position:center %v%%">`, coverURL, position)
		}
		if formatPage != nil && formatPage.PageIcon != "" {
			// TODO: "undefined" is bad in Notion export
			clsCover := "undefined"
			if formatPage.PageCover != "" {
				clsCover = "page-header-icon-with-cover"
			}
			r.Printf(`<div class="page-header-icon %s">`, clsCover)
			r.Printf(`<span class="icon">%s</span>`, formatPage.PageIcon)
			r.Printf(`</div>`)
		}

		r.Printf(`<h1 class="page-title">`)
		{
			r.RenderInlines(block.InlineContent)
		}
		r.Printf(`</h1>`)
	}
	r.Printf(`</header>`)
}

// RenderPage renders BlockPage
func (r *HTMLRenderer) RenderPage(block *notionapi.Block) {
	tp := block.GetPageType()
	if tp == notionapi.BlockPageTopLevel {
		if r.FullHTML {
			r.Printf(`<html>`)
			{
				r.Printf(`<head>`)
				{
					r.Printf(`<meta http-equiv="Content-Type" content="text/html; charset=utf-8">`)
					r.Printf(`<title>%s</title>`, escapeHTML(block.Title))
					r.Printf("<style>%s\t\n</style>", cssNotion)
				}
				r.Printf(`</head>`)
			}
			r.Printf(`<body>`)
		}

		// TODO: sans could be mono, depending on format
		r.Printf(`<article id="%s" class="page sans">`, block.ID)
		r.renderHeader(block)
		{
			r.Printf(`<div class="page-body">`)
			r.RenderChildren(block)
			r.Printf(`</div>`)
		}
		r.Printf(`</article>`)

		if r.FullHTML {
			r.Printf(`</body></html>`)
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

	// Blendle s Employee Handbook/To Do Read in your first week.html
	uri := filePathForPage(block)
	cls := appendClass(getBlockColorClass(block), "link-to-page")
	r.Printf(`<figure id="%s" class="%s">`, block.ID, cls)
	{
		r.Printf(`<a href="%s">`, uri)
		if block.FormatPage != nil && block.FormatPage.PageIcon != "" {
			r.Printf(`<span class="icon">%s</span>`, block.FormatPage.PageIcon)
		}
		// TODO: possibly r.RenderInlines(block.InlineContent)
		r.Printf(escapeHTML(block.Title))
		r.Printf(`</a>`)
	}
	r.Printf(`</figure>`)
}

func appendClass(s, cls string) string {
	if len(s) == 0 {
		return cls
	}
	return s + " " + cls
}

func getBlockColorClass(block *notionapi.Block) string {
	var col string
	if block.FormatText != nil {
		col = block.FormatText.BlockColor
	} else if block.FormatPage != nil {
		col = block.FormatPage.BlockColor
	} else if block.FormatToggle != nil {
		col = block.FormatToggle.BlockColor
	} else if block.FormatHeader != nil {
		col = block.FormatHeader.BlockColor
	} else if block.FormatList != nil {
		col = block.FormatList.BlockColor
	}
	if col != "" {
		return "block-color-" + col
	}
	return ""
}

// RenderText renders BlockText
func (r *HTMLRenderer) RenderText(block *notionapi.Block) {
	cls := getBlockColorClass(block)
	r.Printf(`<p id="%s" class="%s">`, block.ID, cls)
	r.RenderInlines(block.InlineContent)
	r.Printf(`</p>`)
	r.RenderChildren(block)
}

// RenderNumberedList renders BlockNumberedList
func (r *HTMLRenderer) RenderNumberedList(block *notionapi.Block) {
	isPrevSame := r.IsPrevBlockOfType(notionapi.BlockNumberedList)
	if isPrevSame {
		r.ListNo++
	} else {
		r.ListNo = 1
	}

	cls := getBlockColorClass(block)
	cls = appendClass(cls, "numbered-list")
	r.Printf(`<ol id="%s" class="%s" start="%d">`, block.ID, cls, r.ListNo)
	{
		r.Printf(`<li>`)
		{
			r.RenderInlines(block.InlineContent)
			r.RenderChildren(block)
		}
		r.Printf(`</li>`)
	}
	r.Printf(`</ol>`)
}

// RenderBulletedList renders BlockBulletedList
func (r *HTMLRenderer) RenderBulletedList(block *notionapi.Block) {
	cls := getBlockColorClass(block)
	cls = appendClass(cls, "bulleted-list")
	r.Printf(`<ul id="%s" class="%s">`, block.ID, cls)
	{
		r.Printf(`<li>`)
		{
			r.RenderInlines(block.InlineContent)
			r.RenderChildren(block)
		}
		r.Printf(`</li>`)
	}
	r.Printf(`</ul>`)
}

// RenderHeaderLevel renders BlockHeader, SubHeader and SubSubHeader
func (r *HTMLRenderer) RenderHeaderLevel(block *notionapi.Block, level int) {
	cls := getBlockColorClass(block)
	r.Printf(`<h%d id="%s" class="%s">`, level, block.ID, cls)
	id := block.ID
	if r.AddHeaderAnchor {
		r.Printf(`<a class="notion-header-anchor" href="#%s" aria-hidden="true"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 8 8"><path d="M5.88.03c-.18.01-.36.03-.53.09-.27.1-.53.25-.75.47a.5.5 0 1 0 .69.69c.11-.11.24-.17.38-.22.35-.12.78-.07 1.06.22.39.39.39 1.04 0 1.44l-1.5 1.5c-.44.44-.8.48-1.06.47-.26-.01-.41-.13-.41-.13a.5.5 0 1 0-.5.88s.34.22.84.25c.5.03 1.2-.16 1.81-.78l1.5-1.5c.78-.78.78-2.04 0-2.81-.28-.28-.61-.45-.97-.53-.18-.04-.38-.04-.56-.03zm-2 2.31c-.5-.02-1.19.15-1.78.75l-1.5 1.5c-.78.78-.78 2.04 0 2.81.56.56 1.36.72 2.06.47.27-.1.53-.25.75-.47a.5.5 0 1 0-.69-.69c-.11.11-.24.17-.38.22-.35.12-.78.07-1.06-.22-.39-.39-.39-1.04 0-1.44l1.5-1.5c.4-.4.75-.45 1.03-.44.28.01.47.09.47.09a.5.5 0 1 0 .44-.88s-.34-.2-.84-.22z"></path></svg></a>`, id)
	}
	r.RenderInlines(block.InlineContent)
	r.Printf(`</h%d>`, level)
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
	r.Printf(`<ul id="%s" class="to-do-list">`, block.ID)
	{
		r.Printf(`<li>`)
		{
			cls := "checkbox-off"
			if block.IsChecked {
				cls = "checkbox-on"
			}
			r.Printf(`<div class="checkbox %s"></div>`, cls)

			cls = "to-do-children-unchecked"
			if block.IsChecked {
				cls = "to-do-children-checked"
			}
			r.Printf(`<span class="%s">`, cls)
			r.RenderInlines(block.InlineContent)
			r.Printf(`</span>`)

			r.RenderChildren(block)
		}
		r.Printf(`</li>`)
	}
	r.Printf(`</ul>`)
}

// RenderToggle renders BlockToggle
func (r *HTMLRenderer) RenderToggle(block *notionapi.Block) {
	cls := getBlockColorClass(block)
	cls = appendClass(cls, "toggle")
	r.Printf(`<ul id="%s" class="%s">`, block.ID, cls)
	{
		r.Printf(`<li>`)
		{
			r.Printf(`<details open="">`)
			{
				r.Printf(`<summary>`)
				r.RenderInlines(block.InlineContent)
				r.Printf(`</summary>`)
				r.RenderChildren(block)
			}
			r.Printf(`</details>`)
		}
		r.Printf(`</li>`)
	}
	r.Printf(`</ul>`)
}

// RenderQuote renders BlockQuote
func (r *HTMLRenderer) RenderQuote(block *notionapi.Block) {
	r.Printf(`<blockquote id="%s" class="">`, block.ID)
	{
		r.RenderInlines(block.InlineContent)
		// TODO: do they have children?
		r.RenderChildren(block)
	}
	r.Printf(`</blockquote>`)
}

// RenderCallout renders BlockCallout
func (r *HTMLRenderer) RenderCallout(block *notionapi.Block) {
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
func (r *HTMLRenderer) RenderTableOfContents(block *notionapi.Block) {
	// TODO: block-color-gray comes from "format": { "block_color" }
	r.Printf(`<nav id="%s" class="block-color-gray table_of_contents">`, block.ID)
	blocks := getHeaderBlocks(r.Page.Root.Content)
	for _, b := range blocks {
		s := r.GetInlineContent(b.InlineContent)
		// TODO: "indent-0" might probably be differnt
		r.Printf(`<div class="table_of_contents-item table_of_contents-indent-0">`)
		{
			r.Printf(`<a class="table_of_contents-link" href="#%s">%s</a>`, b.ID, s)
		}
		r.Printf(`</div>`)
	}
	r.Printf(`</nav>`)
}

// RenderDivider renders BlockDivider
func (r *HTMLRenderer) RenderDivider(block *notionapi.Block) {
	r.Printf(`<hr id="%s">`, block.ID)
}

// RenderBookmark renders BlockBookmark
func (r *HTMLRenderer) RenderBookmark(block *notionapi.Block) {
	r.Printf(`<figure id="%s">`, block.ID)
	{
		r.Printf(`<div class="bookmark source">`)
		{
			uri := block.Link
			text := block.Title
			r.A(uri, text, "")
			r.Printf(`<br>`)
			r.A(uri, uri, "bookmark-href")
		}
		r.Printf(`</div>`)
	}
	r.Printf(`</figure>`)
}

// RenderVideo renders BlockVideo
func (r *HTMLRenderer) RenderVideo(block *notionapi.Block) {
	r.Printf(`<figure id="%s">`, block.ID)
	{
		r.Printf(`<div class="source">`)
		{
			source := block.Source
			fileName := source
			if len(block.FileIDs) > 0 {
				fileName = getImageFileName(r.Page, block)
			}
			if source == "" {
				r.Printf(`<a></a>`)
			} else {
				r.A(fileName, source, "")
			}
		}
		r.Printf(`</div>`)
	}
	r.Printf(`</figure>`)
}

// RenderTweet renders BlockTweet
func (r *HTMLRenderer) RenderTweet(block *notionapi.Block) {
	r.Printf(`<figure id="%s">`, block.ID)
	{
		r.Printf(`<div class="source">`)
		{
			uri := block.Source
			r.A(uri, uri, "")
		}
		r.Printf(`</div>`)
	}
	r.Printf(`</figure>`)
}

// RenderGist renders BlockGist
func (r *HTMLRenderer) RenderGist(block *notionapi.Block) {
	r.Printf(`<script class="notion-embed-gist" src="%s">`, block.Source+".js")
	r.Printf("</script>")
}

// RenderEmbed renders BlockEmbed
func (r *HTMLRenderer) RenderEmbed(block *notionapi.Block) {
	r.Printf(`<figure id="%s">`, block.ID)
	{
		r.Printf(`<div class="source">`)
		{
			uri := block.Source
			r.A(uri, uri, "")
		}
		r.Printf(`</div>`)
	}
	r.Printf(`</figure>`)
}

// RenderFile renders BlockFile, BlockDrive
func (r *HTMLRenderer) RenderFile(block *notionapi.Block) {
	r.Printf(`<figure id="%s">`, block.ID)
	{
		r.Printf(`<div class="source">`)
		{
			uri := getImageFileName(r.Page, block)
			r.A(uri, block.Source, "")
		}
		r.Printf(`</div>`)
	}
	r.Printf(`</figure>`)
}

// RenderPDF renders BlockPDF
func (r *HTMLRenderer) RenderPDF(block *notionapi.Block) {
	r.Printf(`<figure id="%s">`, block.ID)
	{
		r.Printf(`<div class="source">`)
		uri := getImageFileName(r.Page, block)
		r.A(uri, block.Source, "")
		r.Printf(`</div>`)
	}
	r.Printf(`</figure>`)
}

func getImageFileName(page *notionapi.Page, block *notionapi.Block) string {
	uri := block.Source
	parts := strings.Split(uri, "/")
	lastIdx := len(parts) - 1
	fileName := parts[lastIdx]
	pageName := safeName(page.Root.Title)
	return pageName + "/" + fileName
}

func getImageStyle(block *notionapi.Block) string {
	f := block.FormatImage
	if f == nil || f.BlockWidth == 0 {
		return ""
	}
	return fmt.Sprintf(`style="width:%dpx" `, int(f.BlockWidth))
}

// RenderImage renders BlockImage
func (r *HTMLRenderer) RenderImage(block *notionapi.Block) {
	r.Printf(`<figure id="%s" class="image">`, block.ID)
	{
		uri := getImageFileName(r.Page, block)
		style := getImageStyle(block)
		r.Printf(`<a href="%s">`, uri)
		r.Printf(`<img %s src="%s">`, style, uri)
		r.Printf(`</a>`)
	}
	r.Printf(`</figure>`)
}

// RenderColumnList renders BlockColumnList
// it's children are BlockColumn
func (r *HTMLRenderer) RenderColumnList(block *notionapi.Block) {
	nColumns := len(block.Content)
	if nColumns == 0 {
		maybePanic("has no columns")
		return
	}
	r.Printf(`<div id="%s" class="column-list">`, block.ID)
	r.RenderChildren(block)
	r.Printf(`</div>`)
}

// RenderColumn renders BlockColumn
// it's parent is BlockColumnList
func (r *HTMLRenderer) RenderColumn(block *notionapi.Block) {
	// TODO: get column ration from col.FormatColumn.ColumnRation, which is float 0...1
	// TODO: width probably depends on number of columns
	r.Printf(`<div id="%s" style="width:50%%" class="column">`, block.ID)
	r.RenderChildren(block)
	r.Printf("</div>")
}

// RenderCollectionView renders BlockCollectionView
func (r *HTMLRenderer) RenderCollectionView(block *notionapi.Block) {
	pageID := ""
	if r.Page != nil {
		pageID = notionapi.ToNoDashID(r.Page.ID)
	}

	if len(block.CollectionViews) == 0 {
		log("missing block.CollectionViews for block %s %s in page %s\n", block.ID, block.Type, pageID)
		return
	}
	viewInfo := block.CollectionViews[0]
	view := viewInfo.CollectionView
	if view.Format == nil {
		log("missing view.Format for block %s %s in page %s\n", block.ID, block.Type, pageID)
		return
	}
	columns := view.Format.TableProperties

	r.Printf("\n" + `<table class="notion-collection-view">` + "\n")

	// generate header row
	r.Level++
	r.Printf("<thead>\n")

	r.Level++
	r.Printf("<tr>\n")

	for _, col := range columns {
		colName := col.Property
		colInfo := viewInfo.Collection.CollectionSchema[colName]
		if colInfo != nil {
			name := colInfo.Name
			r.Level++
			r.Printf(`<th>` + html.EscapeString(name) + "</th>\n")
			r.Level--
		} else {
			r.Level++
			r.Printf(`<th>&nbsp;` + "</th>\n")
			r.Level--
		}
	}
	r.Printf("</tr>\n")

	r.Level--
	r.Printf("</thead>\n\n")

	r.Printf("<tbody>\n")

	for _, row := range viewInfo.CollectionRows {
		r.Level++
		r.Printf("<tr>\n")

		props := row.Properties
		for _, col := range columns {
			colName := col.Property
			v := props[colName]
			//fmt.Printf("inline: '%s'\n", fmt.Sprintf("%v", v))
			inlineContent, err := notionapi.ParseTextSpans(v)
			if err != nil {
				maybePanic("ParseTextSpans of '%v' failed with %s\n", v, err)
			}
			//pretty.Print(inlineContent)
			colVal := r.GetInlineContent(inlineContent)
			//fmt.Printf("colVal: '%s'\n", colVal)
			r.Level++
			//colInfo := viewInfo.Collection.CollectionSchema[colName]
			// TODO: format colVal according to colInfo
			r.Printf(`<td>` + colVal + `</td>`)
			r.Level--
		}
		r.Printf("</tr>\n")
		r.Level--
	}

	r.Printf("</tbody>\n")

	r.Level--
	r.Printf("</table>\n")
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
	case notionapi.BlockFile, notionapi.BlockDrive:
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

func (r *HTMLRenderer) RenderChildren(block *notionapi.Block) {
	if len(block.Content) == 0 {
		return
	}

	doIndent := needsIndent(block)
	// provides indentation for children
	if doIndent {
		r.Printf(`<div class="indented">`)
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

	if doIndent {
		r.Printf(`</div>`)
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
