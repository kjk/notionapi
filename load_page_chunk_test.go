package notion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	loadPageJSON1 = `
	{
		"cursor": {
		  "stack": []
		},
		"recordMap": {
		  "block": {
			"300db9dc-27c8-4958-a08b-8d0c37f4cfe5": {
			  "role": "reader",
			  "value": {
				"alive": true,
				"content": [
				  "c969c945-5d7c-4dd7-9c7f-860f3ace6429",
				  "4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d"
				],
				"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
				"created_time": 1530157233456,
				"format": {
				  "page_full_width": true,
				  "page_small_text": true
				},
				"id": "300db9dc-27c8-4958-a08b-8d0c37f4cfe5",
				"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
				"last_edited_time": 1531024591589,
				"parent_id": "cf9fa7dd-b245-42a0-b929-d5a276b3afe0",
				"parent_table": "space",
				"permissions": [
				  {
					"role": "editor",
					"type": "user_permission",
					"user_id": "bb760e2d-d679-4b64-b2a9-03005b21870a"
				  },
				  {
					"role": "reader",
					"type": "public_permission"
				  }
				],
				"properties": {
				  "title": [
					[
					  "Import Jun 27, 2018"
					]
				  ]
				},
				"type": "page",
				"version": 106
			  }
			},
			"4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d": {
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
			},
			"c76d351e-e836-4a04-8f09-85c893660b4e": {
			  "role": "reader",
			  "value": {
				"alive": true,
				"created_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
				"created_time": 1531024387094,
				"id": "c76d351e-e836-4a04-8f09-85c893660b4e",
				"last_edited_by": "bb760e2d-d679-4b64-b2a9-03005b21870a",
				"last_edited_time": 1531024393188,
				"parent_id": "4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d",
				"parent_table": "block",
				"properties": {
				  "title": [
					[
					  "This is a simple text."
					]
				  ]
				},
				"type": "text",
				"version": 66
			  }
			}
		  },
		  "notion_user": {
			"bb760e2d-d679-4b64-b2a9-03005b21870a": {
			  "role": "reader",
			  "value": {
				"email": "kkowalczyk@gmail.com",
				"family_name": "Kowalczyk",
				"given_name": "Krzysztof",
				"id": "bb760e2d-d679-4b64-b2a9-03005b21870a",
				"locale": "en",
				"mobile_onboarding_completed": true,
				"onboarding_completed": true,
				"profile_photo": "https://s3-us-west-2.amazonaws.com/public.notion-static.com/2dcaa66c-7674-4ff6-9924-601785b63561/head-bw-640x960.png",
				"time_zone": "America/Los_Angeles",
				"version": 9
			  }
			}
		  },
		  "space": {}
		}
	}
`
)

func TestLoadPageChunk1(t *testing.T) {
	res, err := parseLoadPageChunk([]byte(loadPageJSON1))
	assert.NoError(t, err)
	blocks := res.RecordMap.Blocks
	assert.Equal(t, 3, len(blocks))
	{
		v := blocks["4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d"].Value
		assert.True(t, v.Alive)
		assert.Equal(t, "4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d", v.ID)
		assert.Equal(t, "300db9dc-27c8-4958-a08b-8d0c37f4cfe5", v.ParentID)
		assert.Equal(t, TypePage, v.Type)
		assert.Equal(t, int64(34), v.Version)
	}
	{
		v := blocks["c76d351e-e836-4a04-8f09-85c893660b4e"].Value
		assert.True(t, v.Alive)
		assert.Equal(t, "c76d351e-e836-4a04-8f09-85c893660b4e", v.ID)
		assert.Equal(t, "4c6a54c6-8b3e-4ea2-af9c-faabcc88d58d", v.ParentID)
		assert.Equal(t, TypeText, v.Type)
		assert.Equal(t, int64(66), v.Version)

	}
}
