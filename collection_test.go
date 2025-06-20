package notionapi

import (
	"testing"

	"github.com/kjk/common/assert"
)

func TestCollection(t *testing.T) {
	c := &Client{}

	notionUrl := "https://www.notion.so/20209ec4150f80f79f2fd3a8965849ef?v=20209ec4150f81be9839000c1487315f&source=copy_link"

	page, err := c.DownloadPage(ExtractNoDashIDFromNotionURL(notionUrl))
	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.Len(t, page.TableViews, 1)

	tv := page.TableViews[0]

	assert.NotNil(t, tv)
	assert.Equal(t, tv.HasMore, true)
	assert.Equal(t, tv.SizeHint > 1000, true)
	assert.Equal(t, len(tv.SpaceShortId) > 5, true)

	limit := (tv.SizeHint + 49) / 50 * 50

	tv, err = c.FetchTableRows(tv, limit)
	assert.NoError(t, err)
	assert.Equal(t, tv.HasMore, false)

	rowsIds := tv.RowIds[tv.SizeHint-10 : tv.SizeHint]

	rows, err := c.FetchTableRowsByIds(tv.SpaceShortId, rowsIds)
	assert.NoError(t, err)
	assert.Len(t, rows, 10)
}
