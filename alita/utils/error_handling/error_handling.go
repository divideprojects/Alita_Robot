package error_handling

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

/*
FatalError logs an error with function and module context.

If an error is provided, it logs the error message with the given function and module names.
Does not exit the program, only logs the error.
*/
func FatalError(funcName, modName string, err error) {
	if err != nil {
		log.WithFields(log.Fields{
			"module":   modName,
			"function": funcName,
		}).Error(err)
	}
}

/*
HandleErr logs the provided error if it is not nil.

Intended for simple error handling where only logging is required.
*/
func HandleErr(err error) {
	if err != nil {
		log.Error(err)
	}
}

/*
WrapError wraps an error with additional context information.

Returns a new error with the provided context message and the original error.
*/
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

/*
WrapErrorf wraps an error with formatted context information.

Returns a new error with the formatted context message and the original error.
*/
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf(format+": %w", append(args, err)...)
}
