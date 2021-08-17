package notionapi

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kjk/siser"
)

const (
	recCacheName = "noahttpcache"
)

type CachingPolicy int

const (
	// PolicyCacheOnly - will only read from cache, no calling Notion server
	PolicyCacheOnly CachingPolicy = iota
	// PolicyDownloadNewer - will only download from Notion server if there is a newer version of the page
	PolicyDownloadNewer
	// PolicyDownloadAlways - will always download from Notion server (and update the cache with updated version)
	PolicyDownloadAlways
)

// RequestCacheEntry has info about request (method/url/body) and response
type RequestCacheEntry struct {
	// request info
	Method string
	URL    string
	Body   string

	bodyPP string // cached pretty printed version
	// response
	Response []byte
}

type CachedPage struct {
	PageFromCache  *Page
	PageFromServer *Page
	LatestVer      int64
}

// CachingClient implements optimized (cached) downloading of pages.
// Cache of pages is stored in CacheDir. We return pages from cache.
// If RedownloadNewerVersions is true, we'll re-download latest version
// of the page (as opposed to returning possibly outdated version
// from cache). We do it more efficiently than just blindly re-downloading.
type CachingClient struct {
	CacheDir string

	// you can set it to over-ride location of where we store cached files
	// if not set, it'll be filepath.Join(CacheDir, "files")
	CacheDirFiles string
	Client        *Client

	Policy CachingPolicy

	// disable pretty-printing of json responses saved in the cache
	NoPrettyPrintResponse bool

	// maps no-dash id to info about a page
	IdToCachedPage map[string]*CachedPage

	DownloadedCount      int
	FromCacheCount       int
	DownloadedFilesCount int
	FilesFromCacheCount  int

	RequestsFromCache      int
	RequestsFromServer     int
	RequestsWrittenToCache int

	pageIDToEntries map[string][]*RequestCacheEntry
	// we cache requests on a per-page basis
	currPageID *NotionID

	currPageRequests      []*RequestCacheEntry
	needSerializeRequests bool
	didCheckVersions      bool

	// names of files in file cache
	fileNamesInCache []string
}

func (c *CachingClient) vlogf(format string, args ...interface{}) {
	c.Client.vlogf(format, args...)
}

func (c *CachingClient) logf(format string, args ...interface{}) {
	c.Client.logf(format, args...)
}

func recGetKey(r *siser.Record, key string, pErr *error) string {
	if *pErr != nil {
		return ""
	}
	v, ok := r.Get(key)
	if !ok {
		*pErr = fmt.Errorf("didn't find key '%s'", key)
	}
	return v
}

func recGetKeyBytes(r *siser.Record, key string, pErr *error) []byte {
	return []byte(recGetKey(r, key, pErr))
}

func serializeCacheEntry(rr *RequestCacheEntry, prettyPrint bool) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := siser.NewWriter(buf)
	w.NoTimestamp = true
	var r siser.Record
	r.Reset()
	r.Write("Method", rr.Method)
	r.Write("URL", rr.URL)
	r.Write("Body", rr.Body)
	if prettyPrint {
		response := PrettyPrintJS(rr.Response)
		r.Write("Response", string(response))
	} else {
		r.Write("Response", string(rr.Response))
	}
	r.Name = recCacheName
	_, err := w.WriteRecord(&r)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func deserializeCacheEntry(d []byte) ([]*RequestCacheEntry, error) {
	br := bufio.NewReader(bytes.NewBuffer(d))
	r := siser.NewReader(br)
	r.NoTimestamp = true
	var err error
	var res []*RequestCacheEntry
	for r.ReadNextRecord() {
		if r.Name != recCacheName {
			return nil, fmt.Errorf("unexpected record type '%s', wanted '%s'", r.Name, recCacheName)
		}
		rr := &RequestCacheEntry{}
		rr.Method = recGetKey(r.Record, "Method", &err)
		rr.URL = recGetKey(r.Record, "URL", &err)
		rr.Body = recGetKey(r.Record, "Body", &err)
		rr.Response = recGetKeyBytes(r.Record, "Response", &err)
		res = append(res, rr)
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *CachingClient) readRequestsCacheFile(dir string) error {
	timeStart := time.Now()
	c.pageIDToEntries = map[string][]*RequestCacheEntry{}
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	nFiles := 0

	for _, fi := range entries {
		if !fi.Mode().IsRegular() {
			continue
		}
		name := fi.Name()
		if !strings.HasSuffix(name, ".txt") {
			continue
		}
		maybeID := strings.Replace(name, ".txt", "", -1)
		nid := NewNotionID(maybeID)
		if nid == nil {
			continue
		}
		nFiles++
		path := filepath.Join(dir, name)
		d, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		entries, err := deserializeCacheEntry(d)
		if err != nil {
			return err
		}
		c.pageIDToEntries[nid.NoDashID] = entries
	}
	c.vlogf("CachingClient.readRequestsCache: loaded %d files in %s\n", nFiles, time.Since(timeStart))
	return nil
}

func NewCachingClient(cacheDir string, client *Client) (*CachingClient, error) {
	if cacheDir == "" {
		return nil, errors.New("must provide cacheDir")
	}
	if client == nil {
		return nil, errors.New("must provide client")
	}
	res := &CachingClient{
		CacheDir:       cacheDir,
		Client:         client,
		IdToCachedPage: map[string]*CachedPage{},
		Policy:         PolicyDownloadNewer,
	}
	// TODO: ignore error?
	err := res.readRequestsCacheFile(cacheDir)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *CachingClient) getFilesCacheDir() string {
	if c.CacheDirFiles != "" {
		return c.CacheDirFiles
	}
	return filepath.Join(c.CacheDir, "files")
}

func (c *CachingClient) findCachedRequest(pageRequests []*RequestCacheEntry, method string, uri string, body string) (*RequestCacheEntry, bool) {
	panicIf(c.Policy == PolicyDownloadAlways)
	bodyPP := ""
	for _, r := range pageRequests {
		if r.Method != method || r.URL != uri {
			continue
		}

		didFind := r.Body == body
		if !didFind {
			// sometimes (e.g. query param to queryCollection) in request body we use raw values
			// that came from the response. the request might not match when response came
			// from cache (pretty-printed) vs. from network (not pretty-printed)
			// for that reason we also try to match cannonical (pretty-printed) version
			// of request body (should be rare)
			if bodyPP == "" {
				bodyPP = string(PrettyPrintJS([]byte(body)))
			}
			if r.bodyPP == "" {
				r.bodyPP = string(PrettyPrintJS([]byte(r.Body)))
			}
			didFind = (bodyPP == r.bodyPP)
		}
		if didFind {
			c.RequestsFromCache++
			return r, true
		}
	}
	return nil, false
}

func (c *CachingClient) doPostCacheOnly(uri string, body []byte) ([]byte, error) {
	pageID := c.currPageID.NoDashID
	pageRequests := c.pageIDToEntries[pageID]
	r, ok := c.findCachedRequest(pageRequests, "POST", uri, string(body))
	if ok {
		return r.Response, nil
	}
	c.Client.vlogf("CachingClient.findCachedRequest: no cache response for page '%s', url: '%s' in %d cached requests\n", pageID, uri, len(pageRequests))
	return nil, fmt.Errorf("no cache response for '%s' of size %d", uri, len(body))
}

func (c *CachingClient) doPostNoCache(uri string, body []byte) ([]byte, error) {
	d, err := c.Client.doPostInternal(uri, body)
	if err != nil {
		return nil, err
	}
	c.RequestsFromServer++

	if c.currPageID != nil {
		r := &RequestCacheEntry{
			Method:   "POST",
			URL:      uri,
			Body:     string(body),
			Response: d,
		}
		c.currPageRequests = append(c.currPageRequests, r)
		c.needSerializeRequests = true
	}

	return d, nil
}

func (c *CachingClient) getCachedPage(pageID *NotionID) *CachedPage {
	cp := c.IdToCachedPage[pageID.NoDashID]
	if cp == nil {
		cp = &CachedPage{}
		c.IdToCachedPage[pageID.NoDashID] = cp
	}
	return cp
}

// PreLoadCache will preload all pages in the cache.
// It does so concurrently (which is not allowed in general)
// so should be faster
func (c *CachingClient) PreLoadCache() {
	if len(c.IdToCachedPage) > 0 {
		return
	}
	files, err := os.ReadDir(c.CacheDir)
	if err != nil {
		// it's valid, the directoy doesn't have to exist
		return
	}

	var ids []*NotionID
	for _, fi := range files {
		name := fi.Name()
		if strings.HasSuffix(name, ".txt") {
			nid := NewNotionID(strings.Split(name, ".")[0])
			if nid != nil {
				ids = append(ids, nid)
			}
		}
	}
	nThreads := runtime.NumCPU() + 1
	sem := make(chan bool, nThreads)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, id := range ids {
		cachedPage := c.getCachedPage(id)
		client := *c.Client
		sem <- true // enter semaphore
		wg.Add(1)
		go func(client *Client, cp *CachedPage, nid *NotionID) {
			client.httpPostOverride = func(uri string, body []byte) ([]byte, error) {
				pageID := nid.NoDashID
				pageRequests := c.pageIDToEntries[pageID]
				mu.Lock()
				r, ok := c.findCachedRequest(pageRequests, "POST", uri, string(body))
				mu.Unlock()
				if ok {
					return r.Response, nil
				}
				return nil, fmt.Errorf("no cache response for '%s' of size %d", uri, len(body))
			}
			fromCache, _ := client.DownloadPage(nid.NoDashID)
			mu.Lock()
			cp.PageFromCache = fromCache
			mu.Unlock()
			<-sem // leave semaphore
			wg.Done()
		}(&client, cachedPage, id)
	}
	wg.Wait()
}

func (c *CachingClient) DownloadPage(pageID string) (*Page, error) {
	currPageID := NewNotionID(pageID)
	if currPageID == nil {
		return nil, fmt.Errorf("'%s' is not a valid notion id", pageID)
	}
	c.currPageRequests = nil
	c.needSerializeRequests = false

	updateVersions := func(c *CachingClient) {
		if c.didCheckVersions {
			return
		}
		if c.Policy != PolicyDownloadNewer {
			return
		}
		ids := c.GetPageIDs()
		if len(ids) == 0 {
			return
		}
		for i, id := range ids {
			ids[i] = ToNoDashID(id)
		}

		timeStart := time.Now()
		// when we're getting new versions, we have to disable all caching
		c.Client.httpPostOverride = nil
		recVals, err := c.Client.GetBlockRecords(ids)
		if err != nil {
			return
		}
		results := recVals.Results
		if len(results) != len(ids) {
			panic(fmt.Sprintf("updateVersions(): got %d results, expected %d", len(results), len(ids)))
		}
		c.vlogf("CachingClient.updateVersion: got versions for %d pages in %s\n", len(ids), time.Since(timeStart))

		c.didCheckVersions = true
		for i, rec := range results {
			b := rec.Block
			// rec.Block might be nil when a page is not publicly visible or was deleted
			if b != nil {
				id := ids[i]
				if !isIDEqual(id, b.ID) {
					panic(fmt.Sprintf("got result in the wrong order, ids[i]: %s, bid: %s", id, b.ID))
				}
				cp := c.getCachedPage(NewNotionID(id))
				cp.LatestVer = b.Version
			}
		}
	}

	updateVersions(c)

	var err error
	c.currPageID = currPageID
	cp := c.getCachedPage(currPageID)

	writeCacheForCurrPage := func(pageID *NotionID) error {
		var buf []byte

		if !c.needSerializeRequests {
			return nil
		}
		for _, rr := range c.currPageRequests {
			d, err := serializeCacheEntry(rr, !c.NoPrettyPrintResponse)
			if err != nil {
				return err
			}
			buf = append(buf, d...)
		}

		// append to a file for this page
		fileName := pageID.NoDashID + ".txt"
		path := filepath.Join(c.CacheDir, fileName)
		err := ioutil.WriteFile(path, buf, 0644)
		if err != nil {
			// judgement call: delete file if failed to append
			// as it might be corrupted
			// could instead try appendAtomically()
			c.logf("CachingClient.writeCacheForCurrPage: ioutil.WriteFile(%s) failed with '%s'\n", fileName, err)
			os.Remove(path)
			return err
		}
		c.RequestsWrittenToCache += len(c.currPageRequests)
		c.vlogf("CachingClient.writeCacheForCurrPage: wrote %d cached requests to '%s'\n", len(c.currPageRequests), fileName)
		c.currPageRequests = nil
		c.needSerializeRequests = false
		return nil
	}

	timeStart := time.Now()
	fromServer := c.RequestsFromServer
	defer func() {
		if err != nil {
			return
		}
		_ = writeCacheForCurrPage(currPageID)
		c.currPageID = nil
		dur := time.Since(timeStart)
		if fromServer != c.RequestsFromServer {
			c.DownloadedCount++
			c.logf("CachingClient.DownloadPage: downloaded page %s in %s\n", currPageID.DashID, dur)
		} else {
			c.FromCacheCount++
			c.logf("CachingClient.DownloadPage: got page from cache %s in %s\n", currPageID.DashID, dur)
		}
	}()

	if c.Policy == PolicyCacheOnly || c.Policy == PolicyDownloadNewer {
		if cp.PageFromCache == nil {
			c.Client.httpPostOverride = c.doPostCacheOnly
			cp.PageFromCache, err = c.Client.DownloadPage(pageID)
		}
		if c.Policy == PolicyCacheOnly {
			return cp.PageFromCache, err
		}
	}

	fromCache := cp.PageFromCache
	if c.Policy == PolicyDownloadNewer && fromCache != nil {
		latestVer := cp.LatestVer
		fromCacheVer := fromCache.Root().Version
		if fromCacheVer == latestVer {
			return fromCache, nil
		}
	}

	c.Client.httpPostOverride = c.doPostNoCache

	cp.PageFromServer, err = c.Client.DownloadPage(pageID)
	if err != nil {
		if c.Policy == PolicyDownloadNewer && fromCache != nil {
			return fromCache, nil
		}
		return nil, err
	}
	cp.LatestVer = cp.PageFromServer.Root().Version
	return cp.PageFromServer, nil
}

type DownloadInfo struct {
	Page               *Page
	RequestsFromCache  int
	ReqeustsFromServer int
	Duration           time.Duration
	FromCache          bool
}

func (c *CachingClient) DownloadPagesRecursively(startPageID string, afterDownload func(*DownloadInfo) error) ([]*Page, error) {
	toVisit := []*NotionID{NewNotionID(startPageID)}
	downloaded := map[string]*Page{}
	for len(toVisit) > 0 {
		pageID := toVisit[0].NoDashID
		toVisit = toVisit[1:]
		if downloaded[pageID] != nil {
			continue
		}
		nFromCache := c.RequestsFromCache
		nFromServer := c.RequestsFromServer
		timeStart := time.Now()
		page, err := c.DownloadPage(pageID)
		if err != nil {
			return nil, err
		}
		downloaded[pageID] = page
		if afterDownload != nil {
			di := &DownloadInfo{
				Page:               page,
				RequestsFromCache:  c.RequestsFromCache - nFromCache,
				ReqeustsFromServer: c.RequestsFromServer - nFromServer,
				Duration:           time.Since(timeStart),
				FromCache:          c.RequestsFromServer == 0,
			}
			err = afterDownload(di)
			if err != nil {
				return nil, err
			}
		}

		subPages := page.GetSubPages()
		toVisit = append(toVisit, subPages...)
	}
	n := len(downloaded)
	if n == 0 {
		return nil, nil
	}
	var ids []string
	for id := range downloaded {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	pages := make([]*Page, n)
	for i, id := range ids {
		pages[i] = downloaded[id]
	}
	return pages, nil
}

// GetPageIDs returns ids of pages in the cache
func (c *CachingClient) GetPageIDs() []string {
	var res []string
	for id := range c.pageIDToEntries {
		res = append(res, id)
	}
	sort.Strings(res)
	return res
}

func sha1OfURL(uri string) string {
	// TODO: could benefit from normalizing url, e.g. with
	// https://github.com/PuerkitoBio/purell
	h := sha1.New()
	h.Write([]byte(uri))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Returns a name of file in files cache that corresponds
// to a given uri.
// Name of file in cache is sha1(uri) + extension.
// We don't always know the extension, so we need to
// check all file names
func (c *CachingClient) findDownloadedFileInCache(uri string) string {
	dir := c.getFilesCacheDir()
	if len(c.fileNamesInCache) == 0 {
		files, err := os.ReadDir(dir)
		if err != nil {
			return ""
		}
		for _, fi := range files {
			if fi.Type().IsRegular() {
				c.fileNamesInCache = append(c.fileNamesInCache, fi.Name())
			}
		}
	}
	name := sha1OfURL(uri)
	for _, f := range c.fileNamesInCache {
		if strings.HasPrefix(f, name) {
			return filepath.Join(dir, f)
		}
	}
	return ""
}

func guessExt(fileName string, contentType string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".webp", ".bmp", ".tiff", ".svg", ".txt":
		return ext
	}

	contentType = strings.ToLower(contentType)
	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	case "image/svg+xml":
		return ".svg"
	}
	if len(ext) <= 5 {
		// allow any extension of up to 4 chars
		return ext
	}
	panic(fmt.Errorf("didn't find ext for file '%s', content type '%s'", fileName, contentType))
}

// DownloadFile downloads a file refered by block with a given blockID and a parent table
// we cache the file
func (c *CachingClient) DownloadFile(uri string, block *Block) (*DownloadFileResponse, error) {

	var data []byte
	var err error
	// first try to get it from cache
	if c.Policy != PolicyDownloadAlways {
		timeStart := time.Now()
		path := c.findDownloadedFileInCache(uri)
		data, err = ioutil.ReadFile(path)
		if err != nil {
			os.Remove(path)
		} else {
			res := &DownloadFileResponse{
				URL:           uri,
				Data:          data,
				CacheFilePath: path,
				FromCache:     true,
			}
			c.vlogf("CachingClient.DownloadFile: got file from cache '%s' in %s\n", uri, time.Since(timeStart))
			c.FilesFromCacheCount++
			return res, nil
		}
	}

	if c.Policy == PolicyCacheOnly {
		return nil, fmt.Errorf("no cached file for url '%s'", uri)
	}

	timeStart := time.Now()
	c.Client.httpPostOverride = nil
	res, err := c.Client.DownloadFile(uri, block)
	if err != nil {
		c.logf("CachingClient.DownloadFile: failed to download %s, error: %s", uri, err)
		return nil, err
	}
	c.vlogf("CachingClient.DownloadFile: downloaded file '%s' in %s\n", uri, time.Since(timeStart))
	ext := guessExt(uri, res.Header.Get("Content-Type"))
	name := sha1OfURL(uri) + ext
	path := filepath.Join(c.getFilesCacheDir(), name)
	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, 0755)

	err = ioutil.WriteFile(path, res.Data, 0644)
	if err != nil {
		return nil, err
	}
	res.CacheFilePath = path
	c.fileNamesInCache = append(c.fileNamesInCache, name)
	c.DownloadedFilesCount++
	return res, nil
}
