package djson

type tokens struct {
	tokens []*Token
	total  int
}

type endsWhen struct {
	total int
	when  []TokenType
}

func newEndsWhen() *endsWhen {
	return &endsWhen{
		when: make([]TokenType, 16),
	}
}

func (ew *endsWhen) push(tt ...TokenType) {
	l := len(ew.when)
	if l <= ew.total {
		tmp := make([]TokenType, l*2)
		copy(tmp, ew.when)
		ew.when = tmp
	}
	for i, t := range tt {
		ew.when[ew.total+i] = t
	}
	ew.total += len(tt)
}

func (ew *endsWhen) pop(n int) {
	ew.total -= n
	if ew.total < 0 {
		ew.total = 0
	}
}

func (ew *endsWhen) ended(t TokenType) bool {
	for i := 0; i < ew.total; i++ {
		if ew.when[i] == t {
			return true
		}
	}
	return false
}

func (ew *endsWhen) reset() {
	ew.total = 0
}

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

func (t *tokenScanner) PopEnds(count int) {
	t.endsWhen.pop(count)
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
	end = t.endsWhen.ended(t.token.Type)
	return
}

func (t *tokenScanner) EndAt() TokenType {
	return t.token.Type
}
