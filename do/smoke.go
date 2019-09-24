package main

import (
	"os"
	"path/filepath"
)

// smoke test is meant to be run after non-trivial changes
// it tries to exercise as many features as possible while still
// being reasonably fast
func smokeTest() {
	dir := filepath.Join("data", "smoke")
	recreateDir(dir)
	logFilePath := filepath.Join(dir, "log.txt")
	logf("Running smokeTest(), log file: '%s'\n", logFilePath)
	f, err := os.Create(logFilePath)
	must(err)
	defer f.Close()
	logFile = f
}
