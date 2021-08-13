package main

import (
	"os"

	"github.com/kjk/notionapi"
)

var (
	didPrintTokenStatus bool
)

func makeNotionClient() *notionapi.Client {
	client := &notionapi.Client{
		AuthToken: getToken(),
		DebugLog:  flgVerbose,
		Logger:    os.Stdout,
	}

	if !didPrintTokenStatus {
		didPrintTokenStatus = true
		if client.AuthToken == "" {
			logf("NOTION_TOKEN env variable not set. Can only access public pages\n")
		} else {
			// TODO: validate that the token looks legit
			logf("NOTION_TOKEN env variable set, can access private pages\n")
		}
	}
	return client
}

var (
	eventsPerID = map[string]string{}
)

func downloadPage(client *notionapi.Client, pageID string) (*notionapi.Page, error) {
	d, err := notionapi.NewCachingClient(cacheDir, client)
	if err != nil {
		return nil, err
	}
	d.Policy = notionapi.PolicyDownloadNewer
	if flgNoCache {
		d.Policy = notionapi.PolicyDownloadAlways
	}
	return d.DownloadPage(pageID)
}

const (
	idNoDashLength = 32
)

// only hex chars seem to be valid
func isValidNoDashIDChar(c byte) bool {
	switch {
	case c >= '0' && c <= '9':
		return true
	case c >= 'a' && c <= 'f':
		return true
	case c >= 'A' && c <= 'F':
		// currently not used but just in case notion starts using them
		return true
	}
	return false
}

// given e.g.:
// /p/foo-395f6c6af50d44e48919a45fcc064d3e
// returns:
// 395f6c6af50d44e48919a45fcc064d3e
func extractNotionIDFromURL(uri string) string {
	n := len(uri)
	if n < idNoDashLength {
		return ""
	}

	s := ""
	for i := n - 1; i > 0; i-- {
		c := uri[i]
		if c == '-' {
			continue
		}
		if isValidNoDashIDChar(c) {
			s = string(c) + s
			if len(s) == idNoDashLength {
				return s
			}
		}
	}
	return ""
}
