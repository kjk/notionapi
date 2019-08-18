package caching_downloader

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/kjk/notionapi"
)

// Cache describes a caching interface
type Cache interface {
	// ReadFile reads a file with a given name from cache
	ReadFile(string) ([]byte, error)
	// WriteFile writes a file with a given name to cache
	WriteFile(string, []byte) error
	// GetPageIDs returns ids of pages in the cache
	GetPageIDs() ([]string, error)
	// Remove removes a file with a given name from cache
	Remove(string)
}

var _ Cache = &DirectoryCache{}

// DirectoryCache implements disk-based Cache interface
type DirectoryCache struct {
	Dir string
	mu  sync.Mutex
}

// ReadFile reads a file with a given name from cache
func (c *DirectoryCache) ReadFile(name string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := filepath.Join(c.Dir, name)
	return ioutil.ReadFile(path)
}

// WriteFile writes a file with a given name to cache
func (c *DirectoryCache) WriteFile(name string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := filepath.Join(c.Dir, name)

	// make sure directory for a file exits
	// ok to ignore error as WriteFile will fail too
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	return ioutil.WriteFile(path, data, 0644)
}

// Remove removes a file with a given name from cache
func (c *DirectoryCache) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := filepath.Join(c.Dir, name)
	os.Remove(path)
}

// GetPageIDs returns ids of pages in the cache
func (c *DirectoryCache) GetPageIDs() ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	files, err := ioutil.ReadDir(c.Dir)
	if err != nil {
		return nil, err
	}
	var ids []string
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
			//d.logf("checkVersionsOfCachedPages: unexpected file '%s' in CacheDir '%s'\n", fi.Name(), d.CacheDir)
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// NewDirectoryCache returns a new DirectoryCache which caches files
// in a directory
func NewDirectoryCache(dir string) (*DirectoryCache, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	return &DirectoryCache{
		Dir: dir,
	}, nil
}
