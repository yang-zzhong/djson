package djson

import (
	"io"
	"strings"
	"testing"
)

func TestLexerV2_NextToken(t *testing.T) {
	r := strings.NewReader(`data = {
    "string": "123",
    "int": 123,
    "float": 1.23,
    "bool": true,
}.set(k == "string" => v + "_new")
# hello world
1 != 2 && true || false`)
	r.Seek(0, io.SeekStart)
	var token Token
	g := NewLexerV2(r, 512)
	for token.Type != TokenEOF {
		if err := g.NextToken(&token); err != nil {
			t.Fatal(err)
		}
	}
}

func BenchmarkLexerV2_NextToken(n *testing.B) {
	r := strings.NewReader(`data = {
    "string": "123",
    "int": 123,
    "float": 1.23,
    "bool": true,
}.set(k == "string" => v + "_new")
# hello world
1 != 2 && true || false`)
	for i := 0; i < n.N; i++ {
		r.Seek(0, io.SeekStart)
		var token Token
		g := NewLexerV2(r, 512)
		for token.Type != TokenEOF {
			if err := g.NextToken(&token); err != nil {
				n.Fatal(err)
			}
		}
	}
}
