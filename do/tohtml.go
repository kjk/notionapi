package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kjk/notionapi"
	"github.com/kjk/notionapi/tohtml"
)

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

func emptyLogDir() {
	// TODO: maybe only if flgUseCache is not used?
	// os.RemoveAll(logDir)
	os.MkdirAll(logDir, 0755)
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

func makeNotionClient() *notionapi.Client {
	client := &notionapi.Client{
		DebugLog: true,
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

func loadCachedPage(pageID string) *notionapi.Page {
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

func downloadPage(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	client.Logger, _ = openLogFileForPageID(pageID)
	if client.Logger != nil {
		defer func() {
			f := client.Logger.(*os.File)
			f.Close()
		}()
	}
	return client.DownloadPage(pageID)
}

func downloadPageMaybeCached(pageID string) *notionapi.Page {
	if flgUseCache {
		log("trying cache\n")
		page := loadCachedPage(pageID)
		if page != nil {
			log("Loaded page %s from cache %s\n", pageID, jsonFilePathForPageID(pageID))
			return page
		}
	}

	client := makeNotionClient()
	pageURL := notionURLForPageID(pageID)

	timeStart := time.Now()
	page, err := downloadPage(client, pageID)
	if err != nil {
		log("Download of page '%s' failed with %s\n", pageURL, err)
		must(err)
		return nil
	}

	log("Downloaded page %s in %s\n", pageURL, time.Since(timeStart))
	jsonPath := savePageAsJSON(page)
	simplePath := savePageAsSimpleStructure(page)
	log("%s : log of HTTP traffic\n", logFilePathForPageID(pageID))
	log("%s : notionapi.Page serialized as JSON\n", jsonPath)
	log("%s : notionapi.Page serialized in simple format\n", simplePath)
	return page
}

func toHTML(pageID string) {
	page := downloadPageMaybeCached(pageID)
	if page == nil {
		return
	}

	r := tohtml.NewHTMLRenderer(page)
	r.AddIDAttribute = true
	r.AddHeaderAnchor = true
	notionapi.PanicOnFailures = true
	html := r.ToHTML()

	html = makeFullHTML(html)
	path := htmlFilePathForPageID(pageID)
	err := ioutil.WriteFile(path, html, 0644)
	must(err)
	log("%s : HTML version of the page\n", path)
	if !flgNoOpen {
		path, err := filepath.Abs(path)
		must(err)
		uri := "file://" + path
		log("Opening browser with %s\n", uri)
		openBrowser(uri)
	}
}
