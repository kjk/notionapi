package main

import (
	"fmt"
	"os"
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

func printCurrDir() {
	dir, err := os.Getwd()
	must(err)
	fmt.Printf("curr dir: %s\n", dir)
}

func doLineCount() int {
	fmt.Printf("doWordCount\n")
	err := os.Chdir(topDir())
	must(err)
	printCurrDir()
	stats := wc.NewLineStats()
	err = stats.CalcInDir(".", srcFiles, true)
	if err != nil {
		fmt.Printf("doWordCount: stats.wcInDir() failed with '%s'\n", err)
		return 1
	}
	wc.PrintLineStats(stats)
	return 0
}
