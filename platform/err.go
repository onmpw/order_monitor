package platform

import "fmt"

type Error struct {
	code uint32
	errMsg string
	where string
}

func (e *Error) Error() string{
	return fmt.Sprintf("code=%d;msg=%s",e.code,e.errMsg)
}

func ErrorNew() *Error{
	return &Error{1,"你错了","不知道在哪错的"}
}
