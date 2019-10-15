package main

import (
	"fmt"

	"github.com/kjk/u"
)

var srcFiles = u.MakeAllowedFileFilterForExts(".go", ".js", ".html", ".css")
var excludeDirs = u.MakeExcludeDirsFilter("node_modules", "tmpdata")
var allFiles = u.MakeFilterAnd(srcFiles, excludeDirs)

func doLineCount() int {
	stats := u.NewLineStats()
	err := stats.CalcInDir(".", allFiles, true)
	if err != nil {
		fmt.Printf("doLineCount: stats.wcInDir() failed with '%s'\n", err)
		return 1
	}
	u.PrintLineStats(stats)
	return 0
}
