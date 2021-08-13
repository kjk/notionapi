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
	"sort"
	"strings"
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

// CachingClient implements optimized (cached) downloading of pages.
// Cache of pages is stored in CacheDir. We return pages from cache.
// If RedownloadNewerVersions is true, we'll re-download latest version
// of the page (as opposed to returning possibly outdated version
// from cache). We do it more efficiently than just blindly re-downloading.
type CachingClient struct {
	CacheDir string
	Client   *Client

	Policy CachingPolicy

	// disable pretty-printing of json responses saved in the cache
	NoPrettyPrintResponse bool

	// stores pages deserialized just from cache
	IdToPage map[string]*Page

	// maps id of the page (in the no-dash format) to latest version
	// of the page available on the server.
	// if doesn't exist, we haven't yet queried the server for the
	// version
	IdToPageLatestVersion map[string]int64

	DownloadedCount      int
	FromCacheCount       int
	DownloadedFilesCount int
	FilesFromCacheCount  int

	RequestsFromCache        int
	RequestsFromNotionServer int
	RequestsWrittenToCache   int

	pageIDToEntries map[string][]*RequestCacheEntry
	// we cache requests on a per-page basis
	currPageID *NotionID

	currPageRequests      []*RequestCacheEntry
	needSerializeRequests bool
	didCheckVersions      bool
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
	c.vlogf("CachingClient.readRequestsCache(): loaded %d files in %s\n", nFiles, time.Since(timeStart))
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
		CacheDir:              cacheDir,
		Client:                client,
		IdToPage:              map[string]*Page{},
		IdToPageLatestVersion: map[string]int64{},
		Policy:                PolicyDownloadNewer,
	}
	// TODO: ignore error?
	err := res.readRequestsCacheFile(cacheDir)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *CachingClient) findCachedRequest(method string, uri string, body string) (*RequestCacheEntry, bool) {
	panicIf(c.Policy == PolicyDownloadAlways)
	pageID := c.currPageID.NoDashID
	pageRequests := c.pageIDToEntries[pageID]
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
				r.bodyPP = string(PrettyPrintJS([]byte(r.bodyPP)))
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

func (c *CachingClient) writeCacheForCurrPage() error {
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
	fileName := c.currPageID.NoDashID + ".txt"
	path := filepath.Join(c.CacheDir, fileName)
	err := ioutil.WriteFile(path, buf, 0644)
	if err != nil {
		// judgement call: delete file if failed to append
		// as it might be corrupted
		// could instead try appendAtomically()
		c.logf("CachingClient.writeCacheForCurrPage(): ioutil.WriteFile(%s) failed with '%s'\n", fileName, err)
		os.Remove(path)
		return err
	}
	c.RequestsWrittenToCache += len(c.currPageRequests)
	c.vlogf("CachingClient.writeCacheForCurrPage(): wrote %d cached requests to '%s'\n", len(c.currPageRequests), fileName)
	c.currPageRequests = nil
	c.needSerializeRequests = false
	return nil
}

func (c *CachingClient) doPostCacheOnly(uri string, body []byte) ([]byte, error) {
	r, ok := c.findCachedRequest("POST", uri, string(body))
	if ok {
		return r.Response, nil
	}
	c.Client.logf("doPostCacheOnly: no cache response for '%s' with body:\n%s\n", uri, string(body))
	return nil, fmt.Errorf("no cache response for '%s' of size %d", uri, len(body))
}

func (c *CachingClient) doPostNoCache(uri string, body []byte) ([]byte, error) {
	d, err := c.Client.doPostInternal(uri, body)
	if err != nil {
		return nil, err
	}
	c.RequestsFromNotionServer++

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

func (c *CachingClient) doPostMaybeCached(uri string, body []byte) ([]byte, error) {
	r, ok := c.findCachedRequest("POST", uri, string(body))
	if ok {
		// remember requests from cache as well so that when just a single request
		// is different, we don't loose past requests on re-serialization
		c.currPageRequests = append(c.currPageRequests, r)
		return r.Response, nil
	}
	return c.doPostNoCache(uri, body)
}

func (c *CachingClient) getVersionsForPages(ids []string) ([]int64, error) {
	timeStart := time.Now()
	normalizeIDS(ids)
	recVals, err := c.Client.GetBlockRecords(ids)
	if err != nil {
		return nil, err
	}
	results := recVals.Results
	if len(results) != len(ids) {
		return nil, fmt.Errorf("getVersionsForPages(): got %d results, expected %d", len(results), len(ids))
	}
	var versions []int64
	for i, rec := range results {
		// res.Value might be nil when a page is not publicly visible or was deleted
		b := rec.Block
		if b == nil {
			versions = append(versions, 0)
			continue
		}
		id := b.ID
		if !isIDEqual(ids[i], id) {
			panic(fmt.Sprintf("got result in the wrong order, ids[i]: %s, id: %s", ids[0], id))
		}
		versions = append(versions, b.Version)
	}
	c.vlogf("CachingClient.getVersionsForPages(): got info about %d pages in %s\n", len(ids), time.Since(timeStart))
	return versions, nil
}

func (c *CachingClient) updateVersions() error {
	if c.didCheckVersions {
		return nil
	}
	if c.Policy == PolicyCacheOnly {
		return nil
	}
	c.didCheckVersions = true
	ids := c.GetPageIDs()
	if len(ids) == 0 {
		return nil
	}
	timeStart := time.Now()
	versions, err := c.getVersionsForPages(ids)
	if err != nil {
		return fmt.Errorf("d.updateVersionsForPages() for %d pages failed with '%s'", len(ids), err)
	}
	if len(ids) != len(versions) {
		return fmt.Errorf("d.updateVersionsForPages() asked for %d pages but got %d results", len(ids), len(versions))
	}

	c.vlogf("CachingClient.updateVersion: got versions for %d pages in %s\n", len(ids), time.Since(timeStart))

	for i := 0; i < len(ids); i++ {
		id := ids[i]
		ver := versions[i]
		id = ToNoDashID(id)
		c.IdToPageLatestVersion[id] = ver
	}
	return nil
}

func (c *CachingClient) DownloadPage(pageID string) (*Page, error) {
	c.currPageRequests = nil
	c.needSerializeRequests = false
	c.currPageID = NewNotionID(pageID)
	if c.currPageID == nil {
		return nil, fmt.Errorf("'%s' is not a valid notion id", pageID)
	}

	err := c.updateVersions()
	if err != nil {
		return nil, err
	}

	timeStart := time.Now()

	// over-write httpPost only for the duration of client.DownloadPage()
	// that way we don't permanently change the client
	prevOverride := c.Client.httpPostOverride
	defer func() {
		// write out cached requests
		// TODO: what happens if only part of requests were from the cache?
		c.writeCacheForCurrPage()
		c.Client.httpPostOverride = prevOverride
		c.currPageID = nil
	}()
	switch c.Policy {
	case PolicyCacheOnly:
		c.Client.httpPostOverride = c.doPostCacheOnly
	case PolicyDownloadNewer:
		c.Client.httpPostOverride = c.doPostMaybeCached
	case PolicyDownloadAlways:
		c.Client.httpPostOverride = c.doPostNoCache
	}

	fromServer := c.RequestsFromNotionServer
	page, err := c.Client.DownloadPage(pageID)
	if err != nil {
		return nil, err
	}
	if fromServer != c.RequestsFromNotionServer {
		c.DownloadedCount++
		c.logf("CachingClient.DownloadPage: downloaded page %s in %s\n", ToDashID(pageID), time.Since(timeStart))
	} else {
		c.FromCacheCount++
		c.logf("CachingClient.DownloadPage: got page from cache %s in %s\n", ToDashID(pageID), time.Since(timeStart))
	}
	c.IdToPage[pageID] = page
	c.IdToPageLatestVersion[pageID] = page.Root().Version
	return page, nil
}

func (c *CachingClient) DownloadPagesRecursively(startPageID string, afterDownload func(*Page) error) ([]*Page, error) {
	toVisit := []string{startPageID}
	downloaded := map[string]*Page{}
	for len(toVisit) > 0 {
		pageID := ToNoDashID(toVisit[0])
		toVisit = toVisit[1:]
		if downloaded[pageID] != nil {
			continue
		}

		page, err := c.DownloadPage(pageID)
		if err != nil {
			return nil, err
		}
		downloaded[pageID] = page
		if afterDownload != nil {
			err = afterDownload(page)
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

// Sha1OfURL returns sha1 of url
func Sha1OfURL(uri string) string {
	// TODO: could benefit from normalizing url, e.g. with
	// https://github.com/PuerkitoBio/purell
	h := sha1.New()
	h.Write([]byte(uri))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// GetCacheFileNameFromURL returns name of the file in cache
// for this URL. The name is files/${sha1OfURL}.${ext}
// It's a consistent, one-way transform
func GetCacheFileNameFromURL(uri string) string {
	parts := strings.Split(uri, "/")
	name := parts[len(parts)-1]
	ext := filepath.Ext(name)
	ext = strings.ToLower(ext)
	name = Sha1OfURL(uri) + ext
	return filepath.Join("files", name)
}

// DownloadFile downloads a file refered by block with a given blockID and a parent table
// we cache the file
func (c *CachingClient) DownloadFile(uri string, block *Block) (*DownloadFileResponse, error) {
	//fmt.Printf("Downloader.DownloadFile('%s'\n", uri)
	cacheFileName := GetCacheFileNameFromURL(uri)
	path := filepath.Join(c.CacheDir, cacheFileName)
	// ensure dif for file exists
	dir := filepath.Dir(path)
	_ = os.MkdirAll(dir, 0755)

	var data []byte
	var err error
	// first try to get it from cache
	if c.Policy != PolicyDownloadAlways {
		timeStart := time.Now()
		data, err = ioutil.ReadFile(cacheFileName)
		if err != nil {
			os.Remove(cacheFileName)
		} else {
			res := &DownloadFileResponse{
				URL:           uri,
				Data:          data,
				CacheFileName: cacheFileName,
			}
			c.vlogf("CachingClient.DownloadFile: got file from cache '%s' in %s\n", uri, time.Since(timeStart))
			c.FilesFromCacheCount++
			return res, nil
		}
	}

	if c.Policy == PolicyCacheOnly {
		return nil, fmt.Errorf("file '%s' for url '%s' not in cache", cacheFileName, uri)
	}

	timeStart := time.Now()
	res, err := c.Client.DownloadFile(uri, block)
	if err != nil {
		c.logf("CachingClient.DownloadFile: failed to download %s, error: %s", uri, err)
		return nil, err
	}
	c.vlogf("CachingClient.DownloadFile: downloaded file '%s' in %s\n", uri, time.Since(timeStart))
	_ = ioutil.WriteFile(path, res.Data, 0644)
	res.CacheFileName = cacheFileName
	c.DownloadedFilesCount++
	return res, nil
}
