package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/kjk/wc"
)

func notDataDir(name string) bool {
	name = filepath.ToSlash(name)
	return !strings.Contains(name, "data/")
}

var srcFiles = wc.MakeAllowedFileFilterForExts(".go")
var allFiles = wc.MakeFilterAnd(srcFiles, notDataDir)

func doLineCount() int {
	stats := wc.NewLineStats()
	err := stats.CalcInDir(".", srcFiles, true)
	if err != nil {
		fmt.Printf("doLineCount: stats.wcInDir() failed with '%s'\n", err)
		return 1
	}
	wc.PrintLineStats(stats)
	return 0
}
