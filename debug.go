package notionapi

import "fmt"

var (
	// PanicOnFailures will force panics on unexpected situations.
	// This is for debugging
	PanicOnFailures bool

	// LogFunc allows intercepting debug logs
	LogFunc func(format string, args ...interface{})
)

// Log is for debug logging, will log using LogFunc (if set)
func Log(format string, args ...interface{}) {
	if LogFunc != nil {
		LogFunc(format, args...)
	}
}

// MaybePanic will panic if PanicOnFailures is true
func MaybePanic(format string, args ...interface{}) {
	if LogFunc != nil {
		LogFunc(format, args...)
	}
	if PanicOnFailures {
		panic(fmt.Sprintf(format, args...))
	}
}
