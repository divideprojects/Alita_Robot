package debug_bot

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// PrettyPrintStruct formats and prints a struct as indented JSON to stdout.
// Returns the formatted JSON string for further use or logging.
func PrettyPrintStruct(v any) string {
	prettyStruct, _ := json.MarshalIndent(v, "", "  ")
	jsonStruct := string(prettyStruct)
	fmt.Printf("%s\n\n", jsonStruct)
	return jsonStruct
}

// PrintInLog outputs a struct's detailed representation to the debug log.
// Uses logrus debug level for development debugging and troubleshooting.
func PrintInLog(v any) {
	log.Debugf("%+v\n\n", v)
}
