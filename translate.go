package djson

import "io"

type Translator interface {
	Translate(r io.Reader, w io.Writer) (int, error)
}

type Encoder interface {
	Encode(val Value, w io.Writer) (int, error)
}

type translator struct {
	encoder Encoder
	bufSize uint
}

func BuffSize(bufSize uint) func(*translator) {
	return func(opt *translator) {
		opt.bufSize = bufSize
	}
}

func NewTranslator(e Encoder, opts ...func(*translator)) *translator {
	t := &translator{encoder: e}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (t translator) Translate(r io.Reader, w io.Writer) (int, error) {
	scanner := NewTokenScanner(NewLexer(r, t.bufSize))
	scanner.PushEnds(TokenSemicolon)
	defer scanner.PopEnds(TokenSemicolon)
	vars := NewContext()
	stmt := NewStmt(scanner, vars)
	var val Value
	for {
		if err := stmt.Execute(); err != nil {
			return 0, err
		}
		if stmt.value.Type != ValueNull {
			val = stmt.value
		}
		if scanner.EndAt() == TokenEOF {
			break
		}
	}
	return t.encoder.Encode(val, w)
}
