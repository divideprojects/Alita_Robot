package error_handling

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// FatalError logs an error and exits the program with exit code 1.
// This function is renamed from the previous misleading implementation that didn't exit.
func FatalError(funcName, modName string, err error) {
	if err != nil {
		log.Errorf("[%s][%s] Fatal error: %v", modName, funcName, err)
		os.Exit(1)
	}
}

// LogError logs an error but doesn't exit the program.
// Use this instead of FatalError when you want to log but continue execution.
func LogError(funcName, modName string, err error) {
	if err != nil {
		log.Errorf("[%s][%s] %v", modName, funcName, err)
	}
}

// HandleErr handles errors by logging them.
func HandleErr(err error) {
	if err != nil {
		log.Error(err)
	}
}

// WrapError wraps an error with additional context information.
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// RecoverFromPanic recovers from a panic and logs it as an error.
// This should be used with defer in goroutines to prevent crashes.
func RecoverFromPanic(funcName, modName string) {
	if r := recover(); r != nil {
		log.Errorf("[%s][%s] Recovered from panic: %v", modName, funcName, r)
	}
}
