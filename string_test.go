package djson

import (
	"bytes"
	"testing"
)

func TestString_Index(t *testing.T) {
	// "hello world".index("world")
	scanner := NewTokenScanner(newLexMock([]*Token{
		{Type: TokenString, Raw: []byte("hello world")},
		{Type: TokenDot},
		{Type: TokenIdentifier, Raw: []byte("index")},
		{Type: TokenParenthesesOpen},
		{Type: TokenString, Raw: []byte("world")},
		{Type: TokenParenthesesClose},
	}))
	stmt := NewStmtExecutor(scanner, NewContext())
	if err := stmt.Execute(); err != nil {
		t.Fatal(err)
	}
	if stmt.value.Type != ValueInt {
		t.Fatal("index error")
	}
	v, _ := stmt.value.Value.(Inter).Int()
	if v != 6 {
		t.Fatal("index error")
	}
}

func TestString_Match(t *testing.T) {
	// "hello world".match("world$")
	scanner := NewTokenScanner(newLexMock([]*Token{
		{Type: TokenString, Raw: []byte("hello world")},
		{Type: TokenDot},
		{Type: TokenIdentifier, Raw: []byte("match")},
		{Type: TokenParenthesesOpen},
		{Type: TokenString, Raw: []byte("world$")},
		{Type: TokenParenthesesClose},
	}))
	stmt := NewStmtExecutor(scanner, NewContext())
	if err := stmt.Execute(); err != nil {
		t.Fatal(err)
	}
	if !(stmt.value.Type == ValueBool && stmt.value.Value.(bool)) {
		t.Fatal("match error")
	}
}

func TestString_sub(t *testing.T) {
	// "hello world".sub([0, 4])
	scanner := NewTokenScanner(newLexMock([]*Token{
		{Type: TokenString, Raw: []byte("hello world")},
		{Type: TokenDot},
		{Type: TokenIdentifier, Raw: []byte("sub")},
		{Type: TokenParenthesesOpen},
		{Type: TokenBracketsOpen},
		{Type: TokenNumber, Raw: []byte{'0'}},
		{Type: TokenComma},
		{Type: TokenNumber, Raw: []byte{'4'}},
		{Type: TokenBracketsClose},
		{Type: TokenParenthesesClose},
	}))
	stmt := NewStmtExecutor(scanner, NewContext())
	if err := stmt.Execute(); err != nil {
		t.Fatal(err)
	}
	if !(stmt.value.Type == ValueString && bytes.Equal(stmt.value.Value.(String).Bytes(), []byte("hello"))) {
		t.Fatal("match error")
	}
}
