package djson

import "testing"

func TestMapRange(t *testing.T) {
	// [1 ... 10].map(i + v)
	scanner := NewTokenScanner(newLexMock([]*Token{
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
	s := NewStmt(scanner, NewContext())
	if err := s.Execute(); err != nil {
		t.Fatal(err)
	}
	if s.value.Type != ValueArray {
		t.Fatal("type error")
	}
	arr := s.value.Value.(Array)
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
