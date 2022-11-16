package djson

import (
	"bytes"
	"strings"
	"testing"
)

func TestLexer_number(t *testing.T) {
	data := []struct {
		data      string
		val       []byte
		typ       TokenType
		shoulderr bool
	}{
		{data: "1", val: []byte{'1'}, typ: TokenNumber},
		{data: "1.24 ", val: []byte("1.24"), typ: TokenNumber},
		{data: "12343\n", val: []byte("12343"), typ: TokenNumber},
		{data: "124a\n", val: []byte("124"), typ: TokenNumber, shoulderr: true},
		{data: "124[\n", val: []byte("124"), typ: TokenNumber},
		{data: "124{\n", val: []byte("124"), typ: TokenNumber},
		{data: "124]\n", val: []byte("124"), typ: TokenNumber},
		{data: "124{\n", val: []byte("124"), typ: TokenNumber},
		{data: "124<\n", val: []byte("124"), typ: TokenNumber},
		{data: "124>\n", val: []byte("124"), typ: TokenNumber},
		{data: "124,\n", val: []byte("124"), typ: TokenNumber},
		{data: "124=\n", val: []byte("124"), typ: TokenNumber},
	}
	for _, item := range data {
		g := NewLexer(bytes.NewBuffer([]byte(item.data)), 32)
		g.state = stateStart
		var token Token
		if err := g.NextToken(&token); err != nil {
			if item.shoulderr {
				continue
			}
			t.Fatal(err)
		}
		if token.Type != item.typ || !bytes.Equal(item.val, token.Raw) {
			t.Fatal("error occurred")
		}
		t.Logf("token: %#v\n", token)
	}
}

func TestLexer_string(t *testing.T) {
	data := []struct {
		data      string
		val       []byte
		typ       TokenType
		shoulderr bool
	}{
		{data: "\"1\"", val: []byte{'1'}, typ: TokenString},
		{data: "\"124\"", val: []byte("124"), typ: TokenString},
		{data: "\"hello world", shoulderr: true, typ: TokenString},
		{data: "\"hello world\nhello\"", val: []byte("hello world\nhello"), typ: TokenString},
		{data: "\"hello \\\"world\nhello\"", val: []byte("hello \\\"world\nhello"), typ: TokenString},
	}
	for i, item := range data {
		g := NewLexer(bytes.NewBuffer([]byte(item.data)), 32)
		g.state = stateStart
		var token Token
		if err := g.NextToken(&token); err != nil {
			if item.shoulderr {
				continue
			}
			t.Fatal(err)
		}
		if token.Type != item.typ || !bytes.Equal(item.val, token.Raw) {
			t.Fatalf("error occurred at %d", i)
		}
		t.Logf("token: %#v\n", token)
	}
}

func TestLexer_range(t *testing.T) {
	data := "[1 ... 10]"
	g := NewLexer(strings.NewReader(data), 16)
	tokens := []*Token{
		{Type: TokenBracketsOpen},
		{Type: TokenNumber, Raw: []byte{'1'}},
		{Type: TokenRange},
		{Type: TokenNumber, Raw: []byte{'1', '0'}},
		{Type: TokenBracketsClose},
	}
	var token Token
	for i, to := range tokens {
		if err := g.NextToken(&token); err != nil {
			t.Fatal(err)
		}
		if to.Type != token.Type {
			t.Fatalf("token type error at %d", i)
		}
	}
}

func TestLexer_bool(t *testing.T) {
	data := []struct {
		data      string
		typ       TokenType
		shoulderr bool
	}{
		{data: "true", typ: TokenTrue},
		{data: "false", typ: TokenFalse},
		{data: "true)", typ: TokenTrue},
		{data: "false]", typ: TokenFalse},
		{data: "false>", typ: TokenFalse},
		{data: "falsed", typ: TokenIdentifier},
		{data: "fals>", typ: TokenIdentifier},
	}
	for _, item := range data {
		g := NewLexer(bytes.NewBuffer([]byte(item.data)), 32)
		g.state = stateStart
		var token Token
		if err := g.NextToken(&token); err != nil {
			if item.shoulderr {
				continue
			}
			t.Fatal(err)
		}
		if token.Type != item.typ {
			t.Fatal("error occurred")
		}
		t.Logf("token: %#v\n", token)
	}
}

func TestLexer_compose(t *testing.T) {
	data := `
data = {
    "string": "123",
    "int": 123,
    "float": 1.23,
    "bool": true,
}.set(k == "string" => v + "_new")
# hello world
1 != 2 && true || false
`
	res := []struct {
		typ      TokenType
		raw      []byte
		row, col int
	}{
		{typ: TokenIdentifier, raw: []byte("data"), row: 1, col: 0},
		{typ: TokenAssignation, row: 1, col: 5},
		{typ: TokenBraceOpen, row: 1, col: 7},

		{typ: TokenString, raw: []byte("string"), row: 2, col: 4},
		{typ: TokenColon, row: 2, col: 12},
		{typ: TokenString, raw: []byte("123"), row: 2, col: 14},
		{typ: TokenComma, row: 2, col: 19},

		{typ: TokenString, raw: []byte("int"), row: 3, col: 4},
		{typ: TokenColon, row: 3, col: 9},
		{typ: TokenNumber, raw: []byte("123"), row: 3, col: 11},
		{typ: TokenComma, row: 3, col: 14},

		{typ: TokenString, raw: []byte("float"), row: 4, col: 4},
		{typ: TokenColon, row: 4, col: 11},
		{typ: TokenNumber, raw: []byte("1.23"), row: 4, col: 13},
		{typ: TokenComma, row: 4, col: 17},

		{typ: TokenString, raw: []byte("bool"), row: 5, col: 4},
		{typ: TokenColon, row: 5, col: 10},
		{typ: TokenTrue, row: 5, col: 12},
		{typ: TokenComma, row: 5, col: 16},

		{typ: TokenBraceClose, row: 6, col: 0},
		{typ: TokenDot, row: 6, col: 1},
		{typ: TokenIdentifier, raw: []byte("set"), row: 6, col: 2},
		{typ: TokenParenthesesOpen, row: 6, col: 5},
		{typ: TokenIdentifier, raw: []byte("k"), row: 6, col: 6},
		{typ: TokenEqual, row: 6, col: 8},
		{typ: TokenString, raw: []byte("string"), row: 6, col: 11},
		{typ: TokenReduction, row: 6, col: 20},
		{typ: TokenIdentifier, raw: []byte("v"), row: 6, col: 23},
		{typ: TokenAddition, row: 6, col: 25},
		{typ: TokenString, raw: []byte("_new"), row: 6, col: 27},
		{typ: TokenParenthesesClose, row: 6, col: 33},

		{typ: TokenComment, row: 7, col: 0},

		{typ: TokenNumber, row: 8, col: 0},
		{typ: TokenNotEqual, row: 8, col: 2},
		{typ: TokenNumber, row: 8, col: 5},
		{typ: TokenAnd, row: 8, col: 7},
		{typ: TokenTrue, row: 8, col: 10},
		{typ: TokenOr, row: 8, col: 15},
		{typ: TokenFalse, row: 8, col: 18},
	}
	g := NewLexer(bytes.NewBuffer([]byte(data)), 128)
	g.state = stateStart
	for i := 0; i < 100; i++ {
		var token Token
		if err := g.NextToken(&token); err != nil {
			t.Fatal(err)
		}
		if token.Type == TokenEOF {
			break
		}
		if res[i].typ != token.Type || !bytes.Equal(res[i].raw, token.Raw) {
			t.Fatalf("type or raw error at %d", i)
		}
		if res[i].row != 0 && res[i].col != 0 && (res[i].row != token.Row || res[i].col != token.Col) {
			t.Fatalf("col or row error at %d", i)
		}
	}
}
func BenchmarkLexer_NextToken(n *testing.B) {
	data := `
data = {
    "string": "123",
    "int": 123,
    "float": 1.23,
    "bool": true,
}.set(k == "string" => v + "_new")
# hello world
1 != 2 && true || false`
	for i := 0; i < n.N; i++ {
		g := NewLexer(bytes.NewBuffer([]byte(data)), 128)
		var token Token
		for token.Type != TokenEOF {
			if err := g.NextToken(&token); err != nil {
				n.Fatal(err)
			}
		}
	}
}
