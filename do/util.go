package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
			fmt.Fprint(logFile, format)
		}
		return
	}
	fmt.Printf(format, args...)
	if logFile != nil {
		fmt.Fprintf(logFile, format, args...)
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

func readZipFile(path string) map[string][]byte {
	r, err := zip.OpenReader(path)
	must(err)
	defer r.Close()
	res := map[string][]byte{}
	for _, f := range r.File {
		rc, err := f.Open()
		must(err)
		d, err := ioutil.ReadAll(rc)
		must(err)
		rc.Close()
		res[f.Name] = d
	}
	return res
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

func dirExists(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return st.Mode().IsDir()
}

func recreateDir(dir string) {
	os.RemoveAll(dir)
	err := os.MkdirAll(dir, 0755)
	must(err)
}

func removeFilesInDir(dir string) {
	err := os.MkdirAll(dir, 0755)
	must(err)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return
	}
	for _, fi := range files {
		if !fi.Mode().IsRegular() {
			continue
		}
		path := filepath.Join(dir, fi.Name())
		err = os.Remove(path)
		must(err)
	}
}

func writeFile(path string, data []byte) {
	err := ioutil.WriteFile(path, data, 0644)
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

func readFile(path string) []byte {
	d, err := ioutil.ReadFile(path)
	must(err)
	return d
}

func areFilesEuqal(path1, path2 string) bool {
	d1 := readFile(path1)
	d2 := readFile(path2)
	return bytes.Equal(d1, d2)
}

func openNotepadWithFile(path string) {
	cmd := exec.Command("notepad.exe", path)
	err := cmd.Start()
	must(err)
}
