package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
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

// a centralized place allows us to tweak loggign, if need be
func log(format string, args ...interface{}) {
	fmt.Printf(format, args...)
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

func reacreateDir(dir string) {
	os.RemoveAll(dir)
	err := os.MkdirAll(dir, 0755)
	must(err)
}

func writeFile(path string, data []byte) {
	err := ioutil.WriteFile(path, data, 0666)
	must(err)
}

func openCodeDiff(path1, path2 string) {
	cmd := exec.Command("code", "--new-window", "--diff", path1, path2)
	err := cmd.Start()
	must(err)
}

func loadFileData(path string) []byte {
	d, err := ioutil.ReadFile(path)
	must(err)
	return d
}

func areFilesEuqal(path1, path2 string) bool {
	d1 := loadFileData(path1)
	d2 := loadFileData(path2)
	return bytes.Equal(d1, d2)
}

func openNotepadWithFile(path string) {
	cmd := exec.Command("notepad.exe", path)
	err := cmd.Start()
	must(err)
}

func formatHTMLFile(path string) {
	cmd := exec.Command("prettier", "--write", path)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("prettier failed with:\n%s\n", string(out))
	}
}

func ensurePrettierExists() {
	cmd := exec.Command("prettier", "-v")
	err := cmd.Run()
	if err != nil {
		log("prettier doesn't seem to be installed. Either run with -no-reformat or install prettier with:\n")
		log("npm install --global prettier\n")
		os.Exit(1)
	}
}
