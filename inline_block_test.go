package notionapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseBlocks(t *testing.T, s string) []*InlineBlock {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	assert.NoError(t, err)
	blocks, err := parseInlineBlocks(m["title"])
	assert.NoError(t, err)
	return blocks
}

const title1 = `{
	"title": [
	  [ "Test page text" ]
	]
}`

func TestParseInlineBlock1(t *testing.T) {
	blocks := parseBlocks(t, title1)
	assert.Equal(t, 1, len(blocks))
	b := blocks[0]
	assert.Equal(t, "Test page text", b.Text)
	assert.True(t, b.IsPlain())
}

const title2 = `{
	"title": [
	  [
		"‣",
		[
		  [
			"u",
			"bb760e2d-d679-4b64-b2a9-03005b21870a"
		  ]
		]
	  ]
	]
}`

func TestParseInlineBlock2(t *testing.T) {
	blocks := parseBlocks(t, title2)
	assert.Equal(t, 1, len(blocks))
	b := blocks[0]
	assert.Equal(t, InlineAt, b.Text)
	assert.Equal(t, 0, int(b.AttrFlags))
	assert.Equal(t, "bb760e2d-d679-4b64-b2a9-03005b21870a", b.UserID)
	assert.False(t, b.IsPlain())
}

const title3 = `{
	"title": [
		["Text block with "],
		[
		  "bold ",
		  [
			["b"]
		  ]
		]
	]
}`

func TestParseInlineBlock3(t *testing.T) {
	blocks := parseBlocks(t, title3)
	assert.Equal(t, 2, len(blocks))
	{
		b := blocks[0]
		assert.Equal(t, "Text block with ", b.Text)
		assert.Equal(t, 0, int(b.AttrFlags))
	}

	{
		b := blocks[1]
		assert.Equal(t, "bold ", b.Text)
		assert.Equal(t, AttrFlag(AttrBold), b.AttrFlags)
		assert.False(t, b.IsPlain())
	}
}

const title4 = `{
	"title": [
		[
			"link inside bold",
			[
			  ["b"],
			  [
				"a",
				"https://www.google.com"
			  ]
			]
		]
	]
}`

func TestParseInlineBlock4(t *testing.T) {
	blocks := parseBlocks(t, title4)
	assert.Equal(t, 1, len(blocks))
	{
		b := blocks[0]
		assert.Equal(t, "link inside bold", b.Text)
		assert.Equal(t, AttrFlag(AttrBold), b.AttrFlags)
		assert.Equal(t, "https://www.google.com", b.Link)
		assert.False(t, b.IsPlain())
	}
}

const title5 = `{
	"title": [
		[
			"‣",
			[
			  [
				"d",
				{
				  "date_format": "relative",
				  "start_date": "2018-07-17",
				  "start_time": "15:00",
				  "time_zone": "America/Los_Angeles",
				  "type": "datetime"
				}
			  ]
			]
		]
	]
}`

func TestParseInlineBlock5(t *testing.T) {
	blocks := parseBlocks(t, title5)
	assert.Equal(t, 1, len(blocks))
	b := blocks[0]
	assert.Equal(t, InlineAt, b.Text)
	assert.Equal(t, 0, int(b.AttrFlags))
	assert.Equal(t, b.Date.DateFormat, "relative")
	assert.False(t, b.IsPlain())
}

const titleBig = `{
	"title": [
	  ["Text block with "],
	  [
		"bold ",
		[
		  ["b"]
		]
	  ],
	  [
		"link inside bold",
		[
		  ["b"],
		  [
			"a",
			"https://www.google.com"
		  ]
		]
	  ],
	  [
		" text",
		[
		  ["b"]
		]
	  ],
	  [", "],
	  [
		"italic text",
		[
		  ["i"]
		]
	  ],
	  [", "],
	  [
		"strikethrough text",
		[
		  ["s"]
		]
	  ],
	  [", "],
	  [
		"code part",
		[
		  ["c"]
		]
	  ],
	  [", "],
	  [
		"link part",
		[
		  [
			"a",
			"http://blog.kowalczyk.info"
		  ]
		]
	  ],
	  [" , "],
	  [
		"‣",
		[
		  [
			"u",
			"bb760e2d-d679-4b64-b2a9-03005b21870a"
		  ]
		]
	  ],
	  [" and "],
	  [
		"‣",
		[
		  [
			"d",
			{
			  "date_format": "relative",
			  "start_date": "2018-07-17",
			  "start_time": "15:00",
			  "time_zone": "America/Los_Angeles",
			  "type": "datetime"
			}
		  ]
		]
	  ],
	  [" and that's it."]
	]
}`

func TestParseInlineBlockBig(t *testing.T) {
	blocks := parseBlocks(t, titleBig)
	assert.Equal(t, 17, len(blocks))
}
