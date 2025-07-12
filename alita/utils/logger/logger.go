package logger

import (
	"fmt"
	"path"
	"runtime"

	log "github.com/sirupsen/logrus"
)

// Setup configures the global logger with the specified debug mode.
// This should be called early in main() before other packages are imported.
func Setup(debug bool) {
	// Set log level based on debug mode
	if debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.SetReportCaller(true)
	log.SetFormatter(
		&log.JSONFormatter{
			DisableHTMLEscape: true,
			PrettyPrint:       true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				return f.Function, fmt.Sprintf("%s:%d", path.Base(f.File), f.Line)
			},
		},
	)
}
