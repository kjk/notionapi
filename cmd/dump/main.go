package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kjk/notionapi"
)

var (
	flgJSON bool
)

func parseFlags() []string {
	flag.BoolVar(&flgJSON, "json", false, "if true, prints in JSON format")
	flag.Parse()
	return flag.Args()
}

func usageAndExit() {
	fmt.Printf("Usage: %s <notion_page_id>\n", filepath.Base(os.Args[0]))
	os.Exit(1)
}

func main() {
	ids := parseFlags()
	if true {
		ids = []string{"4c6a54c68b3e4ea2af9cfaabcc88d58d"}
	}
	if len(ids) == 0 {
		usageAndExit()
	}

	for _, pageID := range ids {
		client := &notionapi.Client{}
		page, err := client.DownloadPage(pageID)
		if err != nil {
			log.Fatalf("client.DownloadPage(%s) failed with '%s'\n", pageID, err)
		}
		if flgJSON {
			d, err := json.MarshalIndent(page, "", "  ")
			if err != nil {
				log.Fatalf("json.Marshal() failed with %s\n", err)
			}
			fmt.Println(string(d))
		} else {
			notionapi.Dump(os.Stdout, page)
		}
	}
}
