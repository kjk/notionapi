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

// absolute path of top directory in the repo
func topDir() string {
	dir, err := filepath.Abs(".")
	must(err)
	return dir
}

// we are executed for do/ directory so top dir is parent dir
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
