package modules

import (
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// autoModuleName returns the title-cased base filename of the caller (without extension).
// It can be used during variable initialisation so that each module automatically
// picks up its name from the file it is defined in, avoiding hard-coding.
func autoModuleName() string {
	// runtime.Caller(1) gives us the file path of the caller of this function –
	// i.e. the *.go file where autoModuleName() is invoked.
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}

	base := filepath.Base(file)                          // e.g. "admin.go"
	name := strings.TrimSuffix(base, filepath.Ext(base)) // -> "admin"

	// Capitalise for consistency with existing translation keys, eg "Admin".
	return cases.Title(language.English).String(name)
}
