package debug_bot

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

//goland:noinspection GoUnusedExportedFunction
func PrettyPrintStruct(v interface{}) string {
	prettyStruct, _ := json.MarshalIndent(v, "", "  ")
	jsonStruct := string(prettyStruct)
	fmt.Printf("%s\n\n", jsonStruct)
	return jsonStruct
}

//goland:noinspection GoUnusedExportedFunction
func PrintInLog(v interface{}) {
	log.Debugf("%+v\n\n", v)
}
