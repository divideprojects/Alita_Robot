package debug_bot

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

/*
PrettyPrintStruct returns a pretty-printed JSON string representation of the given struct.

It also prints the formatted JSON to standard output for debugging purposes.
*/
func PrettyPrintStruct(v interface{}) string {
	prettyStruct, _ := json.MarshalIndent(v, "", "  ")
	jsonStruct := string(prettyStruct)
	fmt.Printf("%s\n\n", jsonStruct)
	return jsonStruct
}

/*
PrintInLog logs the given struct or value using the debug log level.

The output includes field names and values for easier inspection during debugging.
*/
func PrintInLog(v interface{}) {
	log.Debugf("%+v\n\n", v)
}
