package tohtml

import (
	"bytes"
	"fmt"
	"html"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/kjk/notionapi"
)

var (
	byPassPageCover bool
)

func maybePanic(format string, args ...interface{}) {
	notionapi.MaybePanic(format, args...)
}

func logf(format string, args ...interface{}) {
	notionapi.Logf(format, args...)
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

// ByPassPageCover when you do not want to check about the cover white list URL
func ByPassPageCover(b bool) {
	byPassPageCover = b
}

func FilePathFromPageCoverURL(uri string, block *notionapi.Block) string {
	// TODO: not sure about this heuristic. Maybe turn it into a whitelist:
	// if starts with notion.so or aws, then download and convert to local
	// otherwise leave alone
	if byPassPageCover {
		return uri
	}
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
	// TODO: probably need to build multiple dirs
	dir := safeName(block.Title)
	return path.Join(dir, fileName)
}

func filePathForPage(block *notionapi.Block) string {
	name := safeName(block.Title) + ".html"
	for block.Parent != nil && block != block.Parent {
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

func filePathForCollection(page *notionapi.Page, col *notionapi.Collection) string {
	name := safeName(col.GetName()) + ".html"
	name = safeName(page.Root().Title) + "/" + name
	return name
}

// title columns are links to pages. this generates a link to a page
func (c *Converter) tableTitleCellURL(tv *notionapi.TableView, row, col int) string {
	if c.TableTitleCellURLOverride != nil {
		return c.TableTitleCellURLOverride(tv, row, col)
	}
	title := ""
	titleSpans := tv.CellContent(row, col)
	if len(titleSpans) == 0 {
		logf("title is empty)")
	} else {
		title = titleSpans[0].Text
	}
	if title == "" {
		title = "Untitled"
	}
	name := safeName(title) + ".html"
	colName := tv.Collection.GetName()
	if colName == "" {
		colName = "Untitled Database"
	}
	name = safeName(colName) + "/" + name
	block := tv.Page.Root()
	for block.Parent != nil && block != block.Parent {
		block = block.Parent
		if block.Type != notionapi.BlockPage {
			continue
		}
		name = safeName(block.Title) + "/" + name
	}
	return name
}

func getCollectionDownloadedFileName(page *notionapi.Page, col *notionapi.Collection, uri string) string {
	name := urlBaseName(uri)
	name = safeName(col.GetName()) + "/" + name
	name = safeName(page.Root().Title) + "/" + name
	return name
}

func getDownloadedFileName(uri string, block *notionapi.Block) string {
	shouldDownload := false
	if strings.HasPrefix(uri, "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/") {
		shouldDownload = true
	}
	if !shouldDownload {
		return uri
	}
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
	return htmlFileName(page.Root().Title)
}

type PageByIDProvider interface {
	PageByID(id string) *notionapi.Page
}

var _ PageByIDProvider = &PageByIDFromPages{}

type PageByIDFromPages struct {
	pages    []*notionapi.Page
	idToPage map[string]*notionapi.Page
}

// NewPageByIDFromPages creates PageByIDProvider from array of pages
func NewPageByIDFromPages(pages []*notionapi.Page) *PageByIDFromPages {
	res := &PageByIDFromPages{
		pages: pages,
	}
	res.idToPage = map[string]*notionapi.Page{}
	for _, page := range pages {
		id := notionapi.ToDashID(page.ID)
		res.idToPage[id] = page
	}
	return res
}

func (p *PageByIDFromPages) PageByID(pageID string) *notionapi.Page {
	pageID = notionapi.ToDashID(pageID)
	return p.idToPage[pageID]
}

// BlockRenderFunc is a function for rendering a particular block
type BlockRenderFunc func(block *notionapi.Block) bool

// Converter converts a Page to HTML
type Converter struct {
	// Buf is where HTML is being written to
	Buf  *bytes.Buffer
	Page *notionapi.Page

	// tracks current number of numbered lists
	ListNo int

	// if true tries to render as closely to Notion's HTML
	// export as possible
	NotionCompat bool

	// UseKatexToRenderEquation requires katex CLI to be installed
	// https://katex.org/docs/cli.html
	// npm install -g katex
	// If true, converts BlockEquation to HTML using katex
	// Tested with katex 0.10.2
	UseKatexToRenderEquation bool

	// If UseKatexToRenderEquation is true, you can provide path to katex binary
	// here. Otherwise we'll try to locate it using exec.LookPath()
	// If UseKatexToRenderEquation is true but we can't locate katex binary
	// we'll return an error
	KatexPath string

	// if true, adds <a href="#{$NotionID}">svg(anchor-icon)</a>
	// to h1/h2/h3
	AddHeaderAnchor bool

	// allows over-riding rendering of specific blocks
	// return false for default rendering
	RenderBlockOverride BlockRenderFunc

	// RewriteURL allows re-writing URLs e.g. to convert inter-notion URLs
	// to destination URLs
	RewriteURL func(url string) string

	// Returns URL for a title cell (that links to a page)
	TableTitleCellURLOverride func(tv *notionapi.TableView, row, col int) string

	// if true, generates stand-alone HTML with inline CSS
	// otherwise it's just the inner part going inside the body
	FullHTML bool

	// we need this to properly render ordered and numbered lists
	CurrBlocks   []*notionapi.Block
	CurrBlockIdx int

	PageByIDProvider PageByIDProvider

	// data provided by they caller, useful when providing
	// RenderBlockOverride
	Data interface{}

	didImportKatexCSS bool
	bufs              []*bytes.Buffer
}

// NewConverter returns customizable HTML renderer
func NewConverter(page *notionapi.Page) *Converter {
	return &Converter{
		Page: page,
	}
}

// PageByID returns Page given its ID
func (c *Converter) PageByID(pageID string) *notionapi.Page {
	if c.PageByIDProvider != nil {
		return c.PageByIDProvider.PageByID(pageID)
	}
	return nil
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

func (c *Converter) Printf(format string, args ...interface{}) {
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	c.Buf.WriteString(s)
}

// A writes <a></a> element to output
func (c *Converter) A(uri, text, cls string) {
	// TODO: Notion seems to encode url but it's probably not correct
	// (it encodes "&" as "&amp;")
	// at best should only encoede as url
	uri = EscapeHTML(uri)
	text = EscapeHTML(text)
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

// RewrittenURL optionally transforms the url via the
// function provided by the user
func (c *Converter) RewrittenURL(uri string) string {
	if c.RewriteURL != nil {
		return c.RewriteURL(uri)
	}
	return uri
}

// RenderInline renders inline block
func (c *Converter) RenderInline(b *notionapi.TextSpan) {
	var start, end string
	text := b.Text
	for i := range b.Attrs {
		attr := b.Attrs[len(b.Attrs)-i-1]
		switch notionapi.AttrGetType(attr) {
		case notionapi.AttrHighlight:
			// TODO: possibly needs to change b.Highlight
			hl := notionapi.AttrGetHighlight(attr)
			start += fmt.Sprintf(`<mark class="highlight-%s">`, hl)
			end = `</mark>` + end
		case notionapi.AttrBold:
			start += `<strong>`
			end = `</strong>` + end
		case notionapi.AttrItalic:
			start += `<em>`
			end = `</em>` + end
		case notionapi.AttrStrikeThrought:
			start += `<del>`
			end = `</del>` + end
		case notionapi.AttrCode:
			start += `<code>`
			end = `</code>` + end
		case notionapi.AttrPage:
			pageID := notionapi.AttrGetPageID(attr)
			pageTitle := ""
			relURL := notionapi.ToNoDashID(pageID)
			block := c.Page.BlockByID(pageID)
			if block != nil {
				pageTitle = block.Title
			}
			if pageTitle != "" {
				urlName := safeName(pageTitle)
				urlName = strings.Replace(urlName, " ", "-", -1)
				relURL = urlName + "-" + relURL
			}
			uri := c.RewrittenURL("https://www.notion.so/" + relURL)
			start += fmt.Sprintf(`<a href="%s">%s</a>`, uri, EscapeHTML(pageTitle))
			text = ""
		case notionapi.AttrLink:
			uri := c.RewrittenURL(notionapi.AttrGetLink(attr))
			if uri == "" {
				start += `<a>`
			} else {
				// TODO: notion escapes url but it seems to be wrong
				uri = EscapeHTML(uri)
				start += fmt.Sprintf(`<a href="%s">`, uri)
			}
			end = `</a>` + end
		case notionapi.AttrUser:
			userID := notionapi.AttrGetUserID(attr)
			userName := notionapi.GetUserNameByID(c.Page, userID)
			start += fmt.Sprintf(`<span class="user">@%s</span>`, userName)
			text = ""
		case notionapi.AttrDate:
			date := notionapi.AttrGetDate(attr)
			start += c.FormatDate(date)
			text = ""
		}
	}
	c.Printf(start + EscapeHTML(text) + end)
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
	if !c.NotionCompat {
		lang := strings.ToLower(strings.TrimSpace(block.CodeLanguage))
		if lang != "" {
			cls += " lang-" + lang
		}
	}
	c.Printf(`<pre id="%s" class="%s">`, block.ID, cls)
	{
		code := EscapeHTML(block.Code)
		c.Printf(`<code>%s</code>`, code)
	}
	c.Printf("</pre>")
}

// EscapeHTML escapes HTML in the same way as Notion.
func EscapeHTML(s string) string {
	s = html.EscapeString(s)
	// don't get why this is needed but it happens in
	// https://www.notion.so/Blendle-s-Employee-Handbook-3b617da409454a52bc3a920ba8832bf7
	s = strings.Replace(s, "&#39;", "&#x27;", -1)
	s = strings.Replace(s, "&#34;", "&quot;", -1)
	return s
}

func isURL(uri string) bool {
	if strings.HasPrefix(uri, "http://") {
		return true
	}
	if strings.HasPrefix(uri, "https://") {
		return true
	}
	return false
}

func (c *Converter) renderPageHeader(block *notionapi.Block) {
	c.Printf(`<header>`)
	{
		formatPage := block.FormatPage()
		// formatPage == nil happened in bf5d1c1f793a443ca4085cc99186d32f
		pageCover, _ := block.PropAsString("format.page_cover")
		if pageCover != "" {
			position := (1 - formatPage.PageCoverPosition) * 100
			coverURL := FilePathFromPageCoverURL(pageCover, block)
			// TODO: Notion incorrectly escapes them
			coverURL = EscapeHTML(coverURL)
			c.Printf(`<img class="page-cover-image" src="%s" style="object-position:center %v%%"/>`, coverURL, position)
		}
		pageIcon, _ := block.PropAsString("format.page_icon")
		if pageIcon != "" {
			// TODO: "undefined" is a bug in Notion export
			clsCover := "undefined"
			if pageCover != "" {
				clsCover = "page-header-icon-with-cover"
			}
			c.Printf(`<div class="page-header-icon %s">`, clsCover)
			if isURL(pageIcon) {
				fileName := getDownloadedFileName(pageIcon, block)
				c.Printf(`<img class="icon" src="%s"/>`, fileName)
			} else {
				c.Printf(`<span class="icon">%s</span>`, pageIcon)
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
	colID := block.CollectionID
	col := c.Page.CollectionByID(colID)
	icon := col.Icon
	name := col.GetName()
	c.Printf(`<figure id="%s" class="link-to-page">`, block.ID)
	{
		filePath := filePathForCollection(c.Page, col)
		c.Printf(`<a href="%s">`, filePath)
		{
			uri := getCollectionDownloadedFileName(c.Page, col, icon)
			c.Printf(`<img class="icon" src="%s"/>`, uri)
		}
		// TODO: should name be inlines?
		c.Printf(`%s</a>`, name)
	}
	c.Printf(`</figure>`)
}

func (c *Converter) renderLinkToPageNotion(block *notionapi.Block) {
	uri := filePathForPage(block)
	cls := GetBlockColorClass(block) + " link-to-page"
	cls = CleanAttributeValue(cls)
	c.Printf(`<figure id="%s" class="%s">`, block.ID, cls)
	{
		c.Printf(`<a href="%s">`, uri)
		pageIcon, ok := block.PropAsString("format.page_icon")
		if ok {
			if isURL(pageIcon) {
				fileName := getDownloadedFileName(pageIcon, block)
				c.Printf(`<img class="icon" src="%s"/>`, fileName)
			} else {
				c.Printf(`<span class="icon">%s</span>`, pageIcon)
			}
		}
		// TODO: possibly r.RenderInlines(block.InlineContent)
		c.Printf(EscapeHTML(block.Title))
		c.Printf(`</a>`)
	}
	c.Printf(`</figure>`)
}

func (c *Converter) renderLinkToPage(block *notionapi.Block) {
	if c.NotionCompat {
		c.renderLinkToPageNotion(block)
		return
	}

	uri := filePathForPage(block)
	cls := GetBlockColorClass(block) + " link-to-page"
	cls = CleanAttributeValue(cls)
	c.Printf(`<div id="%s" class="%s">`, block.ID, cls)
	{

		c.Printf(`<a href="%s">`, uri)
		pageIcon, ok := block.PropAsString("format.page_icon")
		if ok {
			if isURL(pageIcon) {
				fileName := getDownloadedFileName(pageIcon, block)
				c.Printf(`<img class="icon" src="%s"/>`, fileName)
			} else {
				c.Printf(`<span class="icon">%s</span>`, pageIcon)
			}
		}
		// TODO: possibly r.RenderInlines(block.InlineContent)
		c.Printf(EscapeHTML(block.Title))
		c.Printf(`</a>`)
	}
	c.Printf(`</div>`)
}

func (c *Converter) renderRootPage(block *notionapi.Block) {
	if c.FullHTML {
		c.Printf(`<html>`)
		{
			c.Printf(`<head>`)
			{
				c.Printf(`<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>`)
				c.Printf(`<title>%s</title>`, EscapeHTML(block.Title))
				c.Printf("<style>%s\t\n</style>", CSS)
			}
			c.Printf(`</head>`)
		}
		c.Printf(`<body>`)
	}

	clsFont := "sans"
	fp := block.FormatPage()
	if fp != nil {
		if fp.PageFont != "" {
			clsFont = fp.PageFont
		}
	}
	c.Printf(`<article id="%s" class="page %s">`, block.ID, clsFont)
	c.renderPageHeader(block)
	{
		c.Printf(`<div class="page-body">`)
		c.RenderChildren(block)
		c.Printf(`</div>`)
	}
	c.Printf(`</article>`)

	if c.FullHTML {
		c.Printf(`</body></html>`)
	}
}

func (c *Converter) renderSubPage(block *notionapi.Block) {
	c.renderLinkToPage(block)
}

// RenderPage renders BlockPage
func (c *Converter) RenderPage(block *notionapi.Block) {
	if c.Page.IsRoot(block) {
		c.renderRootPage(block)
		return
	}

	if c.Page.IsSubPage(block) {
		c.renderSubPage(block)
	} else {
		c.renderLinkToPage(block)
	}
}

// GetBlockColorClass returns "block-color-" + format.block_color
// which is name of css class for different colors
func GetBlockColorClass(block *notionapi.Block) string {
	col, _ := block.PropAsString("format.block_color")
	if col == "" {
		return ""
	}
	return "block-color-" + col
}

// RenderText renders BlockText
func (c *Converter) RenderText(block *notionapi.Block) {
	cls := GetBlockColorClass(block)
	if c.NotionCompat {
		c.Printf(`<p id="%s" class="%s">`, block.ID, cls)
		c.RenderInlines(block.InlineContent)
		c.RenderChildren(block)
		c.Printf(`</p>`)
		return
	}
	c.Printf(`<div id="%s" class="%s">`, block.ID, cls)
	c.RenderInlines(block.InlineContent)
	c.RenderChildren(block)
	c.Printf(`</div>`)
}

func equationToHTML(katexPath string, equation string) (string, error) {
	cmd := exec.Command(katexPath, "-d")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Start()
	if err != nil {
		return "", err
	}
	_, err = stdin.Write([]byte(equation))
	if err != nil {
		_ = cmd.Process.Kill()
		return "", err
	}
	err = stdin.Close()
	if err != nil {
		return "", err
	}
	if err = cmd.Wait(); err != nil {
		return "", err
	}
	res := out.String()
	return res, nil
}

// RenderEquation renders BlockEquation
func (c *Converter) RenderEquation(block *notionapi.Block) {
	if !c.UseKatexToRenderEquation {
		c.Printf(`<figure id="%s" class="equation">`, block.ID)
		c.RenderInlines(block.InlineContent)
		c.Printf(`</figure>`)
		return
	}
	ts := block.InlineContent
	s := notionapi.TextSpansToString(ts)
	htmlStr, err := equationToHTML(c.KatexPath, s)
	if err != nil {
		c.Printf(`<figure id="%s" class="equation">`, block.ID)
		c.RenderInlines(block.InlineContent)
		c.Printf(`</figure>`)
		return
	}

	c.Printf(`<figure id="%s" class="equation">`, block.ID)
	{
		if !c.didImportKatexCSS {
			c.Printf(`<style>@import url('https://cdnjs.cloudflare.com/ajax/libs/KaTeX/0.10.0/katex.min.css')</style>`)
			c.didImportKatexCSS = true
		}
		c.Printf(`<div class="equation-container">`)
		{
			c.Printf(htmlStr)
		}
		c.Printf(`</div>`)

	}
	c.Printf(`</figure>`)
}

// RenderNumberedList renders BlockNumberedList
func (c *Converter) RenderNumberedList(block *notionapi.Block) {
	isPrevSame := c.IsPrevBlockOfType(notionapi.BlockNumberedList)
	if isPrevSame {
		c.ListNo++
	} else {
		c.ListNo = 1
	}

	cls := GetBlockColorClass(block) + " numbered-list"
	cls = CleanAttributeValue(cls)

	// Notion puts <ol> around every <li>
	if c.NotionCompat || !isPrevSame {
		c.Printf(`<ol id="%s" class="%s" start="%d">`, block.ID, cls, c.ListNo)
	}
	{
		c.Printf(`<li>`)
		{
			c.RenderInlines(block.InlineContent)
			c.RenderChildren(block)
		}
		c.Printf(`</li>`)
	}
	isNextSame := c.IsNextBlockOfType(notionapi.BlockNumberedList)
	if c.NotionCompat || !isNextSame {
		c.Printf(`</ol>`)
	}
}

// RenderBulletedList renders BlockBulletedList
func (c *Converter) RenderBulletedList(block *notionapi.Block) {
	isPrevSame := c.IsPrevBlockOfType(notionapi.BlockBulletedList)
	cls := GetBlockColorClass(block) + " bulleted-list"
	cls = CleanAttributeValue(cls)
	// Notion puts <ul> around every <li>
	if c.NotionCompat || !isPrevSame {
		c.Printf(`<ul id="%s" class="%s">`, block.ID, cls)
	}
	{
		c.Printf(`<li>`)
		{
			c.RenderInlines(block.InlineContent)
			c.RenderChildren(block)
		}
		c.Printf(`</li>`)
	}
	isNextSame := c.IsNextBlockOfType(notionapi.BlockBulletedList)
	if c.NotionCompat || !isNextSame {
		c.Printf(`</ul>`)
	}
}

// RenderHeaderLevel renders BlockHeader, SubHeader and SubSubHeader
func (c *Converter) RenderHeaderLevel(block *notionapi.Block, level int) {
	cls := GetBlockColorClass(block)
	c.Printf(`<h%d id="%s" class="%s">`, level, block.ID, cls)
	c.RenderInlines(block.InlineContent)
	if c.AddHeaderAnchor {
		c.Printf(`<a class="header-anchor" href="#%s" aria-hidden="true"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 8 8"><path d="M5.88.03c-.18.01-.36.03-.53.09-.27.1-.53.25-.75.47a.5.5 0 1 0 .69.69c.11-.11.24-.17.38-.22.35-.12.78-.07 1.06.22.39.39.39 1.04 0 1.44l-1.5 1.5c-.44.44-.8.48-1.06.47-.26-.01-.41-.13-.41-.13a.5.5 0 1 0-.5.88s.34.22.84.25c.5.03 1.2-.16 1.81-.78l1.5-1.5c.78-.78.78-2.04 0-2.81-.28-.28-.61-.45-.97-.53-.18-.04-.38-.04-.56-.03zm-2 2.31c-.5-.02-1.19.15-1.78.75l-1.5 1.5c-.78.78-.78 2.04 0 2.81.56.56 1.36.72 2.06.47.27-.1.53-.25.75-.47a.5.5 0 1 0-.69-.69c-.11.11-.24.17-.38.22-.35.12-.78.07-1.06-.22-.39-.39-.39-1.04 0-1.44l1.5-1.5c.4-.4.75-.45 1.03-.44.28.01.47.09.47.09a.5.5 0 1 0 .44-.88s-.34-.2-.84-.22z"></path></svg></a>`, block.ID)
	}
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
	cls := GetBlockColorClass(block) + " toggle"
	cls = CleanAttributeValue(cls)
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

// CleanAttributeValue cleans value of an attribute
func CleanAttributeValue(v string) string {
	v = strings.TrimSpace(v)
	for {
		s := strings.Replace(v, "  ", " ", -1)
		if s == v {
			return v
		}
		v = s
	}
}

// RenderCallout renders BlockCallout
func (c *Converter) RenderCallout(block *notionapi.Block) {
	cls := GetBlockColorClass(block) + " callout"
	cls = CleanAttributeValue(cls)
	c.Printf(`<figure class="%s" style="white-space:pre-wrap;display:flex" id="%s">`, cls, block.ID)
	{
		c.Printf(`<div style="font-size:1.5em">`)
		{
			pageIcon, _ := block.PropAsString("format.page_icon")
			c.Printf(`<span class="icon">%s</span>`, pageIcon)
		}
		c.Printf(`</div>`)

		{
			c.Printf("%s", `<div style="width:100%">`)
			c.RenderInlines(block.InlineContent)
			c.Printf(`</div>`)
		}
	}
	c.Printf(`</figure>`)
}

func isHeaderBlock(block *notionapi.Block) bool {
	switch block.Type {
	case notionapi.BlockHeader, notionapi.BlockSubHeader, notionapi.BlockSubSubHeader:
		return true
	}
	return false
}

func getHeaderBlocks(blocks []*notionapi.Block, seen map[string]bool) []*notionapi.Block {
	var res []*notionapi.Block
	for _, b := range blocks {
		id := b.ID
		if seen[id] {
			// crash is better than infinite recursion
			panic("seen the same block twice")
		}
		seen[id] = true
		if isHeaderBlock(b) {
			res = append(res, b)
			continue
		}
		if b.Type == notionapi.BlockPage || b.Type == notionapi.BlockCollectionViewPage {
			continue
		}
		if len(b.Content) == 0 {
			continue
		}
		sub := getHeaderBlocks(b.Content, seen)
		res = append(res, sub...)
	}
	return res
}

func cmpBlockTypes(prev, curr string) int {
	if prev == curr {
		return 0
	}
	if prev == notionapi.BlockHeader {
		return 1
	}
	if prev == notionapi.BlockSubHeader {
		if curr == notionapi.BlockHeader {
			return -1
		}
		return 1
	}
	if prev == notionapi.BlockSubSubHeader {
		return -1
	}
	// shouldn't happen
	return 0
}

func adjustIndent(blocks []*notionapi.Block, i int) int {
	if i == 0 {
		return 0
	}
	prevType := blocks[i-1].Type
	currType := blocks[i].Type
	return cmpBlockTypes(prevType, currType)
}

// RenderTableOfContents renders BlockTableOfContents
func (c *Converter) RenderTableOfContents(block *notionapi.Block) {
	cls := GetBlockColorClass(block) + " table_of_contents"
	cls = CleanAttributeValue(cls)
	c.Printf(`<nav id="%s" class="%s">`, block.ID, cls)
	root := c.Page.Root()
	seen := map[string]bool{}
	blocks := getHeaderBlocks(root.Content, seen)
	indent := 0
	for i, b := range blocks {
		indent += adjustIndent(blocks, i)
		s := c.GetInlineContent(b.InlineContent)
		c.Printf(`<div class="table_of_contents-item table_of_contents-indent-%d">`, indent)
		{
			c.Printf(`<a class="table_of_contents-link" href="#%s">%s</a>`, b.ID, s)
		}
		c.Printf(`</div>`)
	}
	c.Printf(`</nav>`)
}

// RenderDivider renders BlockDivider
func (c *Converter) RenderDivider(block *notionapi.Block) {
	c.Printf(`<hr id="%s"/>`, block.ID)
}

// RenderCaption renders a caption
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
		cls := GetBlockColorClass(block) + " bookmark source"
		cls = CleanAttributeValue(cls)
		c.Printf(`<div class="%s">`, cls)
		{
			uri := block.Link
			text := block.Title
			c.A(uri, text, "")
			c.Printf(`<br/>`)
			c.A(uri, uri, "bookmark-href")
		}
		c.Printf(`</div>`)
		c.RenderCaption(block)
	}
	c.Printf(`</figure>`)
}

// RenderAudio renders BlockAudio
func (c *Converter) RenderAudio(block *notionapi.Block) {
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

func (c *Converter) renderEmbed(block *notionapi.Block) {
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

// RenderTweet renders BlockTweet
func (c *Converter) RenderTweet(block *notionapi.Block) {
	c.renderEmbed(block)
}

// RenderGist renders BlockGist
func (c *Converter) RenderGist(block *notionapi.Block) {
	if c.NotionCompat {
		c.renderEmbed(block)
	} else {
		uri := block.Source + ".js"
		// TODO: support caption
		// TODO: maybe support comments
		// TODO: quote uri
		c.Printf(`<script src="%s", class="notion-embed-gist"></script>`, uri)
	}
}

// RenderCodepen renders BlockCodepen
func (c *Converter) RenderCodepen(block *notionapi.Block) {
	c.renderEmbed(block)
}

// RenderMaps renders BlockMaps
func (c *Converter) RenderMaps(block *notionapi.Block) {
	c.renderEmbed(block)
}

// RenderFigma renders BlockFigma
func (c *Converter) RenderFigma(block *notionapi.Block) {
	c.Printf(`<figure id="%s">`, block.ID)
	{
		c.Printf(`<div class="source">`)
		{
			uri := block.Source
			c.Printf(`<a href="%s">%s</a>`, uri, uri)
		}

		c.Printf(`</div>`)
		c.RenderCaption(block)
	}
	c.Printf(`</figure>`)
}

// RenderFile renders BlockFile
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

// RenderDrive renders BlockDrive
func (c *Converter) RenderDrive(block *notionapi.Block) {
	c.Printf(`<figure id="%s">`, block.ID)
	{
		c.Printf(`<div class="bookmark source">`)
		{
			icon, _ := block.PropAsString("format.drive_properties.icon")
			c.Printf(`<img style="width:1em;height:1em;margin-right:0.5em;vertical-align:text-bottom" src="%s"/>`, icon)

			docURL, _ := block.PropAsString("format.drive_properties.url")
			title, _ := block.PropAsString("format.drive_properties.title")
			c.Printf(`<a href="%s">%s</a>`, docURL, title)
			c.Printf(`<br/>`)
			c.Printf(`<a class="bookmark-href" href="%s">%s</a>`, docURL, docURL)
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
		c.Printf(`<img %ssrc="%s"/>`, style, uri)
		c.Printf(`</a>`)

		c.RenderCaption(block)
	}
	c.Printf(`</figure>`)
}

// RenderColumnList renders BlockColumnList
// Its children are BlockColumn
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
// Its parent is BlockColumnList
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

func (c *Converter) findParentPageID(page *notionapi.Page, id string) string {
	// we traverse blocks upwards until we find a block of type Page
	currID := id
	for {
		block := page.BlockByID(currID)
		// assume it's a Page block that's not in Page's block
		if block == nil {
			return currID
		}
		if block.Type == notionapi.BlockPage {
			return currID
		}
		currID = block.ParentID
	}
}

// RenderBreadcrumb renders BlockBreadcrumb
func (c *Converter) RenderBreadcrumb(block *notionapi.Block) {
	if c.NotionCompat {
		// Notion doesn't render breadcrumbs
		return
	}
	c.Printf(`<div class="breadcrumbs">`)
	pages := []*notionapi.Page{}
	curr := c.Page
	for {
		id := curr.Root().ParentID
		id = c.findParentPageID(curr, id)
		parent := c.PageByID(id)
		if parent == nil {
			break
		}
		pages = append(pages, parent)
		curr = parent
	}
	// TODO: add icon
	// we traverse from the end because they were put
	// in reverse order
	idx := len(pages) - 1
	for i := idx; i >= 0; i-- {
		page := pages[i]
		title := page.Root().Title
		pageID := notionapi.ToNoDashID(page.Root().ID)
		uri := "https://www.notion.so/" + pageID
		uri = c.RewriteURL(uri)
		uri = EscapeHTML(uri)
		c.Printf(`<div><a href="%s">%s</a></div>`, uri, title)
		c.Printf("<div>/</div>")
	}
	title := c.Page.Root().Title
	c.Printf(`<div>%s</div>`, title)
	c.Printf(`</div>`)
}

/*
func hasTitleColumn(columns []*notionapi.ColumnInfo) bool {
	for _, ci := range columns {
		if ci.Type() == notionapi.ColumnTypeTitle {
			return true
		}
	}
	return false
}
*/

func (c *Converter) renderTableHeader(tv *notionapi.TableView, col int) {
	var style, name string
	ci := tv.Columns[col]
	if ci != nil {
		name = ci.Name()
		name = EscapeHTML(name)

		style = fmt.Sprintf(` width="%d"`, ci.Property.Width)
	}
	c.Printf(`<th%s>%s</th>`, style, name)
}

func isEmptyBlock(block *notionapi.Block) bool {
	if block == nil {
		return true
	}
	return len(block.ContentIDs) == 0
}

func (c *Converter) renderTableCell(tv *notionapi.TableView, row, col int) {
	ci := tv.Columns[col]
	tr := tv.Rows[row]
	rowPage := tr.Page
	colName := ci.ID()
	schema := ci.Schema
	textSpans := tv.CellContent(row, col)
	colVal := c.GetInlineContent(textSpans)

	// TODO: see e.g. b20a830016df405ea641936f8a5bd572
	// schema can be nil for relation properties. Notion puts them as last row,
	// the value comes from page and their schema has to be fished out

	if schema == nil {
		colNameCls := EscapeHTML(colName)
		if colVal == "" {
			colVal = "&nbsp;"
		}
		c.Printf(`<td class="cell-%s">%s</td>`, colNameCls, colVal)
		return
	}

	typ := schema.Type

	if typ == notionapi.ColumnTypeTitle {
		if isEmptyBlock(rowPage) {
			// row here is a page. For cosmetic reasons we don't want
			// to link to empty pages.
		} else {
			uri := c.tableTitleCellURL(tv, row, col)
			if colVal == "" {
				colVal = "Untitled"
			}
			colVal = fmt.Sprintf(`<a href="%s">%s</a>`, uri, colVal)
		}
	} else if typ == notionapi.ColumnTypeMultiSelect {
		vals := strings.Split(colVal, ",")
		s := ""
		for idx := range vals {
			val := vals[idx]
			if val == "" {
				continue
			}
			v := EscapeHTML(val)
			col := getMultiSelectoColor(schema.Options, val)
			if col == "" {
				s += fmt.Sprintf(`<span class="selected-value">%s</span>`, v)
			} else {
				s += fmt.Sprintf(`<span class="selected-value block-color-%s_background">%s</span>`, col, v)
			}
		}
		colVal = s
	} else if typ == notionapi.ColumnTypeCreatedTime {
		// TODO: better formatting. Notion seems to be using
		// relative formatting like "Today 3:03pm"
		colVal = rowPage.CreatedOn().Format("2006-01-02")
	} else if typ == notionapi.ColumnTypeLastEditedTime {
		// TODO: better formatting. Notion seems to be using
		// relative formatting like "Today 3:03pm"
		colVal = rowPage.LastEditedOn().Format("2006-01-02")
	} else if typ == notionapi.ColumnTypeNumber {
		// TODO: format number
		colVal = fmtNumber(colVal, schema.NumberFormat)
	} else if typ == notionapi.ColumnTypeLastEditedBy {
		uid := rowPage.LastEditedBy
		colVal = notionapi.GetUserNameByID(tv.Page, uid)
	} else if typ == notionapi.ColumnTypeCreatedBy {
		uid := rowPage.CreatedBy
		colVal = notionapi.GetUserNameByID(tv.Page, uid)
	} else if schema.Type == notionapi.ColumnTypeRelation {
		// TODO: not sure how to format relations
		//colVal = c.GetInlineContent(textSpans)
		colVal = ""
	}

	colNameCls := EscapeHTML(colName)
	if colVal == "" {
		colVal = "&nbsp;"
	}
	c.Printf(`<td class="cell-%s">%s</td>`, colNameCls, colVal)
}

func fmtNumber(v string, numFmt string) string {
	if numFmt == "dollar" {
		v = strings.TrimPrefix(v, "$")
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return v
		}
		return fmt.Sprintf("$%.02f", f)
	}
	// TODO: mmore formats
	return v
}
func getMultiSelectoColor(opts []*notionapi.CollectionColumnOption, val string) string {
	for _, opt := range opts {
		if opt.Value == val {
			return opt.Color
		}
	}
	return ""
}

func (c *Converter) renderTableRow(tv *notionapi.TableView, row int) {
	tr := tv.Rows[row]
	c.Printf(`<tr id="%s">`, tr.Page.ID)
	nCols := tv.ColumnCount()
	for col := 0; col < nCols; col++ {
		c.renderTableCell(tv, row, col)
	}
	c.Printf("</tr>\n")
}

// RenderCollectionView renders BlockCollectionView
func (c *Converter) RenderCollectionView(block *notionapi.Block) {
	pageID := ""
	if c.Page != nil {
		pageID = notionapi.ToNoDashID(c.Page.ID)
	}

	if len(block.TableViews) == 0 {
		logf("missing block.CollectionViews for block %s %s in page %s\n", block.ID, block.Type, pageID)
		return
	}
	// render only the first one
	tv := block.TableViews[0]

	nCols := tv.ColumnCount()
	if nCols == 0 {
		logf("didn't find columns inof in block '%s'\n", tv.CollectionView.ID)
		return
	}
	isList := tv.CollectionView.Type == notionapi.CollectionViewTypeList
	//hasTitle := hasTitleColumn(tv.Columns)

	c.Printf(`<div id="%s" class="collection-content">`, block.ID)
	{
		name := tv.Collection.GetName()
		c.Printf(`<h4 class="collection-title">%s</h4>`, name)
		if isList {
			c.Printf("%s", `<table class="collection-content" style="width: 100%">`)
		} else {
			c.Printf(`<table class="collection-content">`)
		}

		// for lists we don't show header
		if !isList {
			c.Printf(`<thead>`)
			{
				c.Printf(`<tr>`)
				for col := 0; col < nCols; col++ {
					c.renderTableHeader(tv, col)
				}
				c.Printf(`</tr>`)
			}
			c.Printf(`</thead>`)
		}

		c.Printf(`<tbody>`)
		{
			nRows := tv.RowCount()
			for row := 0; row < nRows; row++ {
				c.renderTableRow(tv, row)
			}
		}
		c.Printf(`</tbody>`)

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
	case notionapi.BlockEquation:
		return c.RenderEquation
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
	case notionapi.BlockMaps:
		return c.RenderMaps
	case notionapi.BlockCodepen:
		return c.RenderCodepen
	case notionapi.BlockTweet:
		return c.RenderTweet
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
		return c.RenderTableOfContents
	case notionapi.BlockBreadcrumb:
		return c.RenderBreadcrumb
	case notionapi.BlockFactory:
		return nil
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

func (c *Converter) detectKatex() error {
	katexPath := c.KatexPath
	if katexPath != "" {
		if _, err := os.Stat(c.KatexPath); err == nil {
			return nil
		}
	}
	katexPath, err := exec.LookPath("katex")
	if err != nil {
		if c.KatexPath != "" {
			return fmt.Errorf("UseKatexToRenderEquation is set but KatexPath ('%s') doesn't exist", c.KatexPath)
		}
		return fmt.Errorf("UseKatexToRenderEquation is set but couldn't locate katex binary (see https://katex.org/). You can install Katex with `npm install -g katex`. You can provide the path to katex binary via KatexPath. ")
	}
	c.KatexPath = katexPath
	return nil
}

// ToHTML renders a page to html
func (c *Converter) ToHTML() ([]byte, error) {
	if c.NotionCompat {
		c.UseKatexToRenderEquation = true
	}
	if c.UseKatexToRenderEquation {
		if err := c.detectKatex(); err != nil {
			return nil, err
		}
	}

	c.PushNewBuffer()
	c.RenderBlock(c.Page.Root())
	buf := c.PopBuffer()
	return buf.Bytes(), nil
}

// ToHTML converts a page to HTML
func ToHTML(page *notionapi.Page) []byte {
	r := NewConverter(page)
	// the only error that can happen is katex binary
	// not existing. Since we don't ask
	res, _ := r.ToHTML()
	return res
}
