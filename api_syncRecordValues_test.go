package notionapi

import (
	"testing"

	"github.com/kjk/common/require"
)

const (
	// TODO: I'm seeing a different format in the browser?
	syncRecordValuesJSON_1 = `
{
	"recordMap": {
    "block": {
      "c3039398-9ae5-49c3-a39f-21ca5a681d72": {
        "role": "reader",
        "value": {
          "alive": true,
          "content": [
            "d28e6c26-bdb5-4c59-9bd2-d791a0dee0e6",
            "36566abc-f77c-42c9-b028-6aa03ec03d04",
            "46c47574-37bf-4139-8d72-4d25211eb55c"
          ],
          "copied_from": "dd5c0a81-3dfe-4487-a6cd-432f82c0c2fc",
          "created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
          "created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
          "created_by_table": "notion_user",
          "created_time": 1570044083803,
          "format": {
            "copied_from_pointer": {
            "id": "dd5c0a81-3dfe-4487-a6cd-432f82c0c2fc",
            "spaceId": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
            "table": "block"
            },
            "page_full_width": true,
            "page_small_text": true
          },
          "id": "c3039398-9ae5-49c3-a39f-21ca5a681d72",
          "last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
          "last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
          "last_edited_by_table": "notion_user",
          "last_edited_time": 1633890540000,
          "parent_id": "0367c2db-381a-4f8b-9ce3-60f388a6b2e3",
          "parent_table": "block",
          "permissions": [
            {
              "added_timestamp": 1633890567665,
              "allow_duplicate": false,
              "allow_search_engine_indexing": false,
              "role": "reader",
              "type": "public_permission"
            }
          ],
          "properties": {
            "title": [["Comparing prices of VPS servers"]]
          },
          "space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
          "type": "page",
          "version": 30
          }
        }
    }
	}
}
`
)

func TestSyncRecordValues1(t *testing.T) {
	var rsp SyncRecordValuesResponse
	d := []byte(syncRecordValuesJSON_1)
	err := jsonit.Unmarshal(d, &rsp)
	require.NoError(t, err)
	err = jsonit.Unmarshal(d, &rsp.RawJSON)
	require.NoError(t, err)
	err = ParseRecordMap(rsp.RecordMap)
	require.NoError(t, err)
	blocks := rsp.RecordMap.Blocks
	require.Equal(t, 1, len(blocks))

	{
		blockV := blocks["c3039398-9ae5-49c3-a39f-21ca5a681d72"]
		block := blockV.Block
		require.Equal(t, BlockPage, block.Type)
		require.Equal(t, 3, len(block.ContentIDs))
		require.Equal(t, true, block.FormatPage().PageFullWidth)
	}

}
