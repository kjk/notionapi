package caching_downloader

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/kjk/notionapi"
)

// Cache describes a caching interface
type Cache interface {
	ReadFile(string) ([]byte, error)
	WriteFile(string, []byte) error
	GetPageIDs() ([]string, error)
	Remove(string)
}

var _ Cache = &DirectoryCache{}

// DirectoryCache implements disk-based Cache interface
type DirectoryCache struct {
	Dir string
}

func (c *DirectoryCache) ReadFile(name string) ([]byte, error) {
	path := filepath.Join(c.Dir, name)
	return ioutil.ReadFile(path)
}

func (c *DirectoryCache) WriteFile(name string, data []byte) error {
	path := filepath.Join(c.Dir, name)
	return ioutil.WriteFile(path, data, 0644)
}

func (c *DirectoryCache) Remove(name string) {
	path := filepath.Join(c.Dir, name)
	os.Remove(path)
}

func (c *DirectoryCache) GetPageIDs() ([]string, error) {
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

func NewDirectoryCache(dir string) (*DirectoryCache, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}
	return &DirectoryCache{
		Dir: dir,
	}, nil
}
