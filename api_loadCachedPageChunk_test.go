package notionapi

import (
	"testing"

	"github.com/kjk/common/require"
)

const (
	loadPageJSON1 = `
{
	"cursor": {
		"stack": []
	},
	"recordMap": {
		"block": {
			"0367c2db-381a-4f8b-9ce3-60f388a6b2e3": {
				"role": "reader",
				"value": {
					"alive": true,
					"content": [
						"4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d",
						"f97ffca9-1f89-49b4-8004-999df34ab1f7",
						"6682351e-44bb-4f9c-a0e1-49b703265bdb",
						"42c92ede-8ba2-4c1e-8533-cbfe9d92d98f",
						"97c24351-93d2-4568-8bb5-da7f84edfe45"
					],
					"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_table": "notion_user",
					"created_time": 1531120060950,
					"discussions": ["3342f507-0d13-4f24-9a42-b7951f6fa5f5"],
					"format": {
						"page_cover": "/images/page-cover/rijksmuseum_claesz_1628.jpg",
						"page_cover_position": 0.352,
						"page_full_width": true,
						"page_icon": "🏕",
						"page_small_text": true
					},
					"id": "0367c2db-381a-4f8b-9ce3-60f388a6b2e3",
					"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_table": "notion_user",
					"last_edited_time": 1633890600000,
					"parent_id": "525cd68a-31f3-4e98-a8c1-cb9c39849399",
					"parent_table": "block",
					"permissions": [
						{
							"role": "editor",
							"type": "user_permission",
							"user_id": "bb760e2d-d679-4b64-b2a9-03005b21870a"
						},
						{
							"added_timestamp": 0,
							"allow_duplicate": false,
							"allow_search_engine_indexing": false,
							"role": "reader",
							"type": "public_permission"
						}
					],
					"properties": {
						"title": [["Test pages for notionapi"]]
					},
					"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
					"type": "page",
					"version": 366
				}
			},
			"1790dc30-5a1a-4623-bb87-c080de46d02d": {
				"role": "reader",
				"value": {
					"alive": true,
					"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_table": "notion_user",
					"created_time": 1554788400000,
					"id": "1790dc30-5a1a-4623-bb87-c080de46d02d",
					"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_table": "notion_user",
					"last_edited_time": 1554788400000,
					"parent_id": "4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d",
					"parent_table": "block",
					"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
					"type": "text",
					"version": 4
				}
			},
			"4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d": {
				"role": "reader",
				"value": {
					"alive": true,
					"content": [
						"c76d351e-e836-4a04-8f09-85c893660b4e",
						"7bc42f07-b6e9-406a-bb4f-9d50d68eedb4",
						"6fe7a003-2af0-4c18-bad7-1a3f99caf665",
						"1790dc30-5a1a-4623-bb87-c080de46d02d"
					],
					"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_table": "notion_user",
					"created_time": 1531024380041,
					"id": "4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d",
					"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_table": "notion_user",
					"last_edited_time": 1554788400000,
					"parent_id": "0367c2db-381a-4f8b-9ce3-60f388a6b2e3",
					"parent_table": "block",
					"properties": {
						"title": [["Test text"]]
					},
					"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
					"type": "page",
					"version": 61
				}
			},
			"525cd68a-31f3-4e98-a8c1-cb9c39849399": {
				"role": "reader",
				"value": {
					"alive": true,
					"content": [
						"045c9995-11cf-4eb7-9de5-745d8fc21a3e",
						"0367c2db-381a-4f8b-9ce3-60f388a6b2e3",
						"3b617da4-0945-4a52-bc3a-920ba8832bf7",
						"d6eb49cf-c68f-4028-81af-3aef391443e6",
						"da0b358c-21ab-4ac6-b5c0-f7154b2ecadc"
					],
					"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_table": "notion_user",
					"created_time": 1564868100000,
					"id": "525cd68a-31f3-4e98-a8c1-cb9c39849399",
					"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_table": "notion_user",
					"last_edited_time": 1624230720000,
					"parent_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
					"parent_table": "space",
					"permissions": [
						{
							"role": "editor",
							"type": "user_permission",
							"user_id": "bb760e2d-d679-4b64-b2a9-03005b21870a"
						},
						{
							"added_timestamp": 1624230732521,
							"allow_duplicate": false,
							"role": "reader",
							"type": "public_permission"
						},
						{
							"bot_id": "c9cebcd2-9fc0-4092-aa6e-c2b505c57021",
							"role": {
								"insert_content": true,
								"read_content": true,
								"update_content": true
							},
							"type": "bot_permission"
						}
					],
					"properties": {
						"title": [["Notion testing"]]
					},
					"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
					"type": "page",
					"version": 41
				}
			},
			"6fe7a003-2af0-4c18-bad7-1a3f99caf665": {
				"role": "reader",
				"value": {
					"alive": true,
					"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_table": "notion_user",
					"created_time": 1554788402406,
					"id": "6fe7a003-2af0-4c18-bad7-1a3f99caf665",
					"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_table": "notion_user",
					"last_edited_time": 1554788400000,
					"parent_id": "4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d",
					"parent_table": "block",
					"properties": {
						"title": [["another test"]]
					},
					"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
					"type": "text",
					"version": 16
				}
			},
			"7bc42f07-b6e9-406a-bb4f-9d50d68eedb4": {
				"role": "reader",
				"value": {
					"alive": true,
					"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_table": "notion_user",
					"created_time": 1531033696846,
					"id": "7bc42f07-b6e9-406a-bb4f-9d50d68eedb4",
					"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_table": "notion_user",
					"last_edited_time": 1554788400000,
					"parent_id": "4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d",
					"parent_table": "block",
					"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
					"type": "divider",
					"version": 13
				}
			},
			"c76d351e-e836-4a04-8f09-85c893660b4e": {
				"role": "reader",
				"value": {
					"alive": true,
					"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_table": "notion_user",
					"created_time": 1531024387094,
					"id": "c76d351e-e836-4a04-8f09-85c893660b4e",
					"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"last_edited_by_table": "notion_user",
					"last_edited_time": 1531024393188,
					"parent_id": "4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d",
					"parent_table": "block",
					"properties": {
						"title": [["This is a simple text."]]
					},
					"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
					"type": "text",
					"version": 66
				}
			}
		},
		"comment": {
			"8866ebf3-4a2d-4549-92ab-928c2354e8fc": {
				"role": "reader",
				"value": {
					"alive": true,
					"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
					"created_by_table": "notion_user",
					"created_time": 1566024240000,
					"id": "8866ebf3-4a2d-4549-92ab-928c2354e8fc",
					"last_edited_time": 1566024240000,
					"parent_id": "3342f507-0d13-4f24-9a42-b7951f6fa5f5",
					"parent_table": "discussion",
					"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
					"text": [["a discussion for page\nanother comment about the page"]],
					"version": 6
				}
			}
		},
		"discussion": {
			"3342f507-0d13-4f24-9a42-b7951f6fa5f5": {
				"role": "reader",
				"value": {
					"comments": ["8866ebf3-4a2d-4549-92ab-928c2354e8fc"],
					"id": "3342f507-0d13-4f24-9a42-b7951f6fa5f5",
					"parent_id": "0367c2db-381a-4f8b-9ce3-60f388a6b2e3",
					"parent_table": "block",
					"resolved": false,
					"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
					"version": 1
				}
			}
		},
		"space": {}
	}
}
`
)

func TestLoadCachedPageChunk1(t *testing.T) {
	var rsp LoadCachedPageChunkResponse
	d := []byte(loadPageJSON1)
	err := jsonit.Unmarshal(d, &rsp)
	require.NoError(t, err)
	err = jsonit.Unmarshal(d, &rsp.RawJSON)
	require.NoError(t, err)
	err = ParseRecordMap(rsp.RecordMap)
	require.NoError(t, err)
	blocks := rsp.RecordMap.Blocks
	require.Equal(t, 7, len(blocks))
	for _, rec := range blocks {
		err = parseRecord(TableBlock, rec)
		require.NoError(t, err)
	}
	{
		block := blocks["0367c2db-381a-4f8b-9ce3-60f388a6b2e3"].Block
		require.True(t, block.Alive)
		require.Equal(t, "0367c2db-381a-4f8b-9ce3-60f388a6b2e3", block.ID)
		require.Equal(t, "525cd68a-31f3-4e98-a8c1-cb9c39849399", block.ParentID)
		require.Equal(t, BlockPage, block.Type)
		require.Equal(t, true, block.FormatPage().PageFullWidth)
		require.Equal(t, 0.352, block.FormatPage().PageCoverPosition)
		require.Equal(t, int64(366), block.Version)
		require.Equal(t, 5, len(block.ContentIDs))
	}
}
