package notionapi

import (
	"encoding/json"
	"fmt"
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

// pretty-print if valid JSON. If not, return unchanged
func ppJSON(js []byte) []byte {
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

// log JSON after pretty printing it
func logJSON(client *Client, js []byte) {
	//log(client, "%s\n", string(ppJSON(js)))
}
