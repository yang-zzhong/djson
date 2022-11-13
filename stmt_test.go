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
	expr := newStmt(NewTokenScanner(g), nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.Type == ValueInt && expr.value.Value.(int64) == 6) {
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
	expr = newStmt(NewTokenScanner(g), nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.Type == ValueInt && expr.value.Value.(int64) == 11) {
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
	expr = newStmt(NewTokenScanner(g), nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.Type == ValueInt && expr.value.Value.(int64) == 21) {
		t.Fatal("(5 + 2) * 3 failed")
	}
	// "hello" + "world"
	g = newLexMock([]*Token{
		{Type: TokenString, Raw: []byte("hello")},
		{Type: TokenAddition},
		{Type: TokenString, Raw: []byte("world")},
	})
	expr = newStmt(NewTokenScanner(g), nil)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.Type == ValueString && string(expr.value.Value.([]byte)) == "helloworld") {
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
	expr := newStmt(NewTokenScanner(g), vs)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	val := vs.get([]byte{'a'})
	if !(val.Type == ValueInt && val.Value.(int64) == 8) {
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
	expr := newStmt(NewTokenScanner(g), vs)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	val := vs.get([]byte{'a'})
	if !(val.Type == ValueInt && val.Value.(int64) == 8) {
		t.Fatal("assign fatal")
	}
	// a = false => 5 + 3
	g.tokens[2] = &Token{Type: TokenFalse}
	g.offset = 0
	vs = newVariables()
	expr = newStmt(NewTokenScanner(g), vs)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	val = vs.get([]byte{'a'})
	if val.Type != ValueNull {
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
	expr := newStmt(NewTokenScanner(g), vs)
	if err := expr.execute(); err != nil {
		t.Fatal(err)
	}
	if expr.value.Type != ValueObject {
		t.Fatal("ret type error")
	}
	val := expr.value.Value.(Object)
	hello := val.Get([]byte("hello"))
	if !(hello.Type == ValueString && string(hello.Value.([]byte)) == "world") {
		t.Fatal("value of hello error")
	}
	world := val.Get([]byte("world"))
	if !(world.Type == ValueString && string(world.Value.([]byte)) == "hello") {
		t.Fatal("value of world error")
	}
}

func TestStmtObjectCall(t *testing.T) {
	// {"hello": "world"}.set(k == "hello" => v + " ^_^")
	g := newLexMock([]*Token{
		{Type: TokenBraceOpen},
		{Type: TokenString, Raw: []byte("hello")},
		{Type: TokenColon},
		{Type: TokenString, Raw: []byte("world")},
		{Type: TokenBraceClose},
		{Type: TokenDot},
		{Type: TokenIdentifier, Raw: []byte("set")},
		{Type: TokenParenthesesOpen},
		{Type: TokenIdentifier, Raw: []byte("k")},
		{Type: TokenEqual},
		{Type: TokenString, Raw: []byte("hello")},
		{Type: TokenReduction},
		{Type: TokenIdentifier, Raw: []byte("v")},
		{Type: TokenAddition},
		{Type: TokenString, Raw: []byte(" ^_^")},
		{Type: TokenParenthesesClose},
	})
	vs := newVariables()
	stmt := newStmt(NewTokenScanner(g), vs)
	if err := stmt.execute(); err != nil {
		t.Fatal(err)
	}
	if stmt.value.Type != ValueObject {
		t.Fatal("type error")
	}
	hello := stmt.value.Value.(Object).Get([]byte("hello"))
	if !(hello.Type == ValueString && string(hello.Value.([]byte)) == "world ^_^") {
		t.Fatal("value error")
	}
}
