package agent

import (
	"fmt"
)

type ExtendedError struct {
	Message string
}

func NewError(msg string) *ExtendedError {
	return &ExtendedError{msg}
}

func (e *ExtendedError) Error() string {
	return e.Message
}

func (e *ExtendedError) Add(data any) error {
	var res string
	switch v := data.(type) {
	case byte:
		res = string(v)
	case rune:
		res = string(v)
	default:
		res = fmt.Sprint(data)
	}

	return &ExtendedError{e.Message + ": " + res} // что я творю
}

var (
	errServerInternal = NewError("сервер вернул 500")
	errInvalidSend    = NewError("сервер вернул 422")
	errStatusUnknown  = NewError("непредвиденная ошибка сервера")
)
