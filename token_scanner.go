package djson

type endsWhen struct {
	when map[TokenType]int
}

func newEndsWhen() *endsWhen {
	return &endsWhen{
		when: make(map[TokenType]int, 16),
	}
}

func (ew *endsWhen) push(tt ...TokenType) {
	for _, t := range tt {
		if v, ok := ew.when[t]; ok {
			ew.when[t] = v + 1
			continue
		}
		ew.when[t] = 1
	}
}

func (ew *endsWhen) pop(tt ...TokenType) {
	for _, t := range tt {
		if v, ok := ew.when[t]; ok {
			v -= 1
			ew.when[t] = v
		}
	}
}

func (ew *endsWhen) ended(t TokenType) bool {
	if v, ok := ew.when[t]; ok && v > 0 {
		return true
	}
	return false
}

type TokenScanner interface {
	Forward()
	PushEnds(...TokenType)
	PopEnds(...TokenType)
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
	endsWhen   *endsWhen
}

var _ TokenScanner = &tokenScanner{}

func NewTokenScanner(l Lexer, ends ...TokenType) *tokenScanner {
	n := &tokenScanner{
		lexer:    l,
		token:    &Token{},
		endsWhen: newEndsWhen(),
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
	t.endsWhen.push(tt...)
}

func (t *tokenScanner) PopEnds(tt ...TokenType) {
	t.endsWhen.pop(tt...)
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
	end = t.EndReached()
	return
}

func (t *tokenScanner) EndReached() bool {
	return t.endsWhen.ended(t.token.Type)
}

func (t *tokenScanner) EndAt() TokenType {
	return t.token.Type
}
