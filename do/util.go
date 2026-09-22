package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kjk/common/u"
)

var (
	must        = u.Must
	panicIf     = u.PanicIf
	openBrowser = u.OpenBrowser
)

func recreateDir(dir string) {
	_ = os.RemoveAll(dir)
	err := os.MkdirAll(dir, 0755)
	must(err)
}

func openNotepadWithFile(path string) {
	cmd := exec.Command("notepad.exe", path)
	err := cmd.Start()
	must(err)
}

func openCodeDiff(path1, path2 string) {
	if runtime.GOOS == "darwin" {
		path1 = strings.Replace(path1, ".\\", "./", -1)
		path2 = strings.Replace(path2, ".\\", "./", -1)
	}
	cmd := exec.Command("code", "--new-window", "--diff", path1, path2)
	logf("running: %s\n", strings.Join(cmd.Args, " "))
	err := cmd.Start()
	must(err)
}

func writeFileMust(path string, data []byte) {
	err := ioutil.WriteFile(path, data, 0644)
	must(err)
}

// if set, logf() also writes to this file
var logFile io.Writer

func logf(format string, args ...interface{}) {
	if len(args) == 0 {
		fmt.Print(format)
		if logFile != nil {
			_, _ = fmt.Fprint(logFile, format)
		}
		return
	}
	fmt.Printf(format, args...)
	if logFile != nil {
		_, _ = fmt.Fprintf(logFile, format, args...)
	}
}

func readFileMust(path string) []byte {
	d, err := os.ReadFile(path)
	must(err)
	return d
}

func removeFilesInDirMust(dir string) {
	if !u.DirExists(dir) {
		return
	}
	entries, err := os.ReadDir(dir)
	must(err)
	for _, e := range entries {
		if !e.Type().IsRegular() {
			continue
		}
		err = os.Remove(filepath.Join(dir, e.Name()))
		must(err)
	}
}

func currDirAbsMust() string {
	dir, err := filepath.Abs(".")
	must(err)
	return dir
}

// cdUpDir changes current directory to the closest parent
// (or current) directory named dirName
func cdUpDir(dirName string) {
	startDir := currDirAbsMust()
	dir := startDir
	for {
		if filepath.Base(dir) == dirName && u.DirExists(dir) {
			must(os.Chdir(dir))
			return
		}
		parentDir := filepath.Dir(dir)
		panicIf(dir == parentDir, "invalid startDir: '%s', dir: '%s'", startDir, dir)
		dir = parentDir
	}
}

func runCmdMust(cmd *exec.Cmd) string {
	fmt.Printf("> %s\n", strings.Join(cmd.Args, " "))
	canCapture := (cmd.Stdout == nil) && (cmd.Stderr == nil)
	if canCapture {
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("cmd '%s' failed with '%s'. Output:\n%s\n", cmd, err, string(out))
			must(err)
		} else if len(out) > 0 {
			fmt.Printf("Output:\n%s\n", string(out))
		}
		return string(out)
	}
	must(cmd.Run())
	return ""
}
