package djson

import "testing"

type getter struct {
	offset int
	tokens []*Token
}

func (g *getter) NextToken(token *Token) error {
	if len(g.tokens) == g.offset {
		token.Type = TokenEOF
		return nil
	}
	token.Raw = g.tokens[g.offset].Raw
	token.Type = g.tokens[g.offset].Type
	g.offset++
	return nil
}

func TestExpr(t *testing.T) {
	// 5 + 2 - 1
	g := &getter{tokens: []*Token{
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'2'}},
		{Type: TokenMinus},
		{Type: TokenNumber, Raw: []byte{'1'}},
	}}
	expr := newExpr(g, nil, nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.typ == valueInt && expr.value.value.(int64) == 6) {
		t.Fatal("5 + 2 - 1 failed")
	}
	// 5 + 2 * 3
	g = &getter{tokens: []*Token{
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'2'}},
		{Type: TokenMultiplication},
		{Type: TokenNumber, Raw: []byte{'3'}},
	}}
	expr = newExpr(g, nil, nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.typ == valueInt && expr.value.value.(int64) == 11) {
		t.Fatal("5 + 2 * 3 failed")
	}
	// (5 + 2) * 3
	g = &getter{tokens: []*Token{
		{Type: TokenParenthesesOpen},
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition, Raw: []byte{'+'}},
		{Type: TokenNumber, Raw: []byte{'2'}},
		{Type: TokenParenthesesClose},
		{Type: TokenMultiplication},
		{Type: TokenNumber, Raw: []byte{'3'}},
	}}
	expr = newExpr(g, nil, nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.typ == valueInt && expr.value.value.(int64) == 21) {
		t.Fatal("(5 + 2) * 3 failed")
	}
	// "hello" + "world"
	g = &getter{tokens: []*Token{
		{Type: TokenString, Raw: []byte("\"hello\"")},
		{Type: TokenAddition},
		{Type: TokenString, Raw: []byte("\"world\"")},
	}}
	expr = newExpr(g, nil, nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.typ == valueString && string(expr.value.value.([]byte)) == "helloworld") {
		t.Fatal("hello world failed")
	}
}
