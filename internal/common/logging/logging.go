package logging

import (
	"fmt"
	"log"
)

var global = log.Default()
var colorEnabled = true

// SetLogger sets the global logger used across the application. Nil input leaves current logger unchanged.
func SetLogger(l *log.Logger) {
	if l == nil {
		return
	}
	global = l
}

// SetColorEnabled toggles ANSI coloring for Error logs.
func SetColorEnabled(enabled bool) {
	colorEnabled = enabled
}

// Info mirrors Printf for informational logs.
func Info(format string, args ...any) {
	global.Printf(format, args...)
}

// Error mirrors Printf but renders the message in red for visibility.
func Error(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if colorEnabled {
		global.Printf("%s%s%s", ansiRed, msg, ansiReset)
		return
	}
	global.Printf("%s", msg)
}

const (
	ansiRed   = "\033[31m"
	ansiReset = "\033[0m"
)

// Keep color constants for LogError; only explicit LogError is colored.
