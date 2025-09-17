package notionapi

import (
	"os"
	"testing"

	"github.com/kjk/common/assert"
)

func TestCollection(t *testing.T) {
	c := &Client{}

	notionUrl := os.Getenv("TEST_NOTION_URL")
	if notionUrl == "" {
		t.Skip("TEST_NOTION_URL env var not set")
		return
	}

	page, err := c.DownloadPage(ExtractNoDashIDFromNotionURL(notionUrl))
	assert.NoError(t, err)
	assert.NotNil(t, page)
	// assert.Len(t, page.TableViews, 1)

	tv := page.TableViews[0]

	assert.NotNil(t, tv)
	// assert.Equal(t, tv.HasMore, true)
	assert.Equal(t, tv.SizeHint > 0, true)
	// assert.Equal(t, len(tv.SpaceShortId) > 5, true)

	limit := (tv.SizeHint + 49) / 50 * 50

	tv, err = c.FetchTableRows(tv, limit)
	assert.NoError(t, err)
	// assert.Equal(t, tv.HasMore, false)

	// rowsIds := tv.RowIds[tv.SizeHint-10 : tv.SizeHint]

	// rows, err := c.FetchTableRowsByIds(tv.SpaceShortId, rowsIds)
	// assert.NoError(t, err)
	// assert.Len(t, rows, 10)
}
