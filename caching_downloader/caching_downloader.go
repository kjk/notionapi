package caching_downloader

import (
	"crypto/sha1"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kjk/caching_http_client"
	"github.com/kjk/notionapi"
)

// EventDidDownload is for logging. Emitted when page
// or file is downloaded
type EventDidDownload struct {
	// if page, PageID is set
	PageID string
	// if file, URL is set
	FileURL string
	// how long did it take to download
	Duration time.Duration
}

// EventError is for logging. Emitted when there's error to log
type EventError struct {
	Error string
}

// EventDidReadFromCache is for logging. Emitted when page
// or file is read from cache.
type EventDidReadFromCache struct {
	// if page, PageID is set
	PageID string
	// if file, URL is set
	FileURL string
	// how long did it take to download
	Duration time.Duration
}

// EventGotVersions is for logging. Emitted
type EventGotVersions struct {
	Count    int
	Duration time.Duration
}

// Downloader implements optimized (cached) downloading
// of pages from the server.
// Cache of pages is stored in CacheDir. We return pages from cache.
// If RedownloadNewerVersions is true, we'll re-download latest version
// of the page (as opposed to returning possibly outdated version
// from cache). We do it more efficiently than just blindly re-downloading.
type Downloader struct {
	Client *notionapi.Client
	// cached pages are stored in Cache as ${pageID}.txt files
	Cache Cache
	// NoReadCache disables reading from cache i.e. downloaded pages
	// will be written to cache but not read from it
	NoReadCache bool
	// if true, we'll re-download a page if a newer version is
	// on the server
	RedownloadNewerVersions bool
	// maps id of the page (in the no-dash format) to a cached Page
	IdToPage map[string]*notionapi.Page
	// maps id of the page (in the no-dash format) to latest version
	// of the page available on the server.
	// if doesn't exist, we haven't yet queried the server for the
	// version
	IdToPageLatestVersion map[string]int64

	didCheckVersionsOfCachedPages bool

	// for diagnostics, number of downloaded pages
	DownloadedCount int
	// number of pages we got from cache
	FromCacheCount int

	// for diagnostics, number of downloaded files
	DownloadedFilesCount int
	// number of files we got from cache
	FilesFromCacheCount int

	EventObserver func(interface{})

	// says if last ReadPageFromCache made http requests
	// (can happen if we tweak the logic)
	didMakeHTTPRequests bool
}

// New returns a new Downloader which caches page loads on disk
// and can return pages from that cache
func New(cache Cache, client *notionapi.Client) *Downloader {
	if client == nil {
		client = &notionapi.Client{}
	}
	res := &Downloader{
		Client:                client,
		Cache:                 cache,
		IdToPage:              make(map[string]*notionapi.Page),
		IdToPageLatestVersion: make(map[string]int64),
	}
	return res
}

func (d *Downloader) useReadCache() bool {
	return !d.NoReadCache
}

// NameForPageID returns name of the file used for storing
// HTTP cache for a given page
func (d *Downloader) NameForPageID(pageID string) string {
	pageID = notionapi.ToNoDashID(pageID)
	return pageID + ".txt"
}

// GetClientCopy returns a copy of client
func (d *Downloader) GetClientCopy() *notionapi.Client {
	var c = *d.Client
	return &c
}

// TODO: maybe split into chunks
func (d *Downloader) getVersionsForPages(ids []string) ([]int64, error) {
	// using new client because we don't want caching of http requests here
	normalizeIDS(ids)
	c := d.GetClientCopy()
	recVals, err := c.GetBlockRecords(ids)
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
	return versions, nil
}

func (d *Downloader) updateVersionsForPages(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	sort.Strings(ids)
	timeStart := time.Now()
	versions, err := d.getVersionsForPages(ids)
	if err != nil {
		return fmt.Errorf("d.updateVersionsForPages() for %d pages failed with '%s'", len(ids), err)
	}
	if len(ids) != len(versions) {
		return fmt.Errorf("d.updateVersionsForPages() asked for %d pages but got %d results", len(ids), len(versions))
	}

	ev := &EventGotVersions{
		Count:    len(ids),
		Duration: time.Since(timeStart),
	}
	d.emitEvent(ev)

	for i := 0; i < len(ids); i++ {
		id := ids[i]
		ver := versions[i]
		id = notionapi.ToNoDashID(id)
		d.IdToPageLatestVersion[id] = ver
	}
	return nil
}

// optimization for RedownloadNewerVersions case: check latest
// versions of all cached pages
func (d *Downloader) checkVersionsOfCachedPages() error {
	if !d.RedownloadNewerVersions {
		return nil
	}
	if d.didCheckVersionsOfCachedPages {
		return nil
	}
	ids, err := d.Cache.GetPageIDs()
	if err != nil {
		// ok to ignore
		return nil
	}
	err = d.updateVersionsForPages(ids)
	if err != nil {
		return err
	}
	d.didCheckVersionsOfCachedPages = true
	return nil
}

// ReadPageFromCache returns a page read from the cache
func (d *Downloader) ReadPageFromCache(pageID string) (*notionapi.Page, error) {
	name := d.NameForPageID(pageID)

	data, err := d.Cache.ReadFile(name)
	if err != nil {
		// it's ok if file doesn't exit
		return nil, nil
	}
	httpCache, err := deserializeHTTPCache(data)
	if err != nil {
		d.Cache.Remove(name)
		return nil, err
	}
	httpCache.CompareNormalizedJSONBody = true
	nPrevRequestsFromCache := httpCache.RequestsNotFromCache
	c := d.GetClientCopy()
	c.HTTPClient = caching_http_client.New(httpCache)
	page, err := c.DownloadPage(pageID)
	if err != nil {
		return nil, err
	}
	d.didMakeHTTPRequests = httpCache.RequestsNotFromCache > nPrevRequestsFromCache

	if d.didMakeHTTPRequests {
		// update cached version
		data, err := serializeHTTPCache(httpCache)
		if err != nil {
			return nil, err
		}
		err = d.Cache.WriteFile(name, data)
		if err != nil {
			d.emitError("ReadPageFromCache(): WriteFile('%s') failed with '%s'\n", name, err)
			d.Cache.Remove(name)
		}
		nNew := httpCache.RequestsNotFromCache - nPrevRequestsFromCache
		d.emitError("ReadPageFromCache(): unexpectedly made %d HTTP requests for page %s\n", nNew, pageID)
	}
	return page, nil
}

// see if this page from in-mmemory cache could be a result based on
// RedownloadNewerVersions
func (d *Downloader) canReturnCachedPage(p *notionapi.Page) bool {
	if p == nil {
		return false
	}
	if !d.RedownloadNewerVersions {
		return true
	}
	pageID := notionapi.ToNoDashID(p.ID)
	if _, ok := d.IdToPageLatestVersion[pageID]; !ok {
		// we don't know what the latest version is, so download it
		err := d.updateVersionsForPages([]string{pageID})
		if err != nil {
			return false
		}
	}
	newestVer := d.IdToPageLatestVersion[pageID]
	pageVer := p.Root().Version
	return pageVer >= newestVer
}

func (d *Downloader) getPageFromCache(pageID string) *notionapi.Page {
	if !d.useReadCache() {
		return nil
	}
	d.checkVersionsOfCachedPages()
	p := d.IdToPage[pageID]
	if d.canReturnCachedPage(p) {
		return p
	}
	p, err := d.ReadPageFromCache(pageID)
	if err != nil {
		return nil
	}
	if d.canReturnCachedPage(p) {
		return p
	}
	return nil
}

// I got "connection reset by peer" error once so retry download 3 times
// with a short sleep in-between
func (d *Downloader) downloadPageRetry(pageID string) (*notionapi.Page, *caching_http_client.Cache, error) {
	var res *notionapi.Page
	var err error
	timeout := time.Second
	for i := 0; i < 3; i++ {
		c := d.GetClientCopy()
		httpCache := caching_http_client.NewCache()
		c.HTTPClient = caching_http_client.New(httpCache)
		res, err = c.DownloadPage(pageID)
		if err == nil {
			return res, httpCache, nil
		}
		// only report errors on the first failure
		if i == 0 {
			d.emitError("Download %s failed with: '%s'\n", pageID, err)
		}
		// don't retry if it can't succeed
		// TODO: probably should change to check for temporary
		// network failures
		if notionapi.IsErrPageNotFound(err) {
			return nil, nil, err
		}
		// hacky: response with 401 code means we don't have access
		if strings.Contains(err.Error(), "status code of 401") {
			return nil, nil, err
		}
		time.Sleep(timeout)
		timeout *= 2
	}
	return nil, nil, err
}

func (d *Downloader) emitEvent(ev interface{}) {
	if d.EventObserver == nil {
		return
	}
	d.EventObserver(ev)
}

func (d *Downloader) emitError(format string, args ...interface{}) {
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	ev := &EventError{
		Error: s,
	}
	d.emitEvent(ev)
}

func (d *Downloader) downloadAndCachePage(pageID string) (*notionapi.Page, error) {
	pageID = notionapi.ToNoDashID(pageID)
	page, httpCache, err := d.downloadPageRetry(pageID)
	if err != nil {
		return nil, err
	}
	data, err := serializeHTTPCache(httpCache)
	if err != nil {
		return nil, err
	}
	name := d.NameForPageID(pageID)
	err = d.Cache.WriteFile(name, data)
	if err != nil {
		d.emitError("Downloader.downloadAndCachePage(): d.Cache.WriteFile('%s') failed with '%s'\n", name, err)
		// ignore file writing error
	}

	return page, nil
}

func (d *Downloader) DownloadPage(pageID string) (*notionapi.Page, error) {
	pageID = notionapi.ToNoDashID(pageID)
	timeStart := time.Now()
	page := d.getPageFromCache(pageID)
	if page == nil {
		var err error
		timeStart = time.Now()
		page, err = d.downloadAndCachePage(pageID)
		if err != nil {
			return nil, err
		}
		d.DownloadedCount++
		ev := &EventDidDownload{
			PageID:   notionapi.ToDashID(pageID),
			Duration: time.Since(timeStart),
		}
		d.emitEvent(ev)
	} else {
		d.FromCacheCount++
		ev := &EventDidReadFromCache{
			PageID:   notionapi.ToDashID(pageID),
			Duration: time.Since(timeStart),
		}
		d.emitEvent(ev)
	}

	d.IdToPage[pageID] = page
	d.IdToPageLatestVersion[pageID] = page.Root().Version
	return page, nil
}

func (d *Downloader) DownloadPagesRecursively(startPageID string, afterDownload func(*notionapi.Page) error) ([]*notionapi.Page, error) {
	toVisit := []string{startPageID}
	downloaded := map[string]*notionapi.Page{}
	for len(toVisit) > 0 {
		pageID := notionapi.ToNoDashID(toVisit[0])
		toVisit = toVisit[1:]
		if downloaded[pageID] != nil {
			continue
		}

		page, err := d.DownloadPage(pageID)
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
	pages := make([]*notionapi.Page, n)
	for i, id := range ids {
		pages[i] = downloaded[id]
	}
	return pages, nil
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

// DownloadFile downloads a file, caching in the cache
func (d *Downloader) DownloadFile(uri string, blockID string) (*notionapi.DownloadFileResponse, error) {
	//fmt.Printf("Downloader.DownloadFile('%s'\n", uri)
	cacheFileName := GetCacheFileNameFromURL(uri)
	var data []byte
	var err error
	if d.useReadCache() {
		timeStart := time.Now()
		data, err = d.Cache.ReadFile(cacheFileName)
		if err != nil {
			d.Cache.Remove(cacheFileName)
		} else {
			res := &notionapi.DownloadFileResponse{
				URL:           uri,
				Data:          data,
				CacheFileName: cacheFileName,
			}
			ev := &EventDidReadFromCache{
				FileURL:  uri,
				Duration: time.Since(timeStart),
			}
			d.emitEvent(ev)
			d.FilesFromCacheCount++
			return res, nil
		}
	}

	timeStart := time.Now()
	res, err := d.Client.DownloadFile(uri, blockID)
	if err != nil {
		d.emitError("Downloader.DownloadFile(): failed to download %s, error: %s", uri, err)
		return nil, err
	}
	ev := &EventDidDownload{
		FileURL:  uri,
		Duration: time.Since(timeStart),
	}
	d.emitEvent(ev)
	_ = d.Cache.WriteFile(cacheFileName, res.Data)
	res.CacheFileName = cacheFileName
	d.DownloadedFilesCount++
	return res, nil
}

func normalizeIDS(ids []string) {
	for i, id := range ids {
		ids[i] = notionapi.ToNoDashID(id)
	}
}

func isIDEqual(id1, id2 string) bool {
	id1 = notionapi.ToNoDashID(id1)
	id2 = notionapi.ToNoDashID(id2)
	return id1 == id2
}
