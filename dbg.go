package notionapi

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/pretty"
)

func dbg(client *Client, format string, args ...interface{}) {
	if !client.DebugLog {
		return
	}
	log(client, format, args...)
}

func log(client *Client, format string, args ...interface{}) {
	if client.Logger == nil {
		return
	}
	fmt.Fprintf(client.Logger, format, args...)
}

var prettyOpts = pretty.Options{
	Width:  80,
	Prefix: "",
	Indent: "  ",
	// sorting keys only slightly slower
	SortKeys: true,
}

// pretty-print if valid JSON. If not, return unchanged
// about 4x faster than naive version using json.Unmarshal() + json.Marshal()
func ppJSON(js []byte) []byte {
	if !json.Valid(js) {
		return js
	}
	return pretty.PrettyOptions(js, &prettyOpts)
}

// log JSON after pretty printing it
func logJSON(client *Client, js []byte) {
	//log(client, "%s\n", string(ppJSON(js)))
}
