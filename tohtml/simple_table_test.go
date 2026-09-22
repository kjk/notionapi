package tohtml

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kjk/common/require"
	"github.com/kjk/notionapi"
)

func parseTestBlock(t *testing.T, s string) *notionapi.Block {
	var b notionapi.Block
	require.NoError(t, json.Unmarshal([]byte(s), &b))
	require.NoError(t, json.Unmarshal([]byte(s), &b.RawJSON))
	return &b
}

func TestRenderSimpleTable(t *testing.T) {
	table := parseTestBlock(t, `{"id": "t1", "type": "table", "content": ["r1", "r2"], "format": {"table_block_column_order": ["c1", "c2"], "table_block_column_header": true}}`)
	table.Content = []*notionapi.Block{
		parseTestBlock(t, `{"id": "r1", "type": "table_row", "properties": {"c1": [["Name"]], "c2": [["Value"]]}}`),
		parseTestBlock(t, `{"id": "r2", "type": "table_row", "properties": {"c1": [["a", [["b"]]]], "c2": [["1 < 2"]]}}`),
	}
	c := NewConverter(&notionapi.Page{})
	c.PushNewBuffer()
	c.RenderBlock(table)
	html := c.PopBuffer().String()
	require.True(t, strings.Contains(html, `<table id="t1" class="simple-table">`), html)
	require.True(t, strings.Contains(html, `<th>Name</th><th>Value</th>`), html)
	require.True(t, strings.Contains(html, `<td><strong>a</strong></td><td>1 &lt; 2</td>`), html)
	require.True(t, strings.Contains(html, `</table>`), html)
}
