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

func main() {
	fmt.Printf("Running page tests\n")
	reacreateDir(logDir)
	reacreateDir(cacheDir)

	client := &notionapi.Client{
		DebugLog: true,
	}
	{
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
	fmt.Printf("Finished tests ok!\n")
}
