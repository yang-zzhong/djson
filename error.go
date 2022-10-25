package djson

import "fmt"

type err_ string

type Error struct {
	Row, Col     int
	Info         err_
	CurrentBytes []byte
}

const (
	UnexpectedChar  err_ = "unexpected char"
	UnexpectedToken err_ = "unexpected token"
	UnexpectedEOF   err_ = "unexpected eof"
)

var (
	ErrUnexpectedEOF = &Error{
		Info: UnexpectedEOF,
	}
)

func (e *Error) Error() string {
	if len(e.CurrentBytes) > 0 {
		return fmt.Sprintf("%s '%s' at %d,%d", e.Info, e.CurrentBytes, e.Row, e.Col)
	} else if !(e.Row == 0 && e.Col == 0) {
		return fmt.Sprintf("%s at %d,%d", e.Info, e.Row, e.Col)
	} else {
		return fmt.Sprintf("%s", e.Info)
	}
}

func ErrFromToken(info err_, token *Token) error {
	return &Error{
		Col:          token.Col,
		Row:          token.Row,
		CurrentBytes: token.Raw,
		Info:         info,
	}
}

func Is(err error, typ string) bool {
	if e, ok := err.(*Error); ok {
		return string(e.Info) == typ
	}
	return false
}
