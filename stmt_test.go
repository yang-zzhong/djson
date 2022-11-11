package djson

import "testing"

func TestSimpleStmt(t *testing.T) {
	// 5 + 2 - 1
	g := newLexMock([]*Token{
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'2'}},
		{Type: TokenMinus},
		{Type: TokenNumber, Raw: []byte{'1'}},
	})
	expr := newStmt(newTokenScanner(g), nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.typ == valueInt && expr.value.value.(int64) == 6) {
		t.Fatal("5 + 2 - 1 failed")
	}
	// 5 + 2 * 3
	g = newLexMock([]*Token{
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'2'}},
		{Type: TokenMultiplication},
		{Type: TokenNumber, Raw: []byte{'3'}},
	})
	expr = newStmt(newTokenScanner(g), nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.typ == valueInt && expr.value.value.(int64) == 11) {
		t.Fatal("5 + 2 * 3 failed")
	}
	// (5 + 2) * 3
	g = newLexMock([]*Token{
		{Type: TokenParenthesesOpen},
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'2'}},
		{Type: TokenParenthesesClose},
		{Type: TokenMultiplication},
		{Type: TokenNumber, Raw: []byte{'3'}},
	})
	expr = newStmt(newTokenScanner(g), nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.typ == valueInt && expr.value.value.(int64) == 21) {
		t.Fatal("(5 + 2) * 3 failed")
	}
	// "hello" + "world"
	g = newLexMock([]*Token{
		{Type: TokenString, Raw: []byte("hello")},
		{Type: TokenAddition},
		{Type: TokenString, Raw: []byte("world")},
	})
	expr = newStmt(newTokenScanner(g), nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.typ == valueString && string(expr.value.value.([]byte)) == "helloworld") {
		t.Fatal("hello world failed")
	}
}

func TestStmtAssignation(t *testing.T) {
	// a = 5 + 3
	g := newLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'a'}},
		{Type: TokenAssignation},
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'3'}},
	})
	vs := newVariables()
	expr := newStmt(newTokenScanner(g), vs)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	val := vs.get([]byte{'a'})
	if !(val.typ == valueInt && val.value.(int64) == 8) {
		t.Fatal("assign fatal")
	}
}

func TestStmtAssignationWithReduction(t *testing.T) {
	// a = true => 5 + 3
	g := newLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'a'}},
		{Type: TokenAssignation},
		{Type: TokenTrue},
		{Type: TokenReduction},
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'3'}},
	})
	vs := newVariables()
	expr := newStmt(newTokenScanner(g), vs)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	val := vs.get([]byte{'a'})
	if !(val.typ == valueInt && val.value.(int64) == 8) {
		t.Fatal("assign fatal")
	}
	// a = false => 5 + 3
	g.tokens[2] = &Token{Type: TokenFalse}
	g.offset = 0
	vs = newVariables()
	expr = newStmt(newTokenScanner(g), vs)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	val = vs.get([]byte{'a'})
	if val.typ != valueNull {
		t.Fatal("assign fatal")
	}
}

func TestStmtObjectOperate(t *testing.T) {
	// {"hello": "world"} + {"world": "hello"}
	g := newLexMock([]*Token{
		{Type: TokenBraceOpen},
		{Type: TokenString, Raw: []byte("hello")},
		{Type: TokenColon},
		{Type: TokenString, Raw: []byte("world")},
		{Type: TokenBraceClose},
		{Type: TokenAddition},
		{Type: TokenBraceOpen},
		{Type: TokenString, Raw: []byte("world")},
		{Type: TokenColon},
		{Type: TokenString, Raw: []byte("hello")},
		{Type: TokenBraceClose},
	})
	vs := newVariables()
	expr := newStmt(newTokenScanner(g), vs)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if expr.value.typ != valueObject {
		t.Fatal("ret type error")
	}
	val := expr.value.value.(*object)
	hello := val.get([]byte("hello"))
	if !(hello.typ == valueString && string(hello.value.([]byte)) == "world") {
		t.Fatal("value of hello error")
	}
	world := val.get([]byte("world"))
	if !(world.typ == valueString && string(world.value.([]byte)) == "hello") {
		t.Fatal("value of world error")
	}
}
