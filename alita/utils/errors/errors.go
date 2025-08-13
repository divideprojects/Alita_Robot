package errors

import (
	"fmt"
	"runtime"
	"strings"
)

type WrappedError struct {
	Err      error
	Message  string
	File     string
	Line     int
	Function string
}

func (e *WrappedError) Error() string {
	return fmt.Sprintf("%s at %s:%d in %s: %v", e.Message, e.File, e.Line, e.Function, e.Err)
}

func (e *WrappedError) Unwrap() error {
	return e.Err
}

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("%s: %w", message, err)
	}

	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = fn.Name()
		parts := strings.Split(funcName, "/")
		if len(parts) > 0 {
			funcName = parts[len(parts)-1]
		}
	}

	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		file = strings.Join(parts[len(parts)-2:], "/")
	}

	return &WrappedError{
		Err:      err,
		Message:  message,
		File:     file,
		Line:     line,
		Function: funcName,
	}
}

func Wrapf(err error, format string, args ...interface{}) error {
	return Wrap(err, fmt.Sprintf(format, args...))
}

func New(message string) error {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("%s", message)
	}

	fn := runtime.FuncForPC(pc)
	funcName := "unknown"
	if fn != nil {
		funcName = fn.Name()
		parts := strings.Split(funcName, "/")
		if len(parts) > 0 {
			funcName = parts[len(parts)-1]
		}
	}

	parts := strings.Split(file, "/")
	if len(parts) > 2 {
		file = strings.Join(parts[len(parts)-2:], "/")
	}

	return &WrappedError{
		Err:      fmt.Errorf("%s", message),
		Message:  message,
		File:     file,
		Line:     line,
		Function: funcName,
	}
}

func Newf(format string, args ...interface{}) error {
	return New(fmt.Sprintf(format, args...))
}