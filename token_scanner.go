package djson

type TokenScanner interface {
	Forward()
	PushEnds(...TokenType)
	PopEnds(int)
	Offset() int
	SetOffset(int)
	Scan() (end bool, err error)
	Token() *Token
	EndAt() TokenType
}

type tokenScanner struct {
	lexer      Lexer
	tokens     []*Token
	readOffset int
	token      *Token
	endsWhen   []TokenType
}

var _ TokenScanner = &tokenScanner{}

func NewTokenScanner(l Lexer, ends ...TokenType) *tokenScanner {
	n := &tokenScanner{
		lexer:    l,
		token:    &Token{},
		endsWhen: ends,
	}
	return n
}

func (t *tokenScanner) Forward() {
	t.readOffset++
}

func (ts *tokenScanner) Token() *Token {
	return ts.token
}

func (t *tokenScanner) PushEnds(tt ...TokenType) {
	t.endsWhen = append(t.endsWhen, tt...)
}

func (t *tokenScanner) PopEnds(count int) {
	t.endsWhen = t.endsWhen[:len(t.endsWhen)-count]
}

func (t *tokenScanner) Offset() int {
	return t.readOffset
}

func (t *tokenScanner) SetOffset(offset int) {
	t.readOffset = offset
}

func (t *tokenScanner) Scan() (end bool, err error) {
	for i := len(t.tokens); i <= t.readOffset; i++ {
		if len(t.tokens) > 0 && t.tokens[len(t.tokens)-1].Type == TokenEOF {
			break
		}
		token := &Token{}
		if err = t.lexer.NextToken(token); err != nil {
			return
		}
		// skip comment
		if token.Type == TokenComment {
			i--
			continue
		}
		t.tokens = append(t.tokens, token)
	}
	if len(t.tokens) <= t.readOffset {
		end = true
		return
	}
	t.token = t.tokens[t.readOffset]
	if t.token.Type == TokenEOF {
		end = true
		return
	}
	for _, e := range t.endsWhen {
		if e == t.token.Type {
			end = true
			return
		}
	}
	return
}

func (t *tokenScanner) EndAt() TokenType {
	return t.token.Type
}
