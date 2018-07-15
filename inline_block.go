package notionapi

import (
	"encoding/json"
	"fmt"
)

const (
	// InlineAt is a text used for inline blocks representing
	// for @user and @date
	InlineAt = "‣"
)

// AttrFlag is a compact description of
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
)

// Attr represents one of many complex attributes of inline blocks
type Attr interface {
}

// AttrLink represents a link
type AttrLink struct {
	Link string
}

// AttrUser represents @user attribute
type AttrUser struct {
	UserID string
}

// AttrDate represents @date attribute
type AttrDate struct {
	Date *Date
}

// InlineBlock describes a nested inline block
// It's either Content or Type and Children
type InlineBlock struct {
	Text string
	// compact representation of attribute flags
	AttrFlags AttrFlag
	Attrs     []Attr
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
	case "a", "u":
		v, ok := a[1].(string)
		if !ok {
			return fmt.Errorf("value for 'a' attribute is not string. Type: %T, value: %#v", v, v)
		}
		var attr Attr
		if s == "a" {
			attr = &AttrLink{
				Link: v,
			}
		} else if s == "u" {
			attr = &AttrUser{
				UserID: v,
			}
		}
		b.Attrs = append(b.Attrs, attr)
	case "d":
		v := a[1].(map[string]interface{})
		js, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			panic(err.Error())
		}
		//dbg("date in js:\n%s\n", string(js))
		var d Date
		err = json.Unmarshal(js, &d)
		if err != nil {
			panic(err.Error())
		}
		attr := &AttrDate{
			Date: &d,
		}
		b.Attrs = append(b.Attrs, attr)
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

func parseInlineBlocks(raw interface{}) ([]*InlineBlock, error) {
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
