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
		{data: "falsed", val: []byte("falsed"), typ: TokenString},
		{data: "fals>", val: []byte("fals"), typ: TokenString},
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
