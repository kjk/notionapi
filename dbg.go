package notionapi

import (
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

// log JSON after pretty printing it
func logJSON(client *Client, js []byte) {
	//log(client, "%s\n", string(PrettyPrintJS(js)))
}
