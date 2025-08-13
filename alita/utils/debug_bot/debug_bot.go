package debug_bot

import (
	"encoding/json"
	"fmt"
)

// PrettyPrintStruct formats and prints a struct as indented JSON to stdout.
// Returns the formatted JSON string for further use or logging.
func PrettyPrintStruct(v any) string {
	prettyStruct, _ := json.MarshalIndent(v, "", "  ")
	jsonStruct := string(prettyStruct)
	fmt.Printf("%s\n\n", jsonStruct)
	return jsonStruct
}
