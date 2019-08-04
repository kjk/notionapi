package notionapi

import (
	"encoding/json"
	"fmt"
)

const (
	// TextSpanSpecial is what Notion uses for text to represent @user and @date blocks
	TextSpanSpecial = "‣"
)

const (
	// AttrBold represents bold block
	AttrBold = "b"
	// AttrCode represents code block
	AttrCode = "c"
	// AttrItalic represents italic block
	AttrItalic = "i"
	// AttrStrikeThrought represents strikethrough block
	AttrStrikeThrought = "s"
	// AttrComment represents a comment block
	AttrComment = "m"
	// AttrLink represnts a link (url)
	AttrLink = "a"
	// AttrUser represents an id of a user
	AttrUser = "u"
	// AttrHighlight represents text high-light
	AttrHighlight = "h"
	// AttrDate represents a date
	AttrDate = "d"
	// AtttrPage represents a link to a Notion page
	AttrPage = "p"
)

// TextAttr describes attributes of a span of text
// First element is name of the attribute (e.g. AttrLink)
// The rest are optional information about attribute (e.g.
// for AttrLink it's URL, for AttrUser it's user id etc.)
type TextAttr = []string

// TextSpan describes a text with attributes
type TextSpan struct {
	Text  string     `json:"Text"`
	Attrs []TextAttr `json:"Attrs"`
}

// IsPlain returns true if this InlineBlock is plain text i.e. has no attributes
func (t *TextSpan) IsPlain() bool {
	return len(t.Attrs) == 0
}

func AttrGetType(attr TextAttr) string {
	return attr[0]
}

func panicIfAttrNot(attr TextAttr, fnName string, expectedType string) {
	if AttrGetType(attr) != expectedType {
		panic(fmt.Sprintf("don't call %s on attribute of type %s", fnName, AttrGetType(attr)))
	}
}

func AttrGetLink(attr TextAttr) string {
	panicIfAttrNot(attr, "AttrGetLink", AttrLink)
	// there are links with
	if len(attr) == 1 {
		return ""
	}
	return attr[1]
}

func AttrGetUserID(attr TextAttr) string {
	panicIfAttrNot(attr, "AttrGetUserID", AttrUser)
	return attr[1]
}

func AttrGetPageID(attr TextAttr) string {
	panicIfAttrNot(attr, "AttrGetPageID", AttrPage)
	return attr[1]
}

func AttrGetComment(attr TextAttr) string {
	panicIfAttrNot(attr, "AttrGetComment", AttrComment)
	return attr[1]
}

func AttrGetHighlight(attr TextAttr) string {
	panicIfAttrNot(attr, "AttrGetHighlight", AttrHighlight)
	return attr[1]
}

func AttrGetDate(attr TextAttr) *Date {
	panicIfAttrNot(attr, "AttrGetDate", AttrDate)
	js := []byte(attr[1])
	var d *Date
	err := json.Unmarshal(js, &d)
	if err != nil {
		panic(err.Error())
	}
	return d
}

func parseTextSpanAttribute(b *TextSpan, a []interface{}) error {
	if len(a) == 0 {
		return fmt.Errorf("attribute array is empty")
	}
	s, ok := a[0].(string)
	if !ok {
		return fmt.Errorf("a[0] is not string. a[0] is of type %T and value %#v", a[0], a)
	}
	attr := TextAttr{s}
	if s == AttrDate {
		// date is a special case in that second value is
		if len(a) != 2 {
			return fmt.Errorf("unexpected date attribute. Expected 2 values, got: %#v", a)
		}
		v, ok := a[1].(map[string]interface{})
		if !ok {
			return fmt.Errorf("got unexpected type %T (expected map[string]interface{}", a[1])
		}
		js, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		attr = append(attr, string(js))
		b.Attrs = append(b.Attrs, attr)
		return nil
	}
	for _, v := range a[1:] {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("got unexpected type %T (expected string)", v)
		}
		attr = append(attr, s)
	}
	b.Attrs = append(b.Attrs, attr)
	return nil
}

func parseTextSpanAttributes(b *TextSpan, a []interface{}) error {
	for _, rawAttr := range a {
		attrList, ok := rawAttr.([]interface{})
		if !ok {
			return fmt.Errorf("rawAttr is not []interface{} but %T of value %#v", rawAttr, rawAttr)
		}
		err := parseTextSpanAttribute(b, attrList)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseTextSpan(a []interface{}) (*TextSpan, error) {
	if len(a) == 0 {
		return nil, fmt.Errorf("a is empty")
	}

	if len(a) == 1 {
		s, ok := a[0].(string)
		if !ok {
			return nil, fmt.Errorf("a is of length 1 but not string. a[0] el type: %T, el value: '%#v'", a[0], a[0])
		}
		return &TextSpan{
			Text: s,
		}, nil
	}
	if len(a) != 2 {
		return nil, fmt.Errorf("a is of length != 2. a value: '%#v'", a)
	}

	s, ok := a[0].(string)
	if !ok {
		return nil, fmt.Errorf("a[0] is not string. a[0] type: %T, value: '%#v'", a[0], a[0])
	}
	res := &TextSpan{
		Text: s,
	}
	a, ok = a[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("a[1] is not []interface{}. a[1] type: %T, value: '%#v'", a[1], a[1])
	}
	err := parseTextSpanAttributes(res, a)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ParseTextSpans parses content from JSON into an easier to use form
func ParseTextSpans(raw interface{}) ([]*TextSpan, error) {
	if raw == nil {
		return nil, nil
	}
	var res []*TextSpan
	a, ok := raw.([]interface{})
	if !ok {
		return nil, fmt.Errorf("raw is not of []interface{}. raw type: %T, value: '%#v'", raw, raw)
	}
	if len(a) == 0 {
		return nil, fmt.Errorf("raw is empty")
	}
	for _, v := range a {
		a2, ok := v.([]interface{})
		if !ok {
			return nil, fmt.Errorf("v is not []interface{}. v type: %T, value: '%#v'", v, v)
		}
		span, err := parseTextSpan(a2)
		if err != nil {
			return nil, err
		}
		res = append(res, span)
	}
	return res, nil
}

// TextSpansToString returns flattened content of inline blocks, without formatting
func TextSpansToString(blocks []*TextSpan) string {
	s := ""
	for _, block := range blocks {
		if block.Text == TextSpanSpecial {
			// TODO: how to handle dates, users etc.?
			continue
		}
		s += block.Text
	}
	return s
}

func getFirstInline(inline []*TextSpan) string {
	if len(inline) == 0 {
		return ""
	}
	return inline[0].Text
}

func getFirstInlineBlock(v interface{}) (string, error) {
	inline, err := ParseTextSpans(v)
	if err != nil {
		return "", err
	}
	return getFirstInline(inline), nil
}

func getInlineText(v interface{}) (string, error) {
	inline, err := ParseTextSpans(v)
	if err != nil {
		return "", err
	}
	return TextSpansToString(inline), nil
}
