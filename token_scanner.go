package djson

type tokenScanner struct {
	lexer      lexer
	tokens     []*Token
	readOffset int
	token      *Token
	endsWhen   []TokenType
	ended      bool
}

func newTokenScanner(l lexer, ends ...TokenType) *tokenScanner {
	n := &tokenScanner{
		lexer:    l,
		token:    &Token{},
		endsWhen: ends,
	}
	return n
}

func (t *tokenScanner) forward() {
	t.readOffset++
}

func (t *tokenScanner) pushEnds(tt ...TokenType) {
	t.endsWhen = append(t.endsWhen, tt...)
}

func (t *tokenScanner) popEnds(count int) {
	t.endsWhen = t.endsWhen[:len(t.endsWhen)-count]
}

func (t *tokenScanner) offset() int {
	return t.readOffset
}

func (t *tokenScanner) setOffset(offset int) {
	t.readOffset = offset
}

func (t *tokenScanner) scan() (end bool, err error) {
	for i := len(t.tokens); i <= t.readOffset; i++ {
		if len(t.tokens) > 0 && t.tokens[len(t.tokens)-1].Type == TokenEOF {
			end = true
			return
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

func (t *tokenScanner) endAt() TokenType {
	return t.token.Type
}
