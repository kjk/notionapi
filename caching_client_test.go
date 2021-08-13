package notionapi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

/*
Tests that use pages cached in testdata/ directory.
Because they don't involve that they are good for unit tests.
To create the file for a page, run: ./do/do.sh -to-html ${pageID}
and copy data/cache/${pageID}.txt to testdata/
*/

func testDownloadFromCache(t *testing.T, pageID string) *Page {
	client := &Client{}
	cclient, err := NewCachingClient("caching_client_testdata", client)
	require.NoError(t, err)
	p, err := cclient.DownloadPage(pageID)
	require.NoError(t, err)
	require.True(t, cclient.RequestsFromCache > 0)
	require.Equal(t, 0, cclient.RequestsFromNotionServer)
	return p
}

/*
func convertToMdAndHTML(t *testing.T, page *Page) {
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
*/

// https://www.notion.so/Test-headers-6682351e44bb4f9ca0e149b703265bdb
// test that ForEachBlock() works
func TestPage6682351e44bb4f9ca0e149b703265bdb(t *testing.T) {
	pid := "6682351e44bb4f9ca0e149b703265bdb"
	p := testDownloadFromCache(t, pid)
	blockTypes := []string{}
	cb := func(block *Block) {
		blockTypes = append(blockTypes, block.Type)
	}
	blocks := []*Block{p.Root()}
	ForEachBlock(blocks, cb)
	expected := []string{
		BlockPage,
		BlockHeader,
		BlockSubHeader,
		BlockText,
		BlockSubSubHeader,
		BlockText,
		BlockText,
	}
	require.Equal(t, blockTypes, expected)
}

// https://www.notion.so/Test-table-94167af6567043279811dc923edd1f04
// simple table
func TestPage94167af6567043279811dc923edd1f04(t *testing.T) {
	pid := "94167af6567043279811dc923edd1f04"
	p := testDownloadFromCache(t, pid)
	require.Equal(t, 2, len(p.TableViews))
	//convertToMdAndHTML(t, p)
}

// https://www.notion.so/Test-table-no-title-44f1a38eefe94336907c7576ef4dd19b
// used to crash the API because it has no title column
func TestPage44f1a38eefe94336907c7576ef4dd19b(t *testing.T) {
	// used to crash because has no title column
	pid := "44f1a38eefe94336907c7576ef4dd19b"
	p := testDownloadFromCache(t, pid)
	require.Equal(t, 1, len(p.TableViews))
	//convertToMdAndHTML(t, p)
}
