package notionapi

import (
	"encoding/json"
	"testing"

	"github.com/kjk/common/require"
)

// builds a BlockTable with 2 rows and 2 columns from JSON in the
// shape returned by Notion API
func makeTestSimpleTable(t *testing.T, columnHeader bool) *Block {
	tableJSON := `{
		"id": "t1",
		"type": "table",
		"content": ["r1", "r2"],
		"format": {
			"table_block_column_order": ["c1", "c2"],
			"table_block_column_header": ` + boolStr(columnHeader) + `,
			"table_block_row_header": false,
			"table_block_column_format": {"c1": {"width": 120}}
		}
	}`
	row1JSON := `{"id": "r1", "type": "table_row", "properties": {"c1": [["Name"]], "c2": [["Value"]]}}`
	row2JSON := `{"id": "r2", "type": "table_row", "properties": {"c1": [["a", [["b"]]]], "c2": [["1 | 2"]]}}`

	parse := func(s string) *Block {
		var b Block
		require.NoError(t, json.Unmarshal([]byte(s), &b))
		require.NoError(t, json.Unmarshal([]byte(s), &b.RawJSON))
		return &b
	}
	table := parse(tableJSON)
	table.Content = []*Block{parse(row1JSON), parse(row2JSON)}
	return table
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func TestSimpleTable(t *testing.T) {
	table := makeTestSimpleTable(t, true)
	f := table.FormatSimpleTable()
	require.NotNil(t, f)
	require.True(t, f.ColumnHeader)
	require.Equal(t, []string{"c1", "c2"}, table.TableColumnIDs())
	require.Equal(t, 120.0, f.ColumnFormat["c1"].Width)

	rows := table.TableRows()
	require.Equal(t, 2, len(rows))
	require.Equal(t, "Name", TextSpansToString(rows[0].TableCell("c1")))
	require.Equal(t, "1 | 2", TextSpansToString(rows[1].TableCell("c2")))
	require.Equal(t, AttrBold, rows[1].TableCell("c1")[0].Attrs[0][0])
	require.Nil(t, rows[1].TableCell("no-such-column"))

	// without column order in format, columns come from the first row
	delete(table.RawJSON["format"].(map[string]interface{}), "table_block_column_order")
	require.Equal(t, []string{"c1", "c2"}, table.TableColumnIDs())
}
