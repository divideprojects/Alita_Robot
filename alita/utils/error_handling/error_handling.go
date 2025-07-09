package error_handling

import log "github.com/sirupsen/logrus"

/*
FatalError logs an error with function and module context.

If an error is provided, it logs the error message with the given function and module names.
Does not exit the program, only logs the error.
*/
func FatalError(funcName, modName string, err error) {
	if err != nil {
		log.Errorf("[%s][%s] %v", modName, funcName, err)
		return
	}
}

/*
HandleErr logs the provided error if it is not nil.

Intended for simple error handling where only logging is required.
*/
func HandleErr(err error) {
	if err != nil {
		log.Error(err)
		return
	}
}
