package djson

import (
	"bytes"
	"testing"
)

func TestTokenGetter_number(t *testing.T) {
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
		g := NewTokenGetter(bytes.NewBuffer([]byte(item.data)), 32)
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

func TestTokenGetter_string(t *testing.T) {
	data := []struct {
		data      string
		val       []byte
		typ       TokenType
		shoulderr bool
	}{
		{data: "\"1\"", val: []byte{'"', '1', '"'}, typ: TokenString},
		{data: "\"124\"", val: []byte("\"124\""), typ: TokenString},
		{data: "\"hello world", shoulderr: true, typ: TokenString},
		{data: "\"hello world\nhello\"", val: []byte("\"hello world\nhello\""), typ: TokenString},
		{data: "\"hello \\\"world\nhello\"", val: []byte("\"hello \\\"world\nhello\""), typ: TokenString},
	}
	for _, item := range data {
		g := NewTokenGetter(bytes.NewBuffer([]byte(item.data)), 32)
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

func TestTokenGetter_bool(t *testing.T) {
	data := []struct {
		data      string
		val       []byte
		typ       TokenType
		shoulderr bool
	}{
		{data: "true", val: []byte("true"), typ: TokenBoolean},
		{data: "false", val: []byte("false"), typ: TokenBoolean},
		{data: "true)", val: []byte("true"), typ: TokenBoolean},
		{data: "false]", val: []byte("false"), typ: TokenBoolean},
		{data: "false>", val: []byte("false"), typ: TokenBoolean},
		{data: "falsed", val: []byte("falsed"), typ: TokenVariable},
		{data: "fals>", val: []byte("fals"), typ: TokenVariable},
	}
	for _, item := range data {
		g := NewTokenGetter(bytes.NewBuffer([]byte(item.data)), 32)
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

func TestTokenGetter_compose(t *testing.T) {
	data := `
{
    "string": "123",
    "int": 123,
    "float": 1.23,
    "bool": true,
}<k, v>(k == "hello")
`
	res := []struct {
		typ      TokenType
		raw      []byte
		row, col int
	}{
		{typ: TokenWhitespace, raw: []byte{'\n'}, row: 0, col: 0},
		{typ: TokenBlockSeperator, raw: []byte{'{'}, row: 1, col: 0},
		{typ: TokenWhitespace, raw: []byte{'\n', ' ', ' ', ' ', ' '}, row: 1, col: 1},

		{typ: TokenString, raw: []byte("\"string\""), row: 2, col: 4},
		{typ: TokenBlockSeperator, raw: []byte{':'}, row: 2, col: 12},
		{typ: TokenWhitespace, raw: []byte{' '}, row: 2, col: 13},
		{typ: TokenString, raw: []byte("\"123\""), row: 2, col: 14},
		{typ: TokenBlockSeperator, raw: []byte(","), row: 2, col: 19},

		{typ: TokenWhitespace, raw: []byte{'\n', ' ', ' ', ' ', ' '}, row: 2, col: 20},

		{typ: TokenString, raw: []byte("\"int\"")},
		{typ: TokenBlockSeperator, raw: []byte{':'}},
		{typ: TokenWhitespace, raw: []byte{' '}},
		{typ: TokenNumber, raw: []byte("123")},
		{typ: TokenBlockSeperator, raw: []byte(",")},

		{typ: TokenWhitespace, raw: []byte{'\n', ' ', ' ', ' ', ' '}},

		{typ: TokenString, raw: []byte("\"float\"")},
		{typ: TokenBlockSeperator, raw: []byte{':'}},
		{typ: TokenWhitespace, raw: []byte{' '}},
		{typ: TokenNumber, raw: []byte("1.23")},
		{typ: TokenBlockSeperator, raw: []byte(",")},

		{typ: TokenWhitespace, raw: []byte{'\n', ' ', ' ', ' ', ' '}},

		{typ: TokenString, raw: []byte("\"bool\"")},
		{typ: TokenBlockSeperator, raw: []byte{':'}},
		{typ: TokenWhitespace, raw: []byte{' '}},
		{typ: TokenBoolean, raw: []byte("true")},
		{typ: TokenBlockSeperator, raw: []byte(",")},

		{typ: TokenWhitespace, raw: []byte{'\n'}},

		{typ: TokenBlockSeperator, raw: []byte{'}'}},
		{typ: TokenOperator, raw: []byte{'<'}},
		{typ: TokenVariable, raw: []byte{'k'}},
		{typ: TokenBlockSeperator, raw: []byte{','}},
		{typ: TokenWhitespace, raw: []byte{' '}},
		{typ: TokenVariable, raw: []byte{'v'}},
		{typ: TokenOperator, raw: []byte{'>'}},
		{typ: TokenBlockSeperator, raw: []byte{'('}},
		{typ: TokenVariable, raw: []byte{'k'}},
		{typ: TokenWhitespace, raw: []byte{' '}},
		{typ: TokenOperator, raw: []byte{'=', '='}},
		{typ: TokenWhitespace, raw: []byte{' '}},
		{typ: TokenString, raw: []byte("\"hello\"")},
		{typ: TokenBlockSeperator, raw: []byte{')'}},
		{typ: TokenWhitespace, raw: []byte{'\n'}},
	}
	g := NewTokenGetter(bytes.NewBuffer([]byte(data)), 128)
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
			t.Fatal("type or raw error")
		}
		if res[i].row != 0 && res[i].col != 0 && (res[i].row != token.Row || res[i].col != token.Col) {
			t.Fatal("col or row error")
		}
	}
}
