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
	scanner := NewTokenScanner(NewMatcherLexer(r, t.bufSize))
	p := NewParser(scanner)
	val, _, err := p.Parse()
	if err != nil {
		return 0, err
	}
	return t.encoder.Encode(val, w)
}
