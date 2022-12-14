package djson

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestStmt_arithmatic(t *testing.T) {
	// 5 + 2 - 1
	g := NewLexMock([]*Token{
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'2'}},
		{Type: TokenMinus},
		{Type: TokenNumber, Raw: []byte{'1'}},
	})
	expr := NewStmtExecutor(NewTokenScanner(g), nil)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	if expr.value.Type != ValueInt {
		t.Fatal("5 + 2 - 1 failed")
	}
	if rv, _ := expr.value.Value.(Inter).Int(); rv != 6 {
		t.Fatal("5 + 2 - 1 failed")
	}
	// 5 + 2 * 3
	g = NewLexMock([]*Token{
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'2'}},
		{Type: TokenMultiplication},
		{Type: TokenNumber, Raw: []byte{'3'}},
	})
	expr = NewStmtExecutor(NewTokenScanner(g), nil)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	if expr.value.Type != ValueInt {
		t.Fatal("5 + 2 * 3 failed")
	}
	if rv, _ := expr.value.Value.(Inter).Int(); rv != 11 {
		t.Fatal("5 + 2 * 3 failed")
	}
	// (5 + 2) * 3
	g = NewLexMock([]*Token{
		{Type: TokenParenthesesOpen},
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'2'}},
		{Type: TokenParenthesesClose},
		{Type: TokenMultiplication},
		{Type: TokenNumber, Raw: []byte{'3'}},
	})
	expr = NewStmtExecutor(NewTokenScanner(g), nil)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	if expr.value.Type != ValueInt {
		t.Fatal("(5 + 2) * 3 failed")
	}
	if rv, _ := expr.value.Value.(Inter).Int(); rv != 21 {
		t.Fatal("(5 + 2) * 3 failed")
	}
	// "hello" + "world"
	g = NewLexMock([]*Token{
		{Type: TokenString, Raw: []byte("hello")},
		{Type: TokenAddition},
		{Type: TokenString, Raw: []byte("world")},
	})
	expr = NewStmtExecutor(NewTokenScanner(g), nil)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	if !(expr.value.Type == ValueString && string(expr.value.Value.(String).Bytes()) == "helloworld") {
		t.Fatal("hello world failed")
	}
}

func TestStmt_assignation(t *testing.T) {
	g := NewLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'a'}},
		{Type: TokenAssignation},
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'3'}},
	})
	vs := NewContext(Variable{Name: []byte{'a'}})
	expr := NewStmtExecutor(NewTokenScanner(g), vs)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	val := vs.ValueOf([]byte{'a'})
	if val.Type != ValueInt {
		t.Fatal("assign fatal")
	}
	if rv, _ := val.Value.(Inter).Int(); rv != 8 {
		t.Fatal("assign fatal")
	}
}

func TestStmt_assignationWithReduction(t *testing.T) {
	// a = true => 5 + 3
	g := NewLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte{'a'}},
		{Type: TokenAssignation},
		{Type: TokenTrue},
		{Type: TokenReduction},
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenAddition},
		{Type: TokenNumber, Raw: []byte{'3'}},
	})
	vs := NewContext(Variable{Name: []byte{'a'}})
	expr := NewStmtExecutor(NewTokenScanner(g), vs)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	val := vs.ValueOf([]byte{'a'})
	if val.Type != ValueInt {
		t.Fatal("assign fatal")
	}
	if rv, _ := val.Value.(Inter).Int(); rv != 8 {
		t.Fatal("assign fatal")
	}
	// a = false => 5 + 3
	g.tokens[2] = &Token{Type: TokenFalse}
	g.offset = 0
	vs = NewContext()
	expr = NewStmtExecutor(NewTokenScanner(g), vs)
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
	g := NewLexMock([]*Token{
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
	expr := NewStmtExecutor(NewTokenScanner(g), vs)
	if err := expr.Execute(); err != nil {
		t.Fatal(err)
	}
	if expr.value.Type != ValueObject {
		t.Fatal("ret type error")
	}
	val := expr.value.Value.(Object)
	hello := val.Get([]byte("hello"))
	if !(hello.Type == ValueString && string(hello.Value.(String).Bytes()) == "world") {
		t.Fatal("value of hello error")
	}
	world := val.Get([]byte("world"))
	if !(world.Type == ValueString && string(world.Value.(String).Bytes()) == "hello") {
		t.Fatal("value of world error")
	}
}

func TestStmt_objectCall(t *testing.T) {
	// {"hello": "world"}.map(k == "hello" => v + " ^_^")
	g := NewLexMock([]*Token{
		{Type: TokenBraceOpen},
		{Type: TokenString, Raw: []byte("hello")},
		{Type: TokenColon},
		{Type: TokenString, Raw: []byte("world")},
		{Type: TokenBraceClose},
		{Type: TokenDot},
		{Type: TokenIdentifier, Raw: []byte("map")},
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
	stmt := NewStmtExecutor(NewTokenScanner(g), vs)
	if err := stmt.Execute(); err != nil {
		t.Fatal(err)
	}
	if stmt.value.Type != ValueObject {
		t.Fatal("type error")
	}
	hello := stmt.value.Value.(Object).Get([]byte("hello"))
	if !(hello.Type == ValueString && string(hello.Value.(String).Bytes()) == "world ^_^") {
		t.Fatal("value error")
	}
}

func TestStmt_call(t *testing.T) {
	// {"0": 1}.map(k == "0" => 4).map(v == 4 => 5)
	g := NewLexMock([]*Token{
		{Type: TokenBraceOpen},
		{Type: TokenString, Raw: []byte("0")},
		{Type: TokenColon},
		{Type: TokenNumber, Raw: []byte{'1'}},
		{Type: TokenBraceClose},
		{Type: TokenDot},
		{Type: TokenIdentifier, Raw: []byte("map")},
		{Type: TokenParenthesesOpen},
		{Type: TokenIdentifier, Raw: []byte{'k'}},
		{Type: TokenEqual},
		{Type: TokenString, Raw: []byte{'0'}},
		{Type: TokenReduction},
		{Type: TokenNumber, Raw: []byte{'4'}},
		{Type: TokenParenthesesClose},
		{Type: TokenDot},
		{Type: TokenIdentifier, Raw: []byte("map")},
		{Type: TokenParenthesesOpen},
		{Type: TokenIdentifier, Raw: []byte{'v'}},
		{Type: TokenEqual},
		{Type: TokenNumber, Raw: []byte{'4'}},
		{Type: TokenReduction},
		{Type: TokenNumber, Raw: []byte{'5'}},
		{Type: TokenParenthesesClose},
	})
	stmt := NewStmtExecutor(NewTokenScanner(g), NewContext())
	if err := stmt.Execute(); err != nil {
		t.Fatal(err)
	}
	if stmt.value.Type != ValueObject {
		t.Fatal("failed")
	}
	val := stmt.value.Value.(Object).Get([]byte{'0'})
	if val.Type != ValueInt {
		t.Fatal("call fatal")
	}
	if rv, _ := val.Value.(Inter).Int(); rv != 5 {
		t.Fatal("call fatal")
	}
}

func TestStmt_config(t *testing.T) {
	data := `
# a config test

users = [1 ... 10000].map({
    "username": "user" + v,
    "password": "password" + i,
    "i": i
}).map(v + {"odd": v.i%2 != 0});
    `
	r := strings.NewReader(data)
	scanner := NewTokenScanner(NewLexer(r, 128))
	stmt := NewStmtExecutor(scanner, NewContext())
	if err := stmt.Execute(); err != nil {
		t.Fatal(err)
	}
	t.Log(stmt.Value())
}

func BenchmarkStmt_config(b *testing.B) {
	data := `
# a config test

users = [1 ... 10000].map({
    "username": "user" + v,
    "password": "password" + i,
    "i": i
}).map(v + {"odd": v.i%2 != 0});
    `
	r := strings.NewReader(data)
	for i := 0; i < b.N; i++ {
		r.Seek(0, io.SeekStart)
		scanner := NewTokenScanner(NewLexer(r, 128))
		stmt := NewStmtExecutor(scanner, NewContext())
		if err := stmt.Execute(For(NullValue())); err != nil {
			b.Fatal(err)
		}
	}
}

func Benchmark_goconfig(b *testing.B) {
	for j := 0; j < b.N; j++ {
		m := make([]map[string]interface{}, 10000)
		for i := 0; i < 10000; i++ {
			m[i] = map[string]interface{}{
				"username": "user" + fmt.Sprintf("%d", i+1),
				"password": "password" + fmt.Sprintf("%d", 1),
				"i":        i,
			}
		}
		for i, item := range m {
			item["odd"] = item["i"].(int)%2 != 0
			m[i] = item
		}
	}
}

func BenchmarkStmt_arithmatic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// ((5 + 2) * 3 == 21) || false => "hello world"
		g := NewLexMock([]*Token{
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
		stmt := NewStmtExecutor(NewTokenScanner(g), nil)
		if err := stmt.Execute(); err != nil {
			b.Fatal(err)
		}
		if !(stmt.value.Type == ValueString && bytes.Equal(stmt.value.Value.(String).Bytes(), []byte("hello world"))) {
			b.Fatal("failed")
		}
	}
}
