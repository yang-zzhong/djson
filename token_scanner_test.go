package djson

import "testing"

func TestTokenScanner_next(t *testing.T) {
	lex := NewLexMock([]*Token{
		{Type: TokenIdentifier, Raw: []byte("data"), Row: 1, Col: 0},
		{Type: TokenAssignation, Row: 1, Col: 5},
		{Type: TokenBraceOpen, Row: 1, Col: 7},

		{Type: TokenString, Raw: []byte("\"string\""), Row: 2, Col: 4},
		{Type: TokenColon, Row: 2, Col: 12},
		{Type: TokenString, Raw: []byte("\"123\""), Row: 2, Col: 14},
		{Type: TokenComma, Row: 2, Col: 19},

		{Type: TokenString, Raw: []byte("\"int\""), Row: 3, Col: 4},
		{Type: TokenColon, Row: 3, Col: 9},
		{Type: TokenNumber, Raw: []byte("123"), Row: 3, Col: 11},
		{Type: TokenComma, Row: 3, Col: 14},

		{Type: TokenString, Raw: []byte("\"float\""), Row: 4, Col: 4},
		{Type: TokenColon, Row: 4, Col: 11},
		{Type: TokenNumber, Raw: []byte("1.23"), Row: 4, Col: 13},
		{Type: TokenComma, Row: 4, Col: 17},

		{Type: TokenString, Raw: []byte("\"bool\""), Row: 5, Col: 4},
		{Type: TokenColon, Row: 5, Col: 10},
		{Type: TokenTrue, Row: 5, Col: 12},
		{Type: TokenComma, Row: 5, Col: 16},

		{Type: TokenBraceClose, Row: 6, Col: 0},
		{Type: TokenDot, Row: 6, Col: 1},
		{Type: TokenIdentifier, Raw: []byte("set"), Row: 6, Col: 2},
		{Type: TokenParenthesesOpen, Row: 6, Col: 5},
		{Type: TokenIdentifier, Raw: []byte("k"), Row: 6, Col: 6},
		{Type: TokenEqual, Row: 6, Col: 8},
		{Type: TokenString, Raw: []byte("\"string\""), Row: 6, Col: 11},
		{Type: TokenReduction, Row: 6, Col: 20},
		{Type: TokenIdentifier, Raw: []byte("v"), Row: 6, Col: 23},
		{Type: TokenAddition, Row: 6, Col: 25},
		{Type: TokenString, Raw: []byte("\"_new\""), Row: 6, Col: 27},
		{Type: TokenParenthesesClose, Row: 6, Col: 33},
		{Type: TokenEOF, Row: 7, Col: 0},
	})
	scanner := NewTokenRecordScanner(NewTokenScanner(lex)).(*tokenRecordScanner)
	scanner.PushEnds(TokenEOF)
	for i := 0; i < 5; i++ {
		if _, err := scanner.Scan(); err != nil {
			t.Fatal(err)
		}
		scanner.Forward()
	}
	scanner.Reset()
	scanner.Scan()
	if scanner.Token().Type != TokenIdentifier {
		t.Fatal("token type not match")
	}
	scanner.Forward()
	if _, err := scanner.Scan(); err != nil {
		t.Fatal(err)
	}
	if scanner.Token().Type != TokenAssignation {
		t.Fatal("token type not match after forward")
	}
	scanner.PushEnds(TokenAddition)
	if scanner.tokenScanner.(*tokenScanner).endsWhen.when[TokenAddition] != 1 {
		t.Fatal("push ends error")
	}
	scanner.PopEnds(TokenAddition)
	if scanner.tokenScanner.(*tokenScanner).endsWhen.when[TokenAddition] != 0 {
		t.Fatal("pop ends error")
	}
}
