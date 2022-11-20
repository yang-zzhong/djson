package djson

import "io"

// Translator a translator for translating djson to a json or other result format
// depending the Encoder,
type Translator interface {
	Translate(r io.Reader, w io.Writer) (int, error)
}

// Encoder encode the Value that interpeter constructed from djson to a result format
type Encoder interface {
	// Encode the Value to a result format
	Encode(val Value, w io.Writer) (int, error)
}

type translator struct {
	encoder Encoder
	bufSize uint
	ctx     Context
}

// BufSize set a buffer size for translator
func BuffSize(bufSize uint) func(*translator) {
	return func(opt *translator) {
		opt.bufSize = bufSize
	}
}

// Ctx set a Context for translator
func Ctx(ctx Context) func(*translator) {
	return func(opt *translator) {
		opt.ctx = ctx
	}
}

// NewTranslator new a translator
func NewTranslator(e Encoder, opts ...func(*translator)) *translator {
	t := &translator{encoder: e}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Translate implements ths Translator
func (t *translator) Translate(r io.Reader, w io.Writer) (int, error) {
	scanner := NewTokenScanner(NewLexer(r, t.bufSize))
	scanner.PushEnds(TokenSemicolon)
	defer scanner.PopEnds(TokenSemicolon)
	if t.ctx == nil {
		t.ctx = NewContext()
	}
	stmt := NewStmtExecutor(scanner, t.ctx)
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
