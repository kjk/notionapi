package caching_downloader

import (
	"testing"

	"github.com/kjk/notionapi"

	"github.com/stretchr/testify/require"
)

/*
Tests that use pages cached in testdata/ directory.
Because they don't involve that they are good for unit tests.
To create the file for a page, run: ./do/do.sh -to-html ${pageID}
and copy data/cache/${pageID}.txt to testdata/
*/

func testDownloadFromCache(t *testing.T, pageID string) *notionapi.Page {
	cache, err := NewDirectoryCache("testdata")
	require.NoError(t, err)
	client := &notionapi.Client{}
	downloader := New(cache, client)
	p, err := downloader.DownloadPage(pageID)
	require.NoError(t, err)
	require.Equal(t, 1, downloader.FromCacheCount)
	require.Equal(t, 0, downloader.DownloadedCount)
	return p
}

func TestPage94167af6567043279811dc923edd1f04(t *testing.T) {
	pid := "94167af6567043279811dc923edd1f04"
	p := testDownloadFromCache(t, pid)
	require.Equal(t, 2, len(p.TableViews))
}
