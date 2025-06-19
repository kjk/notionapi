package notionapi

import (
	"testing"

	"github.com/kjk/common/assert"
)

func TestCollection(t *testing.T) {
	c := &Client{}

	notionUrl := "https://www.notion.so/maponai/20f7344323a381fdacc2e77b89ac68e7?v=20f7344323a38173926a000cbccf2d01&source=copy_link"

	page, err := c.DownloadPage(ExtractNoDashIDFromNotionURL(notionUrl))
	assert.NoError(t, err)
	assert.NotNil(t, page)
	assert.Len(t, page.TableViews, 1)

	tv := page.TableViews[0]

	assert.NotNil(t, tv)
	assert.Equal(t, tv.HasMore, true)

	limit := max((tv.SizeHint/100+1)*100, 1000) // max limit is 1000 per request

	tv, err = c.FetchTableRows(tv, limit)
	assert.NoError(t, err)
	assert.Equal(t, tv.HasMore, false)
	assert.Len(t, tv.Rows, tv.SizeHint)
}
