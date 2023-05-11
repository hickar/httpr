package httpr

import (
	"fmt"
)

type RespErr struct {
	Msg  string
	Code int
}

func (e *RespErr) Error() string {
	return fmt.Sprintf("got HTTP error code '%d': %s", e.Code, e.Msg)
}

func (e *RespErr) Is(target error) bool {
	t, ok := target.(*RespErr)
	if !ok {
		return false
	}

	return e.Code == t.Code
}

func newResponseError(msg string, code int) error {
	return &RespErr{Msg: msg, Code: code}
}
