package caching_downloader

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kjk/caching_http_client"
	"github.com/kjk/notionapi"
)

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

// CachingDownloader implements optimized (cached) downloading
// of pages from the server.
// Cache of pages is stored in CacheDir. We return pages from cache.
// If RedownloadNewerVersions is true, we'll re-download latest version
// of the page (as opposed to returning possibly outdated version
// from cache). We do it more efficiently than just blindly re-downloading.
type CachingDownloader struct {
	Client *notionapi.Client
	// cached pages are stored in CacheDir as ${pageID}.txt files
	CacheDir string
	// NoReadCache disables reading from cache i.e. downloaded pages
	// will be written to cache but not read from it
	NoReadCache bool
	// if true, we'll re-download a page if a newer version is
	// on the server
	RedownloadNewerVersions bool
	// Logger is for debugging, we log progress to logger
	Logger io.Writer
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
}

func (d *CachingDownloader) logf(format string, args ...interface{}) {
	if d.Logger == nil {
		return
	}
	s := format
	if len(args) > 0 {
		s = fmt.Sprintf(format, args...)
	}
	d.Logger.Write([]byte(s))
}

// New returns a new CachingDownloader which caches page loads on disk
// and can return pages from that cache
func New(cacheDir string, client *notionapi.Client) (*CachingDownloader, error) {
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return nil, err
	}
	if client == nil {
		client = &notionapi.Client{}
	}
	res := &CachingDownloader{
		Client:                client,
		CacheDir:              cacheDir,
		IdToPage:              make(map[string]*notionapi.Page),
		IdToPageLatestVersion: make(map[string]int64),
	}
	return res, nil
}

func (d *CachingDownloader) useReadCache() bool {
	return !d.NoReadCache
}

func (d *CachingDownloader) pathForPageID(pageID string) string {
	return filepath.Join(d.CacheDir, pageID+".txt")
}

func (d *CachingDownloader) GetClientCopy() *notionapi.Client {
	var c = *d.Client
	return &c
}

// TODO: maybe split into chunks
func (d *CachingDownloader) getVersionsForPages(ids []string) ([]int64, error) {
	// using new client because we don't want caching of http requests here
	normalizeIDS(ids)
	c := d.GetClientCopy()
	recVals, err := c.GetRecordValues(ids)
	if err != nil {
		return nil, err
	}
	results := recVals.Results
	if len(results) != len(ids) {
		return nil, fmt.Errorf("getVersionsForPages(): got %d results, expected %d", len(results), len(ids))
	}
	var versions []int64
	for i, res := range results {
		// res.Value might be nil when a page is not publicly visible or was deleted
		if res.Value == nil {
			versions = append(versions, 0)
			continue
		}
		id := res.Value.ID
		if !isIDEqual(ids[i], id) {
			panic(fmt.Sprintf("got result in the wrong order, ids[i]: %s, id: %s", ids[0], id))
		}
		versions = append(versions, res.Value.Version)
	}
	return versions, nil
}

func (d *CachingDownloader) updateVersionsForPages(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	sort.Strings(ids)
	versions, err := d.getVersionsForPages(ids)
	if err != nil {
		return fmt.Errorf("d.updateVersionsForPages() for %d pages failed with '%s'\n", len(ids), err)
	}
	if len(ids) != len(versions) {
		return fmt.Errorf("d.updateVersionsForPages() asked for %d pages but got %d results\n", len(ids), len(versions))
	}
	d.logf("Got versions for %d pages\n", len(versions))
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
func (d *CachingDownloader) checkVersionsOfCachedPages() error {
	if !d.RedownloadNewerVersions {
		return nil
	}
	if d.didCheckVersionsOfCachedPages {
		return nil
	}
	var ids []string
	files, err := ioutil.ReadDir(d.CacheDir)
	if err != nil {
		// ok to ignore
		return nil
	}
	for _, fi := range files {
		// skip non-files
		if !fi.Mode().IsRegular() {
			continue
		}
		// valid cache files are in the format:
		// ${pageID}.txt
		parts := strings.Split(fi.Name(), ".")
		if len(parts) != 2 || parts[1] != "txt" {
			continue
		}
		id := notionapi.ToNoDashID(parts[0])
		if !notionapi.IsValidNoDashID(id) {
			d.logf("checkVersionsOfCachedPages: unexpected file '%s' in CacheDir '%s'\n", fi.Name(), d.CacheDir)
			continue
		}
		ids = append(ids, id)
	}
	err = d.updateVersionsForPages(ids)
	if err != nil {
		return err
	}
	d.didCheckVersionsOfCachedPages = true
	return nil
}

func loadHTTPCache(path string) *caching_http_client.Cache {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		// it's ok if file doesn't exit
		return nil
	}
	httpCache, err := deserializeHTTPCache(d)
	if err != nil {
		_ = os.Remove(path)
	}
	return httpCache
}

func (d *CachingDownloader) readPageFromDisk(pageID string) (*notionapi.Page, error) {
	path := d.pathForPageID(pageID)
	httpCache := loadHTTPCache(path)
	if httpCache == nil {
		return nil, nil
	}
	httpCache.CompareNormalizedJSONBody = true
	nPrevRequestsFromCache := httpCache.RequestsNotFromCache
	c := d.GetClientCopy()
	c.HTTPClient = caching_http_client.New(httpCache)
	page, err := c.DownloadPage(pageID)
	if err != nil {
		return nil, err
	}
	newHTTPRequests := httpCache.RequestsNotFromCache - nPrevRequestsFromCache
	if newHTTPRequests > 0 {
		d.logf("CachingDownloader.readPageFromDisk() unexpectedly made %d server connections for page %s", newHTTPRequests, pageID)
	}
	return page, nil
}

func (d *CachingDownloader) canReturnCachedPage(p *notionapi.Page) bool {
	if p == nil {
		return false
	}
	if !d.RedownloadNewerVersions {
		return true
	}
	pageID := notionapi.ToNoDashID(p.ID)
	if _, ok := d.IdToPageLatestVersion[pageID]; !ok {
		// we don't have have latest version
		err := d.updateVersionsForPages([]string{pageID})
		if err != nil {
			return false
		}
	}
	ver := d.IdToPageLatestVersion[pageID]
	pageVer := p.Root().Version
	return pageVer >= ver
}

func (d *CachingDownloader) getPageFromCache(pageID string) *notionapi.Page {
	if !d.useReadCache() {
		return nil
	}
	d.checkVersionsOfCachedPages()
	p := d.IdToPage[pageID]
	if d.canReturnCachedPage(p) {
		return p
	}
	p, err := d.readPageFromDisk(pageID)
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
func (d *CachingDownloader) downloadPageRetry(pageID string) (*notionapi.Page, *caching_http_client.Cache, error) {
	var res *notionapi.Page
	var err error
	for i := 0; i < 3; i++ {
		if i > 0 {
			d.logf("Download %s failed with '%s'\n", pageID, err)
			time.Sleep(5 * time.Second) // not sure if it matters
		}
		c := d.GetClientCopy()
		httpCache := caching_http_client.NewCache()
		c.HTTPClient = caching_http_client.New(httpCache)
		res, err = c.DownloadPage(pageID)
		if err == nil {
			return res, httpCache, nil
		}
	}
	return nil, nil, err
}

func (d *CachingDownloader) downloadAndCachePage(pageID string) (*notionapi.Page, error) {
	pageID = notionapi.ToNoDashID(pageID)

	page, httpCache, err := d.downloadPageRetry(pageID)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(d.CacheDir, pageID+".txt")
	data, err := serializeHTTPCache(httpCache)
	if err != nil {
		return nil, err
	}
	os.MkdirAll(filepath.Dir(path), 0755)
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		d.logf("CachingDownloader.downloadAndCachePage(): ioutil.WriteFile('%s') failed with '%s'\n", path, err)
		// ignore file writing error
	}

	return page, nil
}

func (d *CachingDownloader) DownloadPage(pageID string) (*notionapi.Page, error) {
	pageID = notionapi.ToNoDashID(pageID)
	page := d.getPageFromCache(pageID)
	if page == nil {
		var err error
		page, err = d.downloadAndCachePage(pageID)
		if err != nil {
			return nil, err
		}
		d.DownloadedCount++
		d.logf("%s : downloaded\n", pageID)
	} else {
		d.FromCacheCount++
		d.logf("%s : got from cache\n", pageID)
	}

	d.IdToPage[pageID] = page
	d.IdToPageLatestVersion[pageID] = page.Root().Version
	return page, nil
}

func (d *CachingDownloader) DownloadPagesRecursively(startPageID string) ([]*notionapi.Page, error) {
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

		subPages := notionapi.GetSubPages(page.Root().Content)
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
	pages := make([]*notionapi.Page, n, n)
	for i, id := range ids {
		pages[i] = downloaded[id]
	}
	return pages, nil
}
