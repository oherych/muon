package muon

import "fmt"

type Error struct {
	Code int
	Msg  string
}

func (e Error) Error() string {
	return fmt.Sprintf("Code %d: %s", e.Code, e.Msg)
}

func newError(code int, str string, params ...interface{}) error {
	return Error{Code: code, Msg: fmt.Sprintf(str, params...)}
}
