package djson

type TokenScanner interface {
	Forward()
	PushEnds(...TokenType)
	PopEnds(...TokenType)
	Scan() (end bool, err error)
	Token() *Token
	Copy() TokenScanner
	ShouldEnd(token *Token) bool
	EndAt() TokenType
}

type CachedTokenScanner interface {
	TokenScanner
	CacheToEnd() error
	ResetRead()
}

type tokenScanner struct {
	lexer            Lexer
	token            Token
	endsWhen         *endsWhen
	forwardRequested bool
}

var _ TokenScanner = &tokenScanner{}

func NewTokenScanner(l Lexer, ends ...TokenType) *tokenScanner {
	n := &tokenScanner{
		lexer:            l,
		endsWhen:         newEndsWhen(),
		forwardRequested: true,
	}
	return n
}

func (t *tokenScanner) Copy() TokenScanner {
	return &tokenScanner{
		lexer:            t.lexer,
		endsWhen:         t.endsWhen.copy(),
		token:            t.token,
		forwardRequested: t.forwardRequested,
	}
}

func (t *tokenScanner) Forward() {
	t.forwardRequested = true
}

func (ts *tokenScanner) Token() *Token {
	return &ts.token
}

func (t *tokenScanner) PushEnds(tt ...TokenType) {
	t.endsWhen.push(tt...)
}

func (t *tokenScanner) PopEnds(tt ...TokenType) {
	t.endsWhen.pop(tt...)
}

func (t *tokenScanner) Scan() (end bool, err error) {
	defer func() {
		end = t.ShouldEnd(&t.token)
	}()
	if !t.forwardRequested {
		return
	}
	t.forwardRequested = false
	for {
		if err = t.lexer.NextToken(&t.token); err != nil {
			return
		}
		// skip comment
		if t.token.Type == TokenComment {
			continue
		}
		return
	}
}

func (t *tokenScanner) ShouldEnd(token *Token) bool {
	return token.Type == TokenEOF || t.endsWhen.ended(token.Type)
}

func (t *tokenScanner) EndAt() TokenType {
	return t.token.Type
}

type tokenRecordScanner struct {
	tokenScanner TokenScanner
	tokens       []*Token
	token        Token
	readOffset   int
}

func NewCachedTokenScanner(tokenScanner TokenScanner) CachedTokenScanner {
	return &tokenRecordScanner{
		tokenScanner: tokenScanner,
	}
}

func (t *tokenRecordScanner) Forward() {
	t.tokenScanner.Forward()
	t.readOffset++
}

func (ts *tokenRecordScanner) Token() *Token {
	token := ts.token
	return &token
}

func (t *tokenRecordScanner) PushEnds(tt ...TokenType) {
	t.tokenScanner.PushEnds(tt...)
}

func (t *tokenRecordScanner) PopEnds(tt ...TokenType) {
	t.tokenScanner.PopEnds(tt...)
}

func (t *tokenRecordScanner) ResetRead() {
	t.readOffset = 0
}

func (t *tokenRecordScanner) Scan() (end bool, err error) {
	if t.readOffset >= len(t.tokens) {
		var token Token
		if _, err = t.tokenScanner.Scan(); err != nil {
			return
		}
		token = *t.tokenScanner.Token()
		t.tokens = append(t.tokens, &token)
	}
	t.token = *t.tokens[t.readOffset]
	end = t.ShouldEnd(&t.token)
	return
}

func (t *tokenRecordScanner) CacheToEnd() error {
	offset := t.readOffset
	for {
		if end, err := t.Scan(); err != nil || end {
			t.readOffset = offset
			return err
		}
		t.Forward()
	}
}

func (t *tokenRecordScanner) Copy() TokenScanner {
	ret := &tokenRecordScanner{
		tokenScanner: t.tokenScanner.Copy(),
		readOffset:   t.readOffset,
		tokens:       make([]*Token, len(t.tokens)),
	}
	copy(ret.tokens, t.tokens)
	return ret
}

func (t *tokenRecordScanner) EndAt() TokenType {
	return t.token.Type
}

func (t *tokenRecordScanner) ShouldEnd(token *Token) bool {
	return t.tokenScanner.ShouldEnd(token)
}

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

func (ew *endsWhen) copy() *endsWhen {
	e := &endsWhen{
		when: make(map[TokenType]int, 16),
	}
	for i, k := range ew.when {
		e.when[i] = k
	}
	return e
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
