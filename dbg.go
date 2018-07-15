package notionapi

import (
	"encoding/json"
	"fmt"
	"io"
)

var (
	// Logger is used to log requests and responses for debugging.
	// By default is not set.
	Logger io.Writer
	// DebugLog enables debug logging
	DebugLog = false
)

func dbg(format string, args ...interface{}) {
	if !DebugLog {
		return
	}
	fmt.Printf(format, args...)
}

func log(format string, args ...interface{}) {
	if Logger == nil {
		return
	}
	fmt.Fprintf(Logger, format, args...)
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
func logJSON(js []byte) {
	if Logger == nil {
		return
	}
	js = ppJSON(js)
	fmt.Fprintf(Logger, "%s\n", string(js))
}
