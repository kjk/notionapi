package notionapi

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseTextSpans(t *testing.T, s string) []*TextSpan {
	var m map[string]interface{}
	err := json.Unmarshal([]byte(s), &m)
	assert.NoError(t, err)
	blocks, err := ParseTextSpans(m["title"])
	assert.NoError(t, err)
	return blocks
}

func TestParseTextSpans1(t *testing.T) {
	spans := parseTextSpans(t, title1)
	assert.Equal(t, 1, len(spans))
	ts := spans[0]
	assert.Equal(t, "Test page text", ts.Text)
	assert.True(t, ts.IsPlain())
}

func TestParseTextSpans2(t *testing.T) {
	spans := parseTextSpans(t, title2)
	assert.Equal(t, 1, len(spans))
	ts := spans[0]
	assert.Equal(t, TextSpanSpecial, ts.Text)
	assert.Equal(t, 1, len(ts.Attrs))
	attr := ts.Attrs[0]
	assert.Equal(t, Attr2User, attr[0])
	assert.Equal(t, "bb760e2d-d679-4b64-b2a9-03005b21870a", attr[1])
}

func TestParseTextSpans3(t *testing.T) {
	blocks := parseTextSpans(t, title3)
	assert.Equal(t, 2, len(blocks))
	{
		b := blocks[0]
		assert.Equal(t, "Text block with ", b.Text)
		assert.Equal(t, 0, len(b.Attrs))
	}

	{
		b := blocks[1]
		assert.Equal(t, "bold ", b.Text)
		attr := b.Attrs[0]
		assert.Equal(t, Attr2Bold, attr[0])
	}
}

func TestParseTextSpans4(t *testing.T) {
	blocks := parseTextSpans(t, title4)
	assert.Equal(t, 1, len(blocks))
	{
		b := blocks[0]
		assert.Equal(t, "link inside bold", b.Text)
		assert.Equal(t, 2, len(b.Attrs))
		attr := b.Attrs[0]
		assert.Equal(t, Attr2Bold, AttrGetType(attr))
		attr = b.Attrs[1]
		assert.Equal(t, Attr2Link, AttrGetType(attr))
		assert.Equal(t, "https://www.google.com", AttrGetLink(attr))
	}
}

func TestParseTextSpans5(t *testing.T) {
	blocks := parseTextSpans(t, title5)
	assert.Equal(t, 1, len(blocks))
	b := blocks[0]
	assert.Equal(t, TextSpanSpecial, b.Text)
	assert.Equal(t, 1, len(b.Attrs))
	attr := b.Attrs[0]
	assert.Equal(t, Attr2Date, AttrGetType(attr))
	date := AttrGetDate(attr)
	assert.Equal(t, date.DateFormat, "relative")
	assert.Equal(t, date.StartDate, "2018-07-17")
	assert.Equal(t, date.Type, "datetime")
}

func TestParseTextSpansBig(t *testing.T) {
	blocks := parseTextSpans(t, titleBig)
	assert.Equal(t, 17, len(blocks))
}

func TestParseTextSpansComment(t *testing.T) {
	blocks := parseTextSpans(t, titleWithComment)
	assert.Equal(t, 3, len(blocks))

	{
		// "Just"
		b := blocks[0]
		assert.Equal(t, b.Text, "Just")
		assert.Equal(t, 0, len(b.Attrs))
	}
	{
		// "comment"
		b := blocks[1]
		assert.Equal(t, b.Text, "comment")
		attr := b.Attrs[0]
		assert.Equal(t, Attr2Comment, AttrGetType(attr))
		assert.Equal(t, "4a1cc3be-03cf-489a-9542-69d9a02f3534", AttrGetComment(attr))
	}
}

func TestParseTextSpans6(t *testing.T) {
	blocks := parseTextSpans(t, title6)
	assert.Equal(t, 2, len(blocks))

	{
		b := blocks[0]
		assert.Equal(t, b.Text, "colored")
		attr := b.Attrs[0]
		assert.Equal(t, Attr2Highlight, AttrGetType(attr))
		assert.Equal(t, "teal_background", AttrGetHighlight(attr))
	}
	{
		b := blocks[1]
		assert.Equal(t, b.Text, "text")
		attr := b.Attrs[0]
		assert.Equal(t, Attr2Highlight, AttrGetType(attr))
		assert.Equal(t, "blue", AttrGetHighlight(attr))
	}
}

func TestParseTextSpan7(t *testing.T) {
	blocks := parseTextSpans(t, title7)
	assert.Equal(t, 4, len(blocks))
}
