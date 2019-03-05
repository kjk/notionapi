package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kjk/notionapi"
)

const (
	logDir   = "log"
	cacheDir = "cache"
)

var (
	useCache = false
)

func usageAndExit() {
	fmt.Printf("Usage: %s <notion_page_id>\n", filepath.Base(os.Args[0]))
	os.Exit(1)
}

func openLogFileForPageID(pageID string) (io.WriteCloser, error) {
	name := fmt.Sprintf("%s.go.log.txt", pageID)
	path := filepath.Join(logDir, name)
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		return nil, err
	}
	return f, nil
}

func downloadPageCached(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	var page notionapi.Page
	cachedPath := filepath.Join(cacheDir, pageID+".json")
	if useCache {
		d, err := ioutil.ReadFile(cachedPath)
		if err == nil {
			err = json.Unmarshal(d, &page)
			if err == nil {
				fmt.Printf("Got data for pageID %s from cache file %s\n", pageID, cachedPath)
				return &page, nil
			}
			// not a fatal error, just a warning
			fmt.Printf("json.Unmarshal() on '%s' failed with %s\n", cachedPath, err)
		}
	}
	res, err := client.DownloadPage(pageID)
	if err != nil {
		return nil, err
	}
	d, err := json.MarshalIndent(res, "", "  ")
	if err == nil {
		err = ioutil.WriteFile(cachedPath, d, 0644)
		if err != nil {
			// not a fatal error, just a warning
			fmt.Printf("ioutil.WriteFile(%s) failed with %s\n", cachedPath, err)
		}
	} else {
		// not a fatal error, just a warning
		fmt.Printf("json.Marshal() on pageID '%s' failed with %s\n", pageID, err)
	}
	return res, nil
}

func dl(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	client.Logger, _ = openLogFileForPageID(pageID)
	if client.Logger != nil {
		defer func() {
			f := client.Logger.(*os.File)
			f.Close()
		}()
	}
	page, err := downloadPageCached(client, pageID)
	if err != nil {
		fmt.Printf("downloadPageCached('%s') failed with %s\n", pageID, err)
		return nil, err
	}
	return page, nil
}

func panicIfErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func reacreateDir(dir string) {
	os.RemoveAll(dir)
	err := os.MkdirAll(dir, 0755)
	panicIfErr(err)
}

func assert(ok bool, format string, args ...interface{}) {
	if ok {
		return
	}
	s := fmt.Sprintf(format, args...)
	panic(s)
}

func testGist() {
	client := &notionapi.Client{
		DebugLog: true,
	}
	// gist page
	pageID := "7b9cdf3ab2cf405692e9810b0ac8322e"
	page, err := dl(client, pageID)
	panicIfErr(err)
	title := page.Root.Title
	assert(title == "Test Gist", "unexpected title ''%s'", title)
	blocks := page.Root.Content
	assert(len(blocks) == 1, "expected 1 block, got %d", len(blocks))
	block := blocks[0]
	src := block.Source
	assert(src == "https://gist.github.com/kjk/7278df5c7b164fce3c949af197c961eb", "unexpected Source '%s'", src)
}

func testChangeFormat() {
	authToken := os.Getenv("NOTION_TOKEN")
	if authToken == "" {
		fmt.Printf("Skipping testChangeFormat() because NOTION_TOKEN env variable not provided")
		return
	}
	client := &notionapi.Client{
		DebugLog:  true,
		AuthToken: authToken,
	}
	// https://www.notion.so/Test-for-change-title-7e825831be07487e87e756e52914233b
	pageID := "7e825831be07487e87e756e52914233b"
	page, err := client.DownloadPage(pageID)
	if err != nil {
		fmt.Printf("testChangeFormat: client.DownloadPage() failed with '%s'\n", err)
		return
	}
	origFormat := page.Root.FormatPage
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
	format := page2.Root.FormatPage
	assert(newSmallText == format.PageSmallText, "'%v' != '%v' (newSmallText != format.PageSmallText)", newSmallText, format.PageSmallText)
	assert(newFullWidth == format.PageFullWidth, "'%v' != '%v' (newFullWidth != format.PageFullWidth)", newFullWidth, format.PageFullWidth)
}

func testChangeTitle() {
	authToken := os.Getenv("NOTION_TOKEN")
	if authToken == "" {
		fmt.Printf("Skipping testChangeTitle() because NOTION_TOKEN env variable not provided")
		return
	}
	client := &notionapi.Client{
		DebugLog:  true,
		AuthToken: authToken,
	}
	// https://www.notion.so/Test-for-change-title-7e825831be07487e87e756e52914233b
	pageID := "7e825831be07487e87e756e52914233b"
	page, err := client.DownloadPage(pageID)
	if err != nil {
		fmt.Printf("testChangeTitle: client.DownloadPage() failed with '%s'\n", err)
		return
	}
	origTitle := page.Root.Title
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
	title := page2.Root.Title
	assert(title == newTitle, "'%s' != '%s' (title != newTitle)", title, newTitle)

	fmt.Printf("Changing title from '%s' to '%s'\n", title, origTitle)
	err = page2.SetTitle(origTitle)
	if err != nil {
		fmt.Printf("testChangeTitle: page2.SetTitle(origTitle) failed with '%s'\n", err)
	}
}

func main() {
	fmt.Printf("Running page tests\n")
	reacreateDir(logDir)
	reacreateDir(cacheDir)

	testGist()
	testChangeTitle()
	testChangeFormat()

	fmt.Printf("Finished tests ok!\n")
}
