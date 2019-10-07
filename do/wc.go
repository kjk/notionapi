package main

import (
	"fmt"
	"path/filepath"

	"github.com/kjk/wc"
)

func excludeDirs(path string) bool {
	path = filepath.Dir(path)
	for len(path) > 0 {
		if path == "." {
			return true
		}
		name := filepath.Base(path)
		if name == "node_modules" {
			return false
		}
		if name == "tmpdata" {
			return false
		}
		path = filepath.Dir(path)
	}
	return true
}

var srcFiles = wc.MakeAllowedFileFilterForExts(".go", ".js", ".html", ".css")
var allFiles = wc.MakeFilterAnd(srcFiles, excludeDirs)

func doLineCount() int {
	stats := wc.NewLineStats()
	err := stats.CalcInDir(".", allFiles, true)
	if err != nil {
		fmt.Printf("doLineCount: stats.wcInDir() failed with '%s'\n", err)
		return 1
	}
	wc.PrintLineStats(stats)
	return 0
}
