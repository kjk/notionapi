package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	logFile io.Writer
)

func must(err error) {
	if err != nil {
		fmt.Printf("err: %s\n", err)
		panic(err)
	}
}

func assert(ok bool, format string, args ...interface{}) {
	if ok {
		return
	}
	s := fmt.Sprintf(format, args...)
	panic(s)
}

func panicIf(cond bool, args ...interface{}) {
	if !cond {
		return
	}
	if len(args) == 0 {
		panic("condition failed")
	}
	format := args[0].(string)
	if len(args) == 1 {
		panic(format)
	}
	panic(fmt.Sprintf(format, args[1:]...))
}

// a centralized place allows us to tweak logging, if need be
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

// openBrowsers open web browser with a given url
// (can be http:// or file://)
func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	must(err)
}

func getHomeDir() string {
	s, err := os.UserHomeDir()
	must(err)
	return s
}

// absolute path of the current directory
func currDirAbs() string {
	dir, err := filepath.Abs(".")
	must(err)
	return dir
}

// we are executed for do/ directory so top dir is parent dir
func cdUpDir(dirName string) {
	startDir := currDirAbs()
	dir := startDir
	for {
		// we're already in top directory
		if filepath.Base(dir) == dirName {
			err := os.Chdir(dir)
			must(err)
			return
		}
		parentDir := filepath.Dir(dir)
		panicIf(dir == parentDir, "invalid startDir: '%s', dir: '%s'", startDir, dir)
		dir = parentDir
	}
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return st.Mode().IsRegular()
}

func recreateDir(dir string) {
	_ = os.RemoveAll(dir)
	err := os.MkdirAll(dir, 0755)
	must(err)
}

func removeFile(path string) {
	err := os.Remove(path)
	if err == nil {
		logf("removeFile('%s')", path)
		return
	}
	if os.IsNotExist(err) {
		// TODO: maybe should print note
		return
	}
	fmt.Printf("os.Remove('%s') failed with '%s'\n", path, err)
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
