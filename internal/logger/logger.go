package logger

import (
	"log"
	"os"
)

var (
	// DebugEnabled controls whether debug logs are printed.
	DebugEnabled bool
)

// Debugf prints a formatted log message only if DebugEnabled is true.
func Debugf(format string, v ...interface{}) {
	if DebugEnabled {
		log.Printf(format, v...)
	}
}

// Info prints a log message universally (for crucial status info).
func Info(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func init() {
	// Use standard log setup but we'll control verbosity via Debugf
	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(os.Stdout)
}
