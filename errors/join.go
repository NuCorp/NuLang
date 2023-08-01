package errors

import (
	"errors"
	"strings"
)

func Join(errs ...error) error {
	return errors.Join(errs...)
}

type WrappedError []error

func (w WrappedError) Error() string {
	var errorMessage string
	for idx, err := range w {
		if idx == 0 {
			errorMessage += "* (wrapped error)\n| " + strings.ReplaceAll(err.Error(), "\n", "\n| ")
		} else {
			toAdd := "+" + strings.Repeat("-", idx) + " due to: "
			errorMessage += toAdd + strings.ReplaceAll(err.Error(), "\n", "\n|"+strings.Repeat(" ", len(toAdd)-3)+"| ")
		}
		errorMessage += "\n|\n"
	}
	errorMessage += "x (end wrapped error)"
	return errorMessage
}
func (w WrappedError) Unwrap() error {
	if len(w) == 2 {
		return w[1]
	}
	return w[1:]
}

func Wrap(err1, err2 error) error {
	switch err1 := err1.(type) {
	case WrappedError:
		return append(err1, err2)
	default:
		return WrappedError{err1, err2}
	}
}

func Unwrap(err error) error {
	return errors.Unwrap(err)
}
