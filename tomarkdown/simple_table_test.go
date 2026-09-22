package tomarkdown

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

func makeTestTable(t *testing.T, columnHeader string) *notionapi.Block {
	table := parseTestBlock(t, `{"id": "t1", "type": "table", "content": ["r1", "r2"], "format": {"table_block_column_order": ["c1", "c2"], "table_block_column_header": `+columnHeader+`}}`)
	table.Content = []*notionapi.Block{
		parseTestBlock(t, `{"id": "r1", "type": "table_row", "properties": {"c1": [["Name"]], "c2": [["Value"]]}}`),
		parseTestBlock(t, `{"id": "r2", "type": "table_row", "properties": {"c1": [["a", [["b"]]]], "c2": [["1 | 2"]]}}`),
	}
	return table
}

func TestRenderSimpleTable(t *testing.T) {
	c := NewConverter(&notionapi.Page{})
	c.PushNewBuffer()
	c.RenderBlock(makeTestTable(t, "true"))
	exp := "| Name | Value |\n| --- | --- |\n| **a** | 1 \\| 2 |\n\n"
	require.Equal(t, exp, strings.TrimLeft(c.PopBuffer().String(), "\n"))

	c = NewConverter(&notionapi.Page{})
	c.PushNewBuffer()
	c.RenderBlock(makeTestTable(t, "false"))
	exp = "| | |\n| --- | --- |\n| Name | Value |\n| **a** | 1 \\| 2 |\n\n"
	require.Equal(t, exp, strings.TrimLeft(c.PopBuffer().String(), "\n"))
}
