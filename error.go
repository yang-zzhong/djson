package djson

import "fmt"

type Error string

type ParseError struct {
	Row, Col     int
	Info         Error
	CurrentBytes []byte
}

const (
	UnexpectedChar Error = "unexpected char"
	UnexpectedEOF  Error = "unexpected eof"
)

func (e *ParseError) Error() string {
	if len(e.CurrentBytes) > 0 {
		return fmt.Sprintf("ParseError: %s '%s' at %d,%d", e.Info, e.CurrentBytes, e.Row, e.Col)
	}
	return fmt.Sprintf("ParseError: %s at %d,%d", e.Info, e.Row, e.Col)
}

func Is(err error, typ Error) bool {
	if e, ok := err.(*ParseError); ok {
		return e.Info == typ
	}
	return false
}
