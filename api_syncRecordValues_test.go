package notionapi

import (
	"testing"

	"github.com/kjk/common/require"
)

const (
	syncRecordValuesJSON_1 = `{
    "recordMap": {
        "__version__": 3,
        "notion_user": {
            "bb760e2d-d679-4b64-b2a9-03005b21870a": {
                "value": {
                    "role": "editor",
                    "value": {
                        "id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
                        "version": 343,
                        "email": "kkowalczyk@gmail.com",
                        "given_name": "Krzysztof",
                        "family_name": "Kowalczyk",
                        "profile_photo": "https://s3-us-west-2.amazonaws.com/public.notion-static.com/2dcaa66c-7674-4ff6-9924-601785b63561/head-bw-640x960.png",
                        "onboarding_completed": true,
                        "mobile_onboarding_completed": true,
                        "clipper_onboarding_completed": true,
                        "name": "Krzysztof Kowalczyk"
                    }
                }
            }
        },
        "user_settings": {
            "bb760e2d-d679-4b64-b2a9-03005b21870a": {
                "value": {
                    "role": "editor",
                    "value": {
                        "id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
                        "version": 54,
                        "settings": {
                            "type": "personal",
                            "locale": "en-US",
                            "source": "social_media",
                            "persona": "programmer",
                            "time_zone": "America/New_York",
                            "used_mac_app": true,
                            "preferred_locale": "en-US",
                            "used_android_app": true,
                            "used_windows_app": true,
                            "start_day_of_week": 0,
                            "used_mobile_web_app": true,
                            "used_desktop_web_app": true,
                            "seen_views_intro_modal": true,
                            "preferred_locale_origin": "legacy",
                            "seen_comment_sidebar_v2": true,
                            "seen_persona_collection": true,
                            "seen_file_attachment_intro": true,
                            "hidden_collection_descriptions": [
                                "5961573f-24db-4fb0-af46-ad7716148db4"
                            ],
                            "created_evernote_getting_started": true
                        }
                    }
                }
            }
        },
        "user_root": {
            "bb760e2d-d679-4b64-b2a9-03005b21870a": {
                "value": {
                    "role": "editor",
                    "value": {
                        "id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
                        "version": 6,
                        "space_views": [
                            "4e548900-40d1-4140-a0a8-165f0c373f6d",
                            "8f7e31df-6cbd-47c9-84b4-043350acf7d8",
                            "4012258b-6235-490c-9e66-f66e8662a836"
                        ],
                        "left_spaces": [
                            "9c94bb91-eea6-410a-972d-9164b8d55e62"
                        ],
                        "space_view_pointers": [
                            {
                                "id": "4e548900-40d1-4140-a0a8-165f0c373f6d",
                                "table": "space_view",
                                "spaceId": "bc202e06-6caa-4e3f-81eb-f226ab5deef7"
                            },
                            {
                                "id": "8f7e31df-6cbd-47c9-84b4-043350acf7d8",
                                "table": "space_view",
                                "spaceId": "cf9fa7dd-b245-42a0-b929-d5a276b3afe0"
                            },
                            {
                                "id": "4012258b-6235-490c-9e66-f66e8662a836",
                                "table": "space_view",
                                "spaceId": "1b527442-0c9e-4459-ab79-83fadc9d1d38"
                            }
                        ]
                    }
                }
            }
        },
        "block": {
            "08e19004-306b-413a-ba6e-0e86a10fec7a": {
                "spaceId": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
                "value": {
                    "role": "editor",
                    "value": {
                        "id": "08e19004-306b-413a-ba6e-0e86a10fec7a",
                        "version": 109,
                        "type": "page",
                        "properties": {
                            "title": [
                                [
                                    "About me"
                                ]
                            ]
                        },
                        "content": [
                            "fea83db6-6cb0-4b59-a468-8c2c659a2afa",
                            "701d13a3-b801-4351-98d8-3e150c152f2e",
                            "0fd3e576-bc00-45e0-a14d-eae340185e95",
                            "fa2fbed4-72c6-4939-9f31-86873516fd26",
                            "c09289ff-ef13-4193-b5d3-cf64015010e6",
                            "75354906-311d-419f-ae09-54250699e295",
                            "35e9af53-e565-417b-8048-e175f86dd6fd",
                            "f3498097-fd87-4f8f-81c0-4a4cba0a52eb"
                        ],
                        "format": {
                            "page_full_width": true,
                            "page_small_text": true,
                            "copied_from_pointer": {
                                "id": "36859b86-c5ac-423e-a037-4f3a4331b814",
                                "table": "block",
                                "spaceId": "bc202e06-6caa-4e3f-81eb-f226ab5deef7"
                            }
                        },
                        "created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
                        "created_time": 1554265410655,
                        "last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
                        "last_edited_time": 1624572660000,
                        "parent_id": "568ac4c0-64c3-4ef6-a6ad-0b8d77230681",
                        "parent_table": "block",
                        "alive": true,
                        "copied_from": "36859b86-c5ac-423e-a037-4f3a4331b814",
                        "created_by_table": "notion_user",
                        "created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
                        "last_edited_by_table": "notion_user",
                        "last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
                        "space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7"
                    }
                }
            },
            "3a6a5733-3884-4095-8ec2-33f3888fd374": {
                "spaceId": "bc202e06-6caa-4e3f-81eb-f226ab5deef7",
                "value": {
                    "role": "editor",
                    "value": {
                        "id": "3a6a5733-3884-4095-8ec2-33f3888fd374",
                        "version": 8,
                        "type": "text",
                        "created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
                        "created_time": 1554267240000,
                        "last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
                        "last_edited_time": 1554267240000,
                        "parent_id": "08e19004-306b-413a-ba6e-0e86a10fec7a",
                        "parent_table": "block",
                        "alive": true,
                        "created_by_table": "notion_user",
                        "created_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
                        "last_edited_by_table": "notion_user",
                        "last_edited_by_id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
                        "space_id": "bc202e06-6caa-4e3f-81eb-f226ab5deef7"
                    }
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
	require.Equal(t, 2, len(blocks))

	{
		blockV := blocks["08e19004-306b-413a-ba6e-0e86a10fec7a"]
		require.Equal(t, "bc202e06-6caa-4e3f-81eb-f226ab5deef7", blockV.SpaceID)
		block := blockV.Value.Block
		require.Equal(t, BlockPage, block.Type)
		require.Equal(t, 8, len(block.ContentIDs))
	}

}
