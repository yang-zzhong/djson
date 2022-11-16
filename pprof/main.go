package main

import (
	"bytes"
	"djson"
	"flag"
	"os"
	"runtime/pprof"
)

func main() {
	var old bool
	flag.BoolVar(&old, "old", false, "")
	flag.Parse()
	data := `
data = {
    "string": "123",
    "int": 123,
    "float": 1.23,
    "bool": true,
}.set(k == "string" => v + "_new")
# hello world
1 != 2 && true || false`
	var g djson.Lexer
	g = djson.NewMatcherLexer(bytes.NewBuffer([]byte(data)), 128)
	if old {
		g = djson.NewLexer(bytes.NewBuffer([]byte(data)), 128)
	}
	pprof.StartCPUProfile(os.Stdout)
	defer pprof.StopCPUProfile()
	for i := 0; i < 100000; i++ {
		g.(djson.ReplaceSourcer).ReplaceSource(bytes.NewBuffer([]byte(data)), 128)
		var token djson.Token
		for token.Type != djson.TokenEOF {
			if err := g.NextToken(&token); err != nil {
				panic(err)
			}
		}
	}
}
