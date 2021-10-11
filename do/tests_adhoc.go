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
	client := newClient()

	// just enough data for DownloadFile
	b := &notionapi.Block{
		ID:          "5cc81055-1b81-4f31-9df3-390152d272cf",
		ParentTable: "table",
	}
	uri := "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/60550647-d8af-4321-b268-cbb1bab09210/SumatraPDF-dll_iITXbPm55F.png"
	rsp, err := client.DownloadFile(uri, b)
	if err != nil {
		fmt.Printf("c.DownloadFile() failed with '%s'\n", err)
		return
	}
	fmt.Printf("c.DownloadFile() downloaded %d bytes\n", len(rsp.Data))
}

func testDownloadImage() {
	client := newClient()

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
		rsp, err := client.DownloadFile(block.Source, block)
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
		rsp, err := client.DownloadFile(block.Source, block)
		assert(err == nil, "client.DownloadFile(%s) failed with %s", err, block.Source)
		fmt.Printf("Downloaded image %s of size %d\n", block.Source, len(rsp.Data))
		ct := rsp.Header.Get("Content-Type")
		exp = "image/png"
		assert(ct == exp, "unexpected Content-Type, wanted %s, got %s", exp, ct)
	}
}

func testGist() {
	client := newClient()

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

	client := newClient()

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
	client := newClient()

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
	client := newClient()

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
