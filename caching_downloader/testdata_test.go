package caching_downloader

import (
	"testing"

	"github.com/kjk/notionapi/tohtml"

	"github.com/kjk/notionapi/tomarkdown"

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

func convertToMdAndHTML(t *testing.T, page *notionapi.Page) {
	{
		conv := tomarkdown.NewConverter(page)
		md := conv.ToMarkdown()
		require.NotEmpty(t, md)
	}

	{
		conv := tohtml.NewConverter(page)
		html, err := conv.ToHTML()
		require.NoError(t, err)
		require.NotEmpty(t, html)
	}
}

func TestPage94167af6567043279811dc923edd1f04(t *testing.T) {
	pid := "94167af6567043279811dc923edd1f04"
	p := testDownloadFromCache(t, pid)
	require.Equal(t, 2, len(p.TableViews))
	convertToMdAndHTML(t, p)
}

func TestPage44f1a38eefe94336907c7576ef4dd19b(t *testing.T) {
	// used to crash because has no title column
	pid := "44f1a38eefe94336907c7576ef4dd19b"
	p := testDownloadFromCache(t, pid)
	require.Equal(t, 1, len(p.TableViews))
	convertToMdAndHTML(t, p)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// benchmark JSON decoding of requests
func BenchmarkJSONDecode(b *testing.B) {
	pid := "44f1a38eefe94336907c7576ef4dd19b"
	for i := 0; i < b.N; i++ {
		cache, err := NewDirectoryCache("testdata")
		must(err)
		client := &notionapi.Client{}
		downloader := New(cache, client)
		_, err = downloader.DownloadPage(pid)
		must(err)
	}
}
