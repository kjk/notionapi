package notionapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	res, err := parseGetRecordValues([]byte(getRecordValuesJSON1))
	assert.NoError(t, err)
	assert.NotNil(t, res.Results)
	assert.Equal(t, 1, len(res.Results))
	{
		res0 := res.Results[0]
		assert.Equal(t, RoleReader, res0.Role)
		v := res0.Value
		assert.True(t, v.Alive)
		assert.Equal(t, "300db9dc-27c8-4958-a08b-8d0c37f4cfe5", v.ParentID)
		assert.Equal(t, BlockPage, v.Type)
		assert.Equal(t, int64(34), v.Version)
	}
}
