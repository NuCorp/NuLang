package errors

import (
	"errors"
	"fmt"
	"runtime"
	"runtime/debug"
)

func getErrorStackString() string {
	where := ""
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		where = string(debug.Stack())
	} else {
		f := runtime.FuncForPC(pc)
		where = fmt.Sprintf("%v:%v (func %v)", file, line, f.Name())
	}
	return where
}

func Stack(err error) error {
	if err == nil {
		return nil
	}
	return errors.Join(err, fmt.Errorf("| from > %v", getErrorStackString()))
}

func New(format string, v ...any) error {
	where := getErrorStackString()
	return fmt.Errorf(format+" %v", append(v, " at "+where)...)
}

func Format(format string, v ...any) error {
	return fmt.Errorf(format, v...)
}
