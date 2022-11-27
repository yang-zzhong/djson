package djson

import (
	"strings"
	"testing"
)

func TestMapRange(t *testing.T) {
	// [1 ... 10].map(i + v)
	scanner := NewTokenScanner(NewLexMock([]*Token{
		{Type: TokenBracketsOpen},
		{Type: TokenNumber, Raw: []byte{'1'}},
		{Type: TokenRange},
		{Type: TokenNumber, Raw: []byte{'1', '0'}},
		{Type: TokenBracketsClose},
		{Type: TokenDot},
		{Type: TokenIdentifier, Raw: []byte{'m', 'a', 'p'}},
		{Type: TokenParenthesesOpen},
		{Type: TokenIdentifier, Raw: []byte{'i'}},
		{Type: TokenAddition},
		{Type: TokenIdentifier, Raw: []byte{'v'}},
		{Type: TokenParenthesesClose},
	}))
	s := NewStmtExecutor(scanner, NewContext())
	if err := s.Execute(); err != nil {
		t.Fatal(err)
	}
	if s.value.Type != ValueArray {
		t.Fatal("type error")
	}
	arr := s.value.Value.(ItemEachable)
	arr.Each(func(i int, val Value) bool {
		if val.Type != ValueInt {
			t.Fatal("elem type error")
		}
		v, _ := val.Value.(Inter).Int()
		if i+i+1 != int(v) {
			t.Fatal("elem value error")
		}
		return true
	})
}

func TestRange_parallelMap(t *testing.T) {
	data := `
# range parallel map test

[1 ... 100].parallel({
    "key": i,
    "val": v
})
    `
	stmt := NewStmtExecutor(NewTokenScanner(NewLexer(strings.NewReader(data), 128)), NewContext())
	if err := stmt.Execute(); err != nil {
		t.Fatal(err)
	}
	t.Log(stmt.Value())
}

func BenchmarkRange_parallelMap(b *testing.B) {
	data := `
# range parallel map test

[1 ... 10000].parallel({
    "key": i,
    "val": v
})
    `
	for i := 0; i < b.N; i++ {
		stmt := NewStmtExecutor(NewTokenScanner(NewLexer(strings.NewReader(data), 128)), NewContext())
		if err := stmt.Execute(); err != nil {
			b.Fatal(err)
		}
	}
}
