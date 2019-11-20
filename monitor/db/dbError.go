package db


type NoDriverError struct {
	errMsg 	string
	Err 	error
}

func (e *NoDriverError) Error() string {
	var s = e.errMsg
	if e.Err != nil {
		s += ": " + e.Err.Error()
	}
	return s
}