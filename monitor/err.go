package monitor

import "fmt"

type Error struct {
	code uint32
	errMsg string
	where string
}

func (e *Error) Error() string{
	return fmt.Sprintf("code=%d;msg=%s",e.code,e.errMsg)
}

func ErrorNew(code uint32,msg string, where string) *Error{
	return &Error{code,msg,where}
}
