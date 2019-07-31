package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	flgTestToMd bool
)

const (
	logDir   = "log"
	cacheDir = "cache"
)

var (
	useCache = true
)

func parseFlags() {
	flag.BoolVar(&flgTestToMd, "test-to-md", false, "test markdown generation")
	flag.Parse()
}
func topDir() string {
	// we start inside "do" directory so topDir is
	// one dir above
	dir, err := filepath.Abs(".")
	must(err)
	return dir
}

func cdToTopDir() {
	err := os.Chdir("..")
	must(err)
}

func main() {
	cdToTopDir()
	fmt.Printf("topDir: '%s'\n", topDir())
	must(os.MkdirAll(logDir, 0755))
	must(os.MkdirAll(cacheDir, 0755))

	parseFlags()
	if true || flgTestToMd {
		os.Exit(testToMarkdown())
	}
	flag.Usage()
}
