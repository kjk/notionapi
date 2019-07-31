package main

import (
	"archive/zip"
	"io/ioutil"
	"os"
	"os/exec"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
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

func openNotepadWithFile(path string) {
	cmd := exec.Command("notepad.exe", path)
	err := cmd.Start()
	must(err)
}
