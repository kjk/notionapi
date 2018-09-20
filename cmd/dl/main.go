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
	useCache = true
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
	notionapi.Logger = f
	return f, nil
}

func downloadPageCached(pageID string) (*notionapi.Page, error) {
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
	res, err := notionapi.DownloadPage(pageID)
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

func dl(pageID string) (*notionapi.Page, error) {
	lf, _ := openLogFileForPageID(pageID)
	if lf != nil {
		defer lf.Close()
	}
	page, err := downloadPageCached(pageID)
	if err != nil {
		fmt.Printf("downloadPageCached('%s') failed with %s\n", pageID, err)
		return nil, err
	}
	return page, nil
}

func main() {
	if len(os.Args) != 2 {
		usageAndExit()
	}

	os.RemoveAll(logDir)
	os.MkdirAll(logDir, 0755)
	os.RemoveAll(cacheDir)
	os.MkdirAll(cacheDir, 0755)

	pageID := os.Args[1]
	notionapi.DebugLog = true
	_, err := dl(pageID)
	if err != nil {
		fmt.Printf("dl failed with %s\n", err)
	} else {
		fmt.Printf("Downloaded page %s\n", pageID)
	}
}
