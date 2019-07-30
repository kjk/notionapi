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
	blocks, err := ParseInlineBlocks(m["title"])
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

const titleWithComment = `{
	"title": [
	[
		"Just"
	],
	[
		"comment",
		[
			[
				"m",
				"4a1cc3be-03cf-489a-9542-69d9a02f3534"
			]
		]
	],
	[
		"inline."
	]
]
}
`

func TestParseInlineBlockComment(t *testing.T) {
	blocks := parseBlocks(t, titleWithComment)
	assert.Equal(t, 3, len(blocks))

	{
		// "Just"
		b := blocks[0]
		assert.Equal(t, b.Text, "Just")
		assert.Equal(t, int(b.AttrFlags), 0)
	}
	{
		// "comment"
		b := blocks[1]
		assert.Equal(t, b.Text, "comment")
		assert.Equal(t, int(b.AttrFlags), 0)
		assert.Equal(t, b.CommentID, "4a1cc3be-03cf-489a-9542-69d9a02f3534")
	}

}

const title6 = `{
	"title": [
		[
			"colored",
			[
				[
					"h",
					"teal_background"
				]
			]
		],
		[
			"text",
			[
				[
					"h",
					"blue"
				]
			]
		]
	]
}`

func TestParseInlineBlock6(t *testing.T) {
	blocks := parseBlocks(t, title6)
	assert.Equal(t, 2, len(blocks))

	{
		b := blocks[0]
		assert.Equal(t, b.Text, "colored")
		assert.Equal(t, b.Highlight, "teal_background")
	}
	{
		b := blocks[1]
		assert.Equal(t, b.Text, "text")
		assert.Equal(t, b.Highlight, "blue")
	}
}

const title7 = `{
	"title": [
	  [
		"You can log in at: "
	  ],
	  [
		"http",
		[
		  [
			"a",
			"https://www.google.com/a/blendle.com"
		  ]
		]
	  ],
	  [
		"s",
		[
		  [
			"a"
		  ]
		]
	  ],
	  [
		"://www.google.com/a/blendle.com",
		[
		  [
			"a",
			"https://www.google.com/a/blendle.com"
		  ]
		]
	  ]
	]
  }`

func TestParseInlineBlock7(t *testing.T) {
	blocks := parseBlocks(t, title7)
	assert.Equal(t, 4, len(blocks))
}
