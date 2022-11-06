package djson

type lexmock struct {
	offset int
	tokens []*Token
}

func (g *lexmock) NextToken(token *Token) error {
	if len(g.tokens) == g.offset {
		token.Type = TokenEOF
		return nil
	}
	token.Raw = g.tokens[g.offset].Raw
	token.Type = g.tokens[g.offset].Type
	g.offset++
	return nil
}

func newLexMock(tokens []*Token) *lexmock {
	return &lexmock{
		tokens: tokens,
	}
}
