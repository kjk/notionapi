package main

import (
	"fmt"

	"github.com/kjk/notionapi"
)

func assert(ok bool, format string, args ...interface{}) {
	if ok {
		return
	}
	s := fmt.Sprintf(format, args...)
	panic(s)
}

func pageURL(pageID string) string {
	return "https://notion.so/" + pageID
}

func testDownloadFile() {
	client := &notionapi.Client{
		DebugLog:  flgVerbose,
		AuthToken: getToken(),
	}
	uri := "https://s3.us-west-2.amazonaws.com/secure.notion-static.com/cacd0ab4-224c-4276-814f-e80a7fea1b3a/Screen_Shot_2020-03-19_at_6.11.37_PM.png?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=ASIAT73L2G45GQ6I6PXS%2F20200320%2Fus-west-2%2Fs3%2Faws4_request&X-Amz-Date=20200320T011635Z&X-Amz-Expires=86400&X-Amz-Security-Token=IQoJb3JpZ2luX2VjEPb%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaCXVzLXdlc3QtMiJIMEYCIQCARBficIVNgZU6KsQpgVmlsWMvm64cvznOFSIQrViUEAIhAMojl3CRPCOhLsxR5Q6%2F%2BV28mLpwL4TJjgVBzrLPXUqfKr0DCN%2F%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEQABoMMjc0NTY3MTQ5MzcwIgwZyKv1lfjAdtO4I9kqkQNvn90XhdlUgA9UpNvIDDGCf%2FcFtyBSBiDtPgU92QmQjJmrCIfslN7g3UnWak2sfCbWQgScGcHafjZv4OG0IuixJGAZPaaMvlOOOy9G3q5JZz4%2BinwtMLf%2BzJBbUp9DDyhdm6rTWOdLxyTtWgfx3LdyBmhn%2FqvBXO6xLjyz4p58cry%2FSUrdTPe47y%2BUbFgbhFxp2zGYj63gHGRDASDzODaibYmxaprj2hSLHuiVfKL1Y582%2F6Gc8AIGNj3IybXG%2Ftn6db0Qh8mTzClIoyEW9VE0b%2BEX2kXE0lFn51ivcPGK3KoJlbcHVeJoQwfmPieq%2BnLkX97ur5VjYx%2F9eiROoIHpsRykLAuC85ZFmthZTDoR6pnSMiVKtWBGQWCOhUm27Jbr4AcPlC%2BYN23GvZgLRIzs9Xw4afsg8vgxkncAfraVqK9h3SSzW2XlXN7DMNaRaowko552oX%2BgcJEaFRO9zK5vsrCp6q1tkgiSqnafeb9MVkUttXNGd94VGWZgyh3FQL%2B6BqMUfIBppFnTGrXHhMXJaTDkys%2FzBTrqAR17y7hpLP5bM9hVoJQolj3N5PmytUXV5dQmtrriIYUWzKRPi68V2b9HMr3H4R%2BkiEdPELfMcWLUVA3mOZkXI2bfm2pTkan%2BfaDSzMmgWMqUcu0C12DG3iE5AEoOKyHZXjQuvgVjKk3RCntFzwntNn0nljM43FxGFiZ1iy0NnK%2Bf21N1tLy43BpO96ABc%2FYtVQ1DNI2o9lPxo8dfO8yWaxC97f%2BfYL4LHtSzwufnN6CJZiPQ9OmGNIxkOBWO6z5eKRudWs2bxvHkapQZ08lFRAKbuosslWu549iM5hrGqaPdD0hTbDEV4rK7CQ%3D%3D&X-Amz-Signature=6c6b08db4167214f255a31df07e27de3387d10a5a1e8b09c8f27cb789a880d67&X-Amz-SignedHeaders=host&response-content-disposition=filename%20%3D%22Screen_Shot_2020-03-19_at_6.11.37_PM.png%22"
	rsp, err := client.DownloadFile(uri, "8511412cbfde432ba226648e9bdfbec2")
	if err != nil {
		fmt.Printf("c.DownloadFile() failed with '%s'\n", err)
		return
	}
	fmt.Printf("c.DownloadFile() downloaded %d bytes\n", len(rsp.Data))
}

func testDownloadImage() {
	client := &notionapi.Client{
		DebugLog:  flgVerbose,
		AuthToken: getToken(),
	}
	// page with images
	pageID := "8511412cbfde432ba226648e9bdfbec2"
	fmt.Printf("testDownloadImage %s\n", pageURL(pageID))
	page, err := downloadPage(client, pageID)
	must(err)
	block := page.Root()
	assert(block.Title == "Test image", "unexpected title ''%s'", block.Title)
	blocks := block.Content
	assert(len(blocks) == 2, "expected 2 blockSS, got %d", len(blocks))

	block = blocks[0]
	if false {
		fmt.Printf("block.Source: %s\n", block.Source)
		exp := "https://i.imgur.com/NT9NcB6.png"
		assert(block.Source == exp, "expected %s, got %s", exp, block.Source)
		rsp, err := client.DownloadFile(block.Source, block.ID)
		assert(err == nil, "client.DownloadFile(%s) failed with %s", err, block.Source)
		fmt.Printf("Downloaded image %s of size %d\n", block.Source, len(rsp.Data))
		ct := rsp.Header.Get("Content-Type")
		exp = "image/png"
		assert(ct == exp, "unexpected Content-Type, wanted %s, got %s", exp, ct)
		disp := rsp.Header.Get("Content-Disposition")
		exp = "filename=\"NT9NcB6.png\""
		assert(disp == exp, "unexpected Content-Disposition, got %s, wanted %s", disp, exp)
	}

	block = blocks[1]
	if true {
		fmt.Printf("block.Source: %s\n", block.Source)
		exp := "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/e5661303-82e1-43e4-be8e-662d1598cd53/untitled"
		assert(block.Source == exp, "expected '%s', got '%s'", exp, block.Source)
		rsp, err := client.DownloadFile(block.Source, block.ID)
		assert(err == nil, "client.DownloadFile(%s) failed with %s", err, block.Source)
		fmt.Printf("Downloaded image %s of size %d\n", block.Source, len(rsp.Data))
		ct := rsp.Header.Get("Content-Type")
		exp = "image/png"
		assert(ct == exp, "unexpected Content-Type, wanted %s, got %s", exp, ct)
	}
}

func testGist() {
	client := &notionapi.Client{
		DebugLog:  flgVerbose,
		AuthToken: getToken(),
	}
	// gist page
	pageID := "7b9cdf3ab2cf405692e9810b0ac8322e"
	fmt.Printf("testGist %s\n", pageURL(pageID))
	page, err := downloadPage(client, pageID)
	must(err)
	title := page.Root().Title
	assert(title == "Test Gist", "unexpected title ''%s'", title)
	blocks := page.Root().Content
	assert(len(blocks) == 1, "expected 1 block, got %d", len(blocks))
	block := blocks[0]
	src := block.Source
	assert(src == "https://gist.github.com/kjk/7278df5c7b164fce3c949af197c961eb", "unexpected Source '%s'", src)
}

func testChangeFormat() {
	authToken := getToken()
	if authToken == "" {
		fmt.Printf("Skipping testChangeFormat() because NOTION_TOKEN env variable not provided")
		return
	}

	client := &notionapi.Client{
		DebugLog:  flgVerbose,
		AuthToken: authToken,
	}
	// https://www.notion.so/Test-for-change-title-7e825831be07487e87e756e52914233b
	pageID := "7e825831be07487e87e756e52914233b"
	pageID = "0fc3a590ba5f4e128e7c750e8ecc961d"
	page, err := client.DownloadPage(pageID)
	if err != nil {
		fmt.Printf("testChangeFormat: client.DownloadPage() failed with '%s'\n", err)
		return
	}
	origFormat := page.Root().FormatPage()
	if origFormat == nil {
		origFormat = &notionapi.FormatPage{}
	}
	newSmallText := !origFormat.PageSmallText
	newFullWidth := !origFormat.PageFullWidth

	args := map[string]interface{}{
		"page_small_text": newSmallText,
		"page_full_width": newFullWidth,
	}
	fmt.Printf("Setting format to: page_small_text: %v, page_full_width: %v\n", newSmallText, newFullWidth)
	err = page.SetFormat(args)
	if err != nil {
		fmt.Printf("testChangeFormat: page.SetFormat() failed with '%s'\n", err)
		return
	}
	page2, err := client.DownloadPage(pageID)
	if err != nil {
		fmt.Printf("testChangeFormat: client.DownloadPage() failed with '%s'\n", err)
		return
	}
	format := page2.Root().FormatPage()
	assert(newSmallText == format.PageSmallText, "'%v' != '%v' (newSmallText != format.PageSmallText)", newSmallText, format.PageSmallText)
	assert(newFullWidth == format.PageFullWidth, "'%v' != '%v' (newFullWidth != format.PageFullWidth)", newFullWidth, format.PageFullWidth)
}

func testChangeTitle() {
	authToken := getToken()
	if authToken == "" {
		fmt.Printf("Skipping testChangeTitle() because NOTION_TOKEN env variable not provided")
		return
	}
	client := &notionapi.Client{
		DebugLog:  flgVerbose,
		AuthToken: authToken,
	}
	// https://www.notion.so/Test-for-change-title-7e825831be07487e87e756e52914233b
	pageID := "7e825831be07487e87e756e52914233b"
	page, err := client.DownloadPage(pageID)
	if err != nil {
		fmt.Printf("testChangeTitle: client.DownloadPage() failed with '%s'\n", err)
		return
	}
	origTitle := page.Root().Title
	newTitle := origTitle + " changed"
	fmt.Printf("Changing title from '%s' to '%s'\n", origTitle, newTitle)
	err = page.SetTitle(newTitle)
	if err != nil {
		fmt.Printf("testChangeTitle: page.SetTitle(newTitle) failed with '%s'\n", err)
	}

	page2, err := client.DownloadPage(pageID)
	if err != nil {
		fmt.Printf("testChangeTitle: client.DownloadPage() failed with '%s'\n", err)
		return
	}
	title := page2.Root().Title
	assert(title == newTitle, "'%s' != '%s' (title != newTitle)", title, newTitle)

	fmt.Printf("Changing title from '%s' to '%s'\n", title, origTitle)
	err = page2.SetTitle(origTitle)
	if err != nil {
		fmt.Printf("testChangeTitle: page2.SetTitle(origTitle) failed with '%s'\n", err)
	}
}

func testDownloadBig() {
	// this tests downloading a page that has (hopefully) all kinds of elements
	// for notion, for testing that we handle everything
	// page is c969c9455d7c4dd79c7f860f3ace6429 https://www.notion.so/Test-page-all-not-c969c9455d7c4dd79c7f860f3ace6429
	client := &notionapi.Client{
		DebugLog:  flgVerbose,
		AuthToken: getToken(),
	}
	// page with images
	pageID := "c969c9455d7c4dd79c7f860f3ace6429"
	fmt.Printf("testDownloadImage %s\n", pageURL(pageID))
	page, err := downloadPage(client, pageID)
	must(err)
	s := notionapi.DumpToString(page)
	fmt.Printf("Downloaded page %s, %s\n%s\n", page.ID, pageURL(pageID), s)
}

func adhocTests() {
	fmt.Printf("Running page tests\n")
	recreateDir(cacheDir)

	//testDownloadBig()
	testDownloadImage()
	//testGist()
	//testChangeTitle()
	//testChangeFormat()

	fmt.Printf("Finished tests ok!\n")
}
