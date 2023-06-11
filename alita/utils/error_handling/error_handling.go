package error_handling

import log "github.com/sirupsen/logrus"

// FatalError logs an error and exits the program.
func FatalError(funcName, modName string, err error) {
	if err != nil {
		log.Errorf("[%s][%s] %v", modName, funcName, err)
		return
	}
}

// HandleErr handles errors by logging them.
func HandleErr(err error) {
	if err != nil {
		log.Error(err)
		return
	}
}
