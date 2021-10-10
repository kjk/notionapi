package notionapi

import (
	"encoding/json"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/pretty"
)

var jsonit = jsoniter.ConfigCompatibleWithStandardLibrary

var prettyOpts = pretty.Options{
	Width:  80,
	Prefix: "",
	Indent: "  ",
	// sorting keys only slightly slower
	SortKeys: true,
}

// TODO: doesn't work with some of Notion json responses?
// pretty-print if valid JSON. If not, return unchanged
// about 4x faster than naive version using json.Unmarshal() + json.Marshal()
func PrettyPrintJSBroken(js []byte) []byte {
	if !jsonit.Valid(js) {
		return js
	}
	return pretty.PrettyOptions(js, &prettyOpts)
}

func jsonUnmarshalFromMap(m map[string]interface{}, v interface{}) error {
	d, err := jsonit.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(d, v)
}

func jsonGetMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return nil
}

// pretty-print if valid JSON. If not, return unchanged
// about 4x faster than naive version using json.Unmarshal() + json.Marshal()
func PrettyPrintJS(js []byte) []byte {
	var m map[string]interface{}
	err := json.Unmarshal(js, &m)
	if err != nil {
		return js
	}
	d, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return js
	}
	return d
}
