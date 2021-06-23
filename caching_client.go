package notionapi

import (
	"bufio"
	"bytes"
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

// RequestCacheEntry has info about request (method/url/body) and response
type RequestCacheEntry struct {
	// request info
	Method string
	URL    string
	Body   []byte
	// response
	Response []byte
}

type RequestsCache struct {
	dir string
	// allEntries      []*RequestCacheEntry
	pageIDToEntries map[string][]*RequestCacheEntry
	// we cache requests on a per-page basis
	currPageID *NotionID
}

type CachingClient struct {
	CacheDir         string
	client           *Client
	DisableCacheRead bool // TODO: rename DisableReadsFromCache
	cache            *RequestsCache

	RequestsFromCache      int
	RequestsNotFromCache   int
	RequestsWrittenToCache int
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

func serializeCacheEntry(rr *RequestCacheEntry) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	w := siser.NewWriter(buf)
	w.NoTimestamp = true
	var r siser.Record
	r.Reset()
	r.Write("Method", rr.Method)
	r.Write("URL", rr.URL)
	r.Write("Body", string(rr.Body))
	response := PrettyPrintJS(rr.Response)
	r.Write("Response", string(response))
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
		rr.Body = recGetKeyBytes(r.Record, "Body", &err)
		rr.Response = recGetKeyBytes(r.Record, "Response", &err)
		res = append(res, rr)
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *CachingClient) logf(format string, args ...interface{}) {
	c.client.logf(format, args...)
}

func (c *CachingClient) vlogf(format string, args ...interface{}) {
	c.client.vlogf(format, args...)
}

func (cc *CachingClient) readRequestsCacheFile(dir string) (*RequestsCache, error) {
	c := &RequestsCache{
		dir:             dir,
		pageIDToEntries: map[string][]*RequestCacheEntry{},
	}
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	timeStart := time.Now()
	entries, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		entries, err := deserializeCacheEntry(d)
		if err != nil {
			return nil, err
		}
		c.pageIDToEntries[nid.NoDashID] = entries
	}
	cc.vlogf("readRequestsCache() loaded %d files in %s\n", nFiles, time.Since(timeStart))
	return c, nil
}

func NewCachingClient(cacheDir string, client *Client) (*CachingClient, error) {
	if cacheDir == "" {
		return nil, errors.New("must provide cacheDir")
	}
	if client == nil {
		return nil, errors.New("must provide client")
	}
	res := &CachingClient{
		CacheDir: cacheDir,
		client:   client,
	}
	// TODO: ignore error?
	cache, err := res.readRequestsCacheFile(cacheDir)
	if err != nil {
		return nil, err
	}
	res.cache = cache
	client.httpPostOverride = res.doPostMaybeCached
	return res, nil
}

func (c *CachingClient) tryReadFromCache(method string, uri string, body []byte) ([]byte, bool) {
	if c.DisableCacheRead {
		return nil, false
	}
	pageID := c.cache.currPageID.NoDashID
	pageRequests := c.cache.pageIDToEntries[pageID]
	for _, r := range pageRequests {
		if r.Method != method {
			continue
		}
		if r.URL != uri {
			continue
		}
		if !bytes.Equal(r.Body, body) {
			continue
		}
		return r.Response, true
	}
	return nil, false
}

func (c *CachingClient) cacheRequest(method string, uri string, body []byte, response []byte) {
	// this is not in the context of any page so we don't cache it
	if c.cache.currPageID == nil {
		return
	}
	rr := &RequestCacheEntry{
		Method:   method,
		URL:      uri,
		Body:     body,
		Response: response,
	}
	d, err := serializeCacheEntry(rr)
	if err != nil {
		return
	}

	// append to a file for this page
	fileName := c.cache.currPageID.NoDashID + ".txt"
	path := filepath.Join(c.CacheDir, fileName)
	err = appendToFile(path, d)
	if err != nil {
		// judgement call: delete file if failed to append
		// as it might be corrupted
		// could instead try appendAtomically()
		os.Remove(path)
		return
	}
	c.RequestsWrittenToCache++
}

func (c *CachingClient) doPostMaybeCached(uri string, body []byte) ([]byte, error) {
	d, ok := c.tryReadFromCache("POST", uri, body)
	if ok {
		c.RequestsFromCache++
		return d, nil
	}
	d, err := c.client.doPostInternal(uri, body)
	if err != nil {
		return nil, err
	}
	c.RequestsNotFromCache++

	c.cacheRequest("POST", uri, body, d)
	return d, nil
}

func (c *CachingClient) DownloadPage(pageID string) (*Page, error) {
	c.cache.currPageID = NewNotionID(pageID)
	if c.cache.currPageID == nil {
		return nil, fmt.Errorf("'%s' is not a valid notion id", pageID)
	}
	return c.client.DownloadPage(pageID)
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
	for id := range c.cache.pageIDToEntries {
		res = append(res, id)
	}
	return res
}
