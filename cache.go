package notionapi

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

func readRequestsCacheFile(dir string) (*RequestsCache, error) {
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
	fmt.Printf("readRequestsCache() loaded %d files in %s\n", nFiles, time.Since(timeStart))
	return c, nil
}

func (c *Client) tryReadFromCache(method string, uri string, body []byte) ([]byte, bool) {
	// no dir, no cache
	if c.CacheDir == "" {
		return nil, false
	}
	var err error
	if c.cache == nil {
		// lazily allocate cache
		c.cache, err = readRequestsCacheFile(c.CacheDir)
		if err != nil {
			return nil, false
		}
	}
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

func appendToFile(path string, d []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(d)
	return err
}

func (c *Client) cacheRequest(method string, uri string, body []byte, response []byte) {
	if c.CacheDir == "" {
		return
	}
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
}
