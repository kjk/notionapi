package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjk/notionapi"
)

func makeNotionClient() *notionapi.Client {
	client := &notionapi.Client{
		DebugLog:  true,
		AuthToken: getToken(),
	}
	notionToken := strings.TrimSpace(os.Getenv("NOTION_TOKNE"))
	if notionToken == "" {
		log("NOTION_TOKEN env variable not set. Can only access public pages\n")
	} else {
		log("NOTION_TOKEN env variable set, can access private pages\n")
		// TODO: validate that the token looks legit
		client.AuthToken = notionToken
	}
	return client
}

func notionURLForPageID(pageID string) string {
	return "https://notion.so/" + pageID
}

func logFilePathForPageID(pageID string) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.log.txt", pageID)
	return filepath.Join(logDir, name)
}

func jsonFilePathForPageID(pageID string) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.page.json", pageID)
	return filepath.Join(logDir, name)
}

func htmlFilePathForPageID(pageID string) string {
	pageID = notionapi.ToNoDashID(pageID)
	name := fmt.Sprintf("%s.page.html", pageID)
	return filepath.Join(logDir, name)
}

func simpleStructureFilePathForPageID(pageID string) string {
	name := fmt.Sprintf("%s.page.structure.txt", pageID)
	return filepath.Join(logDir, name)
}

func openLogFileForPageID(pageID string) (io.WriteCloser, error) {
	path := logFilePathForPageID(pageID)
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		return nil, err
	}
	return f, nil
}

func loadCachedPage(pageID string) *notionapi.Page {
	if flgNoCache {
		return nil
	}

	jsonPath := jsonFilePathForPageID(pageID)
	d, err := ioutil.ReadFile(jsonPath)
	if err != nil {
		// it's ok if file doesn't exit
		return nil
	}
	var page *notionapi.Page
	err = json.Unmarshal(d, &page)
	if err == nil {
		return page
	}
	if err != nil {
		log("json.Unmarshal() failed with %s decoding file %s\n", err, jsonPath)
		err = os.Remove(jsonPath)
		must(err)
		log("Deleted file %s\n", jsonPath)
	}
	return nil
}

// returns path of the created file
func savePageAsJSON(page *notionapi.Page) string {
	d, err := json.MarshalIndent(page, "", "  ")
	must(err)
	path := jsonFilePathForPageID(page.ID)
	err = ioutil.WriteFile(path, d, 0644)
	must(err)
	return path
}

// returns path of the created file
func savePageAsSimpleStructure(page *notionapi.Page) string {
	path := simpleStructureFilePathForPageID(page.ID)
	f, err := os.Create(path)
	must(err)
	defer f.Close()
	notionapi.Dump(f, page)
	return path
}

func downloadPageCached(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	page := loadCachedPage(pageID)
	if page != nil {
		return page, nil
	}

	client.Logger, _ = openLogFileForPageID(pageID)
	if client.Logger != nil {
		defer func() {
			f := client.Logger.(*os.File)
			f.Close()
			client.Logger = nil
		}()
	}
	res, err := client.DownloadPage(pageID)
	if err != nil {
		return nil, err
	}
	savePageAsJSON(res)
	return res, nil
}

func downloadPage(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	page, err := downloadPageCached(client, pageID)
	if err != nil {
		fmt.Printf("downloadPageCached('%s') failed with %s\n", pageID, err)
		return nil, err
	}
	return page, nil
}
