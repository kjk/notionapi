package notionapi

import (
	"testing"

	"github.com/kjk/common/require"
)

const (
	getRecordValuesJSON1 = `{
	"results": [
		{
			"role": "reader",
			"value": {
				"alive": true,
				"content": [
					"c76d351e-e836-4a04-8f09-85c893660b4e"
				],
				"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
				"created_time": 1531024380041,
				"id": "4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d",
				"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
				"last_edited_time": 1531024387094,
				"parent_id": "300db9dc-27c8-4958-a08b-8d0c37f4cfe5",
				"parent_table": "block",
				"properties": {
					"title": [
						[
						"Test page text"
						]
					]
				},
				"type": "page",
				"version": 34
			}
		}
	]
}`
)

func TestGetRecordValues1(t *testing.T) {
	var res GetRecordValuesResponse
	err := jsonit.Unmarshal([]byte(getRecordValuesJSON1), &res)
	require.NoError(t, err)
	require.NotNil(t, res.Results)
	require.Equal(t, 1, len(res.Results))
	for _, rec := range res.Results {
		err = parseRecord(TableBlock, rec)
		require.NoError(t, err)
	}

	{
		res0 := res.Results[0]
		require.Equal(t, RoleReader, res0.Role)
		v := res0.Block
		require.True(t, v.Alive)
		require.Equal(t, "300db9dc-27c8-4958-a08b-8d0c37f4cfe5", v.ParentID)
		require.Equal(t, BlockPage, v.Type)
		require.Equal(t, int64(34), v.Version)
	}
}
