package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/kjk/u"
)

func must(err error) {
	u.Must(err)
}

// a centralized place allows us to tweak logging, if need be
func logf(format string, args ...interface{}) {
	u.Logf(format, args...)
}

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
