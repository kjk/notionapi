package main

import (
	"fmt"
	"os/exec"
	"runtime"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
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
