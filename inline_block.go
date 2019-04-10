package notionapi

import (
	"encoding/json"
	"fmt"
)

const (
	// InlineAt is what Notion uses for text to represent @user and @date blocks
	InlineAt = "‣"
)

// AttrFlag is a compact description of some flags
type AttrFlag int

const (
	// AttrBold represents bold block
	AttrBold AttrFlag = 1 << iota
	// AttrCode represents code block
	AttrCode
	// AttrItalic represents italic block
	AttrItalic
	// AttrStrikeThrought represents strikethrough block
	AttrStrikeThrought
	// AttrComment represents a comment block
	AttrComment
)

// InlineBlock describes a nested inline block
// It's either Content or Type and Children
type InlineBlock struct {
	Text string `json:"Text"`
	// compact representation of attribute flags
	AttrFlags AttrFlag `json:"AttrFlags,omitempty"`

	// those values are set for other attributes

	// represents link attribute
	Link string `json:"Link,omitempty"`
	// represents user attribute
	UserID string `json:"UserID,omitempty"`
	// represents comment block (I think)
	CommentID string `json:"CommentID,omitempty"`
	// represents date attribute
	Date *Date `json:"Date,omitempty"`
	// represents highlight (text or background) color.
	// text color looks like: "blue"
	// bg color looks like: "teal_background"
	Highlight string `json:"Highlight,omitempty"`
}

// IsPlain returns true if this InlineBlock is plain text i.e. has no attributes
func (b *InlineBlock) IsPlain() bool {
	return b.AttrFlags == 0 && b.Link == "" && b.UserID == "" && b.Date == nil
}

func parseAttribute(b *InlineBlock, a []interface{}) error {
	if len(a) == 0 {
		return fmt.Errorf("attribute array is empty")
	}
	s, ok := a[0].(string)
	if !ok {
		return fmt.Errorf("a[0] is not string. a[0] is of type %T and value %#v", a[0], a)
	}

	if len(a) == 1 {
		switch s {
		case "b":
			b.AttrFlags |= AttrBold
		case "i":
			b.AttrFlags |= AttrItalic
		case "s":
			b.AttrFlags |= AttrStrikeThrought
		case "c":
			b.AttrFlags |= AttrCode
		default:
			return fmt.Errorf("unexpected attribute '%s'", s)
		}
		return nil
	}

	if len(a) != 2 {
		return fmt.Errorf("len(a) is %d and should be 2", len(a))
	}

	switch s {
	case "a", "u", "m", "h":
		v, ok := a[1].(string)
		if !ok {
			return fmt.Errorf("value for 'a' attribute is not string. Type: %T, value: %#v", v, v)
		}
		if s == "a" {
			b.Link = v
		} else if s == "u" {
			b.UserID = v
		} else if s == "m" {
			b.CommentID = v
		} else if s == "h" {
			b.Highlight = v
		} else {
			panic(fmt.Errorf("unexpected val '%s'", s))
		}
	case "d":
		v := a[1].(map[string]interface{})
		js, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			panic(err.Error())
		}
		//dbg("date in js:\n%s\n", string(js))
		var d *Date
		err = json.Unmarshal(js, &d)
		if err != nil {
			panic(err.Error())
		}
		b.Date = d
	default:
		return fmt.Errorf("unexpected attribute '%s'", s)
	}
	return nil
}

func parseAttributes(b *InlineBlock, a []interface{}) error {
	for _, rawAttr := range a {
		attrList, ok := rawAttr.([]interface{})
		if !ok {
			return fmt.Errorf("rawAttr is not []interface{} but %T of value %#v", rawAttr, rawAttr)
		}
		err := parseAttribute(b, attrList)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseInlineBlock(a []interface{}) (*InlineBlock, error) {
	if len(a) == 0 {
		return nil, fmt.Errorf("a is empty")
	}

	if len(a) == 1 {
		s, ok := a[0].(string)
		if !ok {
			return nil, fmt.Errorf("a is of length 1 but not string. a[0] el type: %T, el value: '%#v'", a[0], a[0])
		}
		return &InlineBlock{
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
	res := &InlineBlock{
		Text: s,
	}
	a, ok = a[1].([]interface{})
	if !ok {
		return nil, fmt.Errorf("a[1] is not []interface{}. a[1] type: %T, value: '%#v'", a[1], a[1])
	}
	err := parseAttributes(res, a)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// ParseInlineBlocks parses content from JSON into an easier to use form
func ParseInlineBlocks(raw interface{}) ([]*InlineBlock, error) {
	if raw == nil {
		return nil, nil
	}
	var res []*InlineBlock
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
		block, err := parseInlineBlock(a2)
		if err != nil {
			return nil, err
		}
		res = append(res, block)
	}
	return res, nil
}
