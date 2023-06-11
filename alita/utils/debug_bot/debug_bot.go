package debug_bot

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// PrettyPrintStruct prints the struct in a pretty format.
func PrettyPrintStruct(v interface{}) string {
	prettyStruct, _ := json.MarshalIndent(v, "", "  ")
	jsonStruct := string(prettyStruct)
	fmt.Printf("%s\n\n", jsonStruct)
	return jsonStruct
}

// PrintInLog prints the struct in the log.
func PrintInLog(v interface{}) {
	log.Debugf("%+v\n\n", v)
}
