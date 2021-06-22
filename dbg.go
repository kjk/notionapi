package notionapi

import (
	"fmt"
)

// vlogf is for verbose logging
func (c *Client) vlogf(format string, args ...interface{}) {
	if !c.DebugLog {
		return
	}
	c.logf(format, args...)
}

// TODO: replace with Client.vlogf()
func dbg(client *Client, format string, args ...interface{}) {
	client.vlogf(format, args...)
}

func (c *Client) logf(format string, args ...interface{}) {
	if c.Logger == nil {
		return
	}
	if len(args) == 0 {
		fmt.Fprint(c.Logger, format)
		return
	}
	fmt.Fprintf(c.Logger, format, args...)
}

// TODO: replace with Client.logf()
func log(client *Client, format string, args ...interface{}) {
	client.logf(format, args...)
}

// TODO: add option to enable this logging
// log JSON after pretty printing it
func logJSON(client *Client, js []byte) {
	//log(client, "%s\n", string(PrettyPrintJS(js)))
}
