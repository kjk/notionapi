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

func dirExists(path string) bool {
	st, err := os.Stat(path)
	if err != nil {
		return false
	}
	return st.Mode().IsDir()
}

func recreateDir(dir string) {
	_ = os.RemoveAll(dir)
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

func ls(path string) {
	fi, err := os.Stat(path)
	must(err)
	fmt.Printf("%s %d\n", path, fi.Size())
}

func cpFile(dstPath, srcPath string) {
	d, err := ioutil.ReadFile(srcPath)
	must(err)
	err = ioutil.WriteFile(dstPath, d, 0666)
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

func readFile(path string) []byte {
	d, err := ioutil.ReadFile(path)
	must(err)
	return d
}

func fileClose(f io.Closer) {
	_ = f.Close()
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

func runCmd(cmd *exec.Cmd) string {
	fmt.Printf("> %s\n", cmd)
	canCapture := (cmd.Stdout == nil) && (cmd.Stderr == nil)
	if canCapture {
		out, err := cmd.CombinedOutput()
		if err == nil {
			if len(out) > 0 {
				fmt.Printf("Output:\n%s\n", string(out))
			}
			return string(out)
		}
		fmt.Printf("cmd '%s' failed with '%s'. Output:\n%s\n", cmd, err, string(out))
		must(err)
		return string(out)
	}
	err := cmd.Run()
	if err == nil {
		return ""
	}
	fmt.Printf("cmd '%s' failed with '%s'\n", cmd, err)
	must(err)
	return ""
}

func gitPull(dir string) {
	cmd := exec.Command("git", "pull")
	if dir != "" {
		cmd.Dir = dir
	}
	runCmd(cmd)
}

func gitStatus(dir string) string {
	cmd := exec.Command("git", "status")
	if dir != "" {
		cmd.Dir = dir
	}
	return runCmd(cmd)
}

func checkGitClean(dir string) {
	s := gitStatus(dir)
	expected := []string{
		"On branch master",
		"Your branch is up to date with 'origin/master'.",
		"nothing to commit, working tree clean",
	}
	for _, exp := range expected {
		if !strings.Contains(s, exp) {
			fmt.Printf("Git repo in '%s' not clean.\nDidn't find '%s' in output of git status:\n%s\n", dir, exp, s)
			os.Exit(1)
		}
	}
}

func readZipFile(path string) map[string][]byte {
	r, err := zip.OpenReader(path)
	must(err)
	defer fileClose(r)
	res := map[string][]byte{}
	for _, f := range r.File {
		rc, err := f.Open()
		must(err)
		d, err := ioutil.ReadAll(rc)
		must(err)
		_ = rc.Close()
		res[f.Name] = d
	}
	return res
}

func zipAddFile(zw *zip.Writer, zipName string, path string) {
	zipName = filepath.ToSlash(zipName)
	d, err := ioutil.ReadFile(path)
	must(err)
	w, err := zw.Create(zipName)
	_, err = w.Write(d)
	must(err)
	fmt.Printf("  added %s from %s\n", zipName, path)
}

func zipDirRecur(zw *zip.Writer, baseDir string, dirToZip string) {
	dir := filepath.Join(baseDir, dirToZip)
	files, err := ioutil.ReadDir(dir)
	must(err)
	for _, fi := range files {
		if fi.IsDir() {
			zipDirRecur(zw, baseDir, filepath.Join(dirToZip, fi.Name()))
		} else if fi.Mode().IsRegular() {
			zipName := filepath.Join(dirToZip, fi.Name())
			path := filepath.Join(baseDir, zipName)
			zipAddFile(zw, zipName, path)
		} else {
			path := filepath.Join(baseDir, fi.Name())
			s := fmt.Sprintf("%s is not a dir or regular file", path)
			panic(s)
		}
	}
}

func createZipFile(dst string, baseDir string, toZip ...string) {
	removeFile(dst)
	if len(toZip) == 0 {
		panic("must provide toZip args")
	}
	fmt.Printf("Creating zip file %s\n", dst)
	w, err := os.Create(dst)
	must(err)
	defer fileClose(w)
	zw := zip.NewWriter(w)
	must(err)
	for _, name := range toZip {
		path := filepath.Join(baseDir, name)
		fi, err := os.Stat(path)
		must(err)
		if fi.IsDir() {
			zipDirRecur(zw, baseDir, name)
		} else if fi.Mode().IsRegular() {
			zipAddFile(zw, name, path)
		} else {
			s := fmt.Sprintf("%s is not a dir or regular file", path)
			panic(s)
		}
	}
	err = zw.Close()
	must(err)
}
