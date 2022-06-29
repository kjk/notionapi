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
			"__version__": 3,
			"block": {
				"568ac4c0-64c3-4ef6-a6ad-0b8d77230681": {
					"value": {
						"role": "editor",
						"value": {
							"id": "568ac4c0-64c3-4ef6-a6ad-0b8d77230681",
							"version": 423,
							"type": "page",
							"properties": {
								"title": [
									[
										"Website"
									]
								]
							},
							"content": [
								"08e19004-306b-413a-ba6e-0e86a10fec7a",
								"623523b6-7e15-48a0-b525-749d6921465c",
								"f5144c42-a3e5-4e47-9466-2c8ecdc0bcb6",
								"30da6655-040f-47e6-93f8-e66eacd308c1"
							],
							"format": {
								"page_full_width": true,
								"page_small_text": true
							},
							"permissions": [
								{
									"role": "editor",
									"type": "user_permission",
									"user_id": "bb760e2d-d679-4b64-b2a9-03005b21870a"
								},
								{
									"role": {
										"read_content": true
									},
									"type": "bot_permission",
									"bot_id": "161740aa-8f29-495f-8507-66cb8bb9c365"
								}
							],
							"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"created_time": 1528059171080,
							"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"last_edited_time": 1655141640000,
							"parent_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
							"parent_table": "space",
							"alive": true,
							"created_by_table": "notion_user",
							"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"last_edited_by_table": "notion_user",
							"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7"
						}
					}
				},
				"d61b4f94-b10d-4d80-8d3d-238a4e7c4d10": {
					"value": {
						"role": "editor",
						"value": {
							"id": "d61b4f94-b10d-4d80-8d3d-238a4e7c4d10",
							"version": 7,
							"type": "page",
							"properties": {
								"title": [
									[
										"Programming"
									]
								]
							},
							"content": [
								"f60311c9-c6ed-4678-a1e8-497fb0ab8545"
							],
							"format": {
								"page_full_width": true,
								"page_small_text": true
							},
							"permissions": [
								{
									"role": "editor",
									"type": "user_permission",
									"user_id": "bb760e2d-d679-4b64-b2a9-03005b21870a"
								}
							],
							"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"created_time": 1481054370082,
							"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"last_edited_time": 1528059358122,
							"parent_id": "568ac4c0-64c3-4ef6-a6ad-0b8d77230681",
							"parent_table": "block",
							"alive": true,
							"created_by_table": "notion_user",
							"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"last_edited_by_table": "notion_user",
							"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7"
						}
					}
				},
				"aea20e01-890c-4874-ae08-4557d7789195": {
					"value": {
						"role": "editor",
						"value": {
							"id": "aea20e01-890c-4874-ae08-4557d7789195",
							"version": 48,
							"type": "text",
							"properties": {
								"title": [
									[
										"Programming:"
									]
								]
							},
							"content": [
								"6f70163e-a5b8-4ba9-928a-faa2e45d1f51",
								"ed055f63-753e-42ef-9025-e11ac9062c35"
							],
							"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"created_time": 1530068313902,
							"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"last_edited_time": 1554270960000,
							"parent_id": "568ac4c0-64c3-4ef6-a6ad-0b8d77230681",
							"parent_table": "block",
							"alive": true,
							"created_by_table": "notion_user",
							"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"last_edited_by_table": "notion_user",
							"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7"
						}
					}
				},
				"03ece883-f7df-4ce7-8596-73d04811479e": {
					"value": {
						"role": "editor",
						"value": {
							"id": "03ece883-f7df-4ce7-8596-73d04811479e",
							"version": 231,
							"type": "page",
							"properties": {
								"title": [
									[
										"This developer's life"
									]
								]
							},
							"content": [
								"da8bb0eb-517b-42f6-bd47-243c353a26cf",
								"5b4d7c6e-1895-4e5c-b120-e33ae53e9413",
								"b7c27f52-4ad1-4a5e-9dbc-5c1b6db0ee4a",
								"751d064a-76bf-46fa-9b29-3da1a88cd033"
							],
							"format": {
								"block_locked": false,
								"block_locked_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
								"page_full_width": true,
								"page_small_text": true
							},
							"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"created_time": 1554244717057,
							"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"last_edited_time": 1655141640000,
							"parent_id": "568ac4c0-64c3-4ef6-a6ad-0b8d77230681",
							"parent_table": "block",
							"alive": true,
							"created_by_table": "notion_user",
							"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"last_edited_by_table": "notion_user",
							"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7"
						}
					}
				},
				"30da6655-040f-47e6-93f8-e66eacd308c1": {
					"value": {
						"role": "editor",
						"value": {
							"id": "30da6655-040f-47e6-93f8-e66eacd308c1",
							"version": 937,
							"type": "page",
							"properties": {
								"title": [
									[
										"Diary of a solo dev building a web app"
									]
								]
							},
							"content": [
								"1e323661-7746-4761-b561-ffccc1a3ce74",
								"9b975c30-e16b-46e1-b235-63fa1942a264",
								"dabd3815-292a-4dcc-bc94-4a97d69c6181"
							],
							"format": {
								"page_cover": "https://s3-us-west-2.amazonaws.com/secure.notion-static.com/56c80314-5a27-4407-8980-bbc4a1a9929d/cover-2.jpg",
								"page_full_width": true,
								"page_small_text": true,
								"page_cover_position": 0.5
							},
							"created_time": 1652518422592,
							"last_edited_time": 1656481440000,
							"parent_id": "568ac4c0-64c3-4ef6-a6ad-0b8d77230681",
							"parent_table": "block",
							"alive": true,
							"file_ids": [
								"48387103-f51d-43b8-ab23-3ff320312aa9",
								"56c80314-5a27-4407-8980-bbc4a1a9929d"
							],
							"created_by_table": "notion_user",
							"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"last_edited_by_table": "notion_user",
							"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7"
						}
					}
				}
			},
			"space": {
				"bc202e06-6caa-4e3f-81eb-f226ab5deef7": {
					"spaceId": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
					"value": {
						"role": "editor",
						"value": {
							"id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
							"version": 346,
							"name": "Main",
							"permissions": [
								{
									"role": "editor",
									"type": "user_permission",
									"user_id": "bb760e2d-d679-4b64-b2a9-03005b21870a"
								}
							],
							"icon": "https://s3-us-west-2.amazonaws.com/public.notion-static.com/23900417-cede-4b11-8a9d-4dfa679dfcbd/head-bw-sq-120.png",
							"beta_enabled": false,
							"pages": [
								"381f674e-a6ce-4131-9d98-90967b1d3c14",
								"581f70b2-4002-4e31-a140-46f50934535d",
								"6667f3d0-4141-41ae-9ea2-c2f922acab18"
							],
							"disable_public_access": false,
							"disable_guests": false,
							"disable_move_to_space": false,
							"disable_export": false,
							"created_time": 1537253154666,
							"last_edited_time": 1652978220000,
							"created_by_table": "notion_user",
							"created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"last_edited_by_table": "notion_user",
							"last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
							"plan_type": "personal",
							"invite_link_enabled": false
						}
					}
				}
			}
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
	require.Equal(t, 5, len(blocks))
	for _, rec := range blocks {
		err = parseRecord(TableBlock, rec.Value)
		require.NoError(t, err)
	}
	{
		block := blocks["568ac4c0-64c3-4ef6-a6ad-0b8d77230681"].Value.Block
		require.True(t, block.Alive)
		require.Equal(t, "568ac4c0-64c3-4ef6-a6ad-0b8d77230681", block.ID)
		require.Equal(t, "bc202e06-6caa-4e3f-81eb-f226ab5deef7", block.ParentID)
		require.Equal(t, BlockPage, block.Type)
		require.Equal(t, true, block.FormatPage().PageFullWidth)
		require.Equal(t, false, block.FormatPage().BlockLocked)
		require.Equal(t, int64(423), block.Version)
		require.Equal(t, 4, len(block.ContentIDs))
	}
	{
		block := blocks["aea20e01-890c-4874-ae08-4557d7789195"].Value.Block
		require.True(t, block.Alive)
		require.Equal(t, "aea20e01-890c-4874-ae08-4557d7789195", block.ID)
		require.Equal(t, "568ac4c0-64c3-4ef6-a6ad-0b8d77230681", block.ParentID)
		require.Equal(t, BlockText, block.Type)
		require.Equal(t, int64(48), block.Version)
		require.Equal(t, 2, len(block.ContentIDs))
		require.Equal(t, "bc202e06-6caa-4e3f-81eb-f226ab5deef7", block.SpaceID)
	}

	{
		block := blocks["30da6655-040f-47e6-93f8-e66eacd308c1"].Value.Block
		require.Equal(t, 2, len(block.FileIDs))
		require.Equal(t, "56c80314-5a27-4407-8980-bbc4a1a9929d", block.FileIDs[1])
	}
}
