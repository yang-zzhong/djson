package djson

import (
	"bytes"
	"testing"
)

func TestStmt_arithmatic(t *testing.T) {
	// 5 + 2 - 1
	g := newLexMock([]*Token{
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'2'}},
		{Type: TokenMinus},
		{Type: TokenNumber, Raw: []byte{'1'}},
	})
	expr := NewStmt(NewTokenScanner(g), nil)
	if err := expr.Execute(); err != nil {
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
	expr = NewStmt(NewTokenScanner(g), nil)
	if err := expr.Execute(); err != nil {
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
	expr = NewStmt(NewTokenScanner(g), nil)
	if err := expr.Execute(); err != nil {
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
	expr = NewStmt(NewTokenScanner(g), nil)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.Type == ValueString && string(expr.value.Value.(String).Literal()) == "helloworld") {
		t.Fatal("hello world failed")
	}
}

func TestStmt_assignation(t *testing.T) {
	// a = 5 + 3
	g := newLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'a'}},
		{Type: TokenAssignation},
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'3'}},
	})
	vs := NewContext()
	expr := NewStmt(NewTokenScanner(g), vs)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	val := vs.ValueOf([]byte{'a'})
	if !(val.Type == ValueInt && val.Value.(int64) == 8) {
		t.Fatal("assign fatal")
	}
}

func TestStmt_assignationWithReduction(t *testing.T) {
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
	vs := NewContext()
	expr := NewStmt(NewTokenScanner(g), vs)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	val := vs.ValueOf([]byte{'a'})
	if !(val.Type == ValueInt && val.Value.(int64) == 8) {
		t.Fatal("assign fatal")
	}
	// a = false => 5 + 3
	g.tokens[2] = &Token{Type: TokenFalse}
	g.offset = 0
	vs = NewContext()
	expr = NewStmt(NewTokenScanner(g), vs)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	val = vs.ValueOf([]byte{'a'})
	if val.Type != ValueNull {
		t.Fatal("assign fatal")
	}
}

func TestStmt_objectOperate(t *testing.T) {
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
	vs := NewContext()
	expr := NewStmt(NewTokenScanner(g), vs)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	if expr.value.Type != ValueObject {
		t.Fatal("ret type error")
	}
	val := expr.value.Value.(Object)
	hello := val.Get([]byte("hello"))
	if !(hello.Type == ValueString && string(hello.Value.(String).Literal()) == "world") {
		t.Fatal("value of hello error")
	}
	world := val.Get([]byte("world"))
	if !(world.Type == ValueString && string(world.Value.(String).Literal()) == "hello") {
		t.Fatal("value of world error")
	}
}

func TestStmt_objectCall(t *testing.T) {
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
	vs := NewContext()
	stmt := NewStmt(NewTokenScanner(g), vs)
	if err := stmt.Execute(); err != nil {
		t.Fatal(err)
	}
	if stmt.value.Type != ValueObject {
		t.Fatal("type error")
	}
	hello := stmt.value.Value.(Object).Get([]byte("hello"))
	if !(hello.Type == ValueString && string(hello.Value.(String).Literal()) == "world ^_^") {
		t.Fatal("value error")
	}
}

func BenchmarkStmt_arithmatic(b *testing.B) {

	for i := 0; i < b.N; i++ {
		// ((5 + 2) * 3 == 21) || false => "hello world"
		g := newLexMock([]*Token{
			{Type: TokenParenthesesOpen},
			{Type: TokenParenthesesOpen},
			{Type: TokenNumber, Raw: []byte{'5'}},
			{Type: TokenAddition},
			{Type: TokenNumber, Raw: []byte{'2'}},
			{Type: TokenParenthesesClose},
			{Type: TokenMultiplication},
			{Type: TokenNumber, Raw: []byte{'3'}},
			{Type: TokenEqual},
			{Type: TokenNumber, Raw: []byte{'2', '1'}},
			{Type: TokenOr},
			{Type: TokenTrue},
			{Type: TokenReduction},
			{Type: TokenString, Raw: []byte("hello world")},
		})
		stmt := NewStmt(NewTokenScanner(g), nil)
		if err := stmt.Execute(); err != nil {
			b.Fatal(err)
		}
		if !(stmt.value.Type == ValueString && bytes.Equal(stmt.value.Value.(String).Literal(), []byte("hello world"))) {
			b.Fatal("(5 + 2) * 3 failed")
		}
	}
}
