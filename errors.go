package httpr

import (
	"errors"
	"fmt"
)

type ResponseError struct {
	Msg  string
	Code int
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("got HTTP error code '%d': %s", e.Code, e.Msg)
}

func (e *ResponseError) Is(target error) bool {
	var respErr *ResponseError
	if ok := errors.As(target, &respErr); !ok {
		return false
	}

	return e.Code == respErr.Code
}

func newResponseError(msg string, code int) error {
	return &ResponseError{Msg: msg, Code: code}
}
