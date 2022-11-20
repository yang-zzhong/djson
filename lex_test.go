package djson

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestMatcherLexer_number(t *testing.T) {
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
	for i, item := range data {
		g := NewLexer(bytes.NewBuffer([]byte(item.data)), 32)
		var token Token
		if err := g.NextToken(&token); err != nil {
			if item.shoulderr {
				continue
			}
			t.Fatalf("error at %d: %s", i, err.Error())
		}
		if token.Type != item.typ || !bytes.Equal(item.val, token.Raw) {
			t.Fatalf("error occurred at %d", i)
		}
		t.Logf("token: %#v\n", token)
	}
}

func TestMatcherLexer_string(t *testing.T) {
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

func TestMatcherLexer_range(t *testing.T) {
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

func TestMatcherLexer_bool(t *testing.T) {
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

func TestMatcherLexer_compose(t *testing.T) {
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
		{typ: TokenIdentifier, raw: []byte("data"), row: 2, col: 1},
		{typ: TokenAssignation, row: 2, col: 6},
		{typ: TokenBraceOpen, row: 2, col: 8},

		{typ: TokenString, raw: []byte("string"), row: 3, col: 5},
		{typ: TokenColon, row: 3, col: 13},
		{typ: TokenString, raw: []byte("123"), row: 3, col: 15},
		{typ: TokenComma, row: 3, col: 20},

		{typ: TokenString, raw: []byte("int"), row: 4, col: 5},
		{typ: TokenColon, row: 4, col: 10},
		{typ: TokenNumber, raw: []byte("123"), row: 4, col: 12},
		{typ: TokenComma, row: 4, col: 15},

		{typ: TokenString, raw: []byte("float"), row: 5, col: 5},
		{typ: TokenColon, row: 5, col: 12},
		{typ: TokenNumber, raw: []byte("1.23"), row: 5, col: 14},
		{typ: TokenComma, row: 5, col: 18},

		{typ: TokenString, raw: []byte("bool"), row: 6, col: 5},
		{typ: TokenColon, row: 6, col: 11},
		{typ: TokenTrue, row: 6, col: 13},
		{typ: TokenComma, row: 6, col: 17},

		{typ: TokenBraceClose, row: 7, col: 1},
		{typ: TokenDot, row: 7, col: 2},
		{typ: TokenIdentifier, raw: []byte("set"), row: 7, col: 3},
		{typ: TokenParenthesesOpen, row: 7, col: 6},
		{typ: TokenIdentifier, raw: []byte("k"), row: 7, col: 7},
		{typ: TokenEqual, row: 7, col: 9},
		{typ: TokenString, raw: []byte("string"), row: 7, col: 12},
		{typ: TokenReduction, row: 7, col: 21},
		{typ: TokenIdentifier, raw: []byte("v"), row: 7, col: 24},
		{typ: TokenAddition, row: 7, col: 26},
		{typ: TokenString, raw: []byte("_new"), row: 7, col: 28},
		{typ: TokenParenthesesClose, row: 7, col: 34},

		{typ: TokenComment, row: 8, col: 1, raw: []byte(" hello world")},

		{typ: TokenNumber, row: 9, col: 1, raw: []byte("1")},
		{typ: TokenNotEqual, row: 9, col: 3},
		{typ: TokenNumber, row: 9, col: 6, raw: []byte("2")},
		{typ: TokenAnd, row: 9, col: 8},
		{typ: TokenTrue, row: 9, col: 11},
		{typ: TokenOr, row: 9, col: 16},
		{typ: TokenFalse, row: 9, col: 19},
	}
	g := NewLexer(bytes.NewBuffer([]byte(data)), 128)
	for i := 0; i < 100; i++ {
		var token Token
		if i == 39 {
			t.Logf("#%d", i)
		}
		if err := g.NextToken(&token); err != nil {
			t.Fatalf("token error at %d: %s", i, err.Error())
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

func BenchmarkMatcherLexer_NextToken(n *testing.B) {
	f, err := os.Open("./testdata/bench_lexer.djson")
	if err != nil {
		n.Fatal(err)
	}
	for i := 0; i < n.N; i++ {
		f.Seek(0, io.SeekStart)
		var token Token
		g := NewLexer(f, 512)
		for token.Type != TokenEOF {
			if err := g.NextToken(&token); err != nil {
				n.Fatal(err)
			}
		}
	}
}
