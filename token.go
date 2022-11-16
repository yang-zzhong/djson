package djson

type TokenType int

const (
	TokenBraceOpen        = TokenType(iota) // {
	TokenBraceClose                         // }
	TokenBracketsOpen                       //  [
	TokenBracketsClose                      //  ]
	TokenParenthesesOpen                    //  (
	TokenParenthesesClose                   //  )
	TokenAssignation                        // =
	TokenEqual                              // ==
	TokenNotEqual                           // !=
	TokenGreateThan                         // >
	TokenLessThan                           // <
	TokenGreateThanEqual                    // >=
	TokenLessThanEqual                      // <=
	TokenOr                                 // ||
	TokenAnd                                // &&
	TokenSemicolon                          // ;
	TokenAddition                           // +
	TokenMinus                              // -
	TokenMultiplication                     // *
	TokenDevision                           // /
	TokenColon                              // :
	TokenComma                              // ,
	TokenDot                                // .
	TokenEOF                                // eof
	TokenExclamation                        // !
	TokenComment                            // comment
	TokenNull                               // null
	TokenTrue                               // true
	TokenFalse                              // false
	TokenReduction                          // =>
	TokenNumber                             // number
	TokenString                             // string
	TokenRange                              // ... // [1 ... 10].map({"key": "" + v + "_x"})
	TokenIdentifier                         // identifier
	TokenWhitespace
)

type Token struct {
	Type     TokenType
	Raw      []byte
	Row, Col int
}

func (t Token) Skip() bool {
	return t.Type == TokenWhitespace
}

func exclodeRawToken(tokenType TokenType) bool {
	tokens := []TokenType{
		TokenBraceOpen,
		TokenBraceClose,
		TokenBracketsOpen,
		TokenBracketsClose,
		TokenParenthesesOpen,
		TokenParenthesesClose,
		TokenAssignation,
		TokenEqual,
		TokenGreateThan,
		TokenLessThan,
		TokenGreateThanEqual,
		TokenLessThanEqual,
		TokenOr,
		TokenAnd,
		TokenAddition,
		TokenMinus,
		TokenMultiplication,
		TokenDevision,
		TokenDot,
		TokenEOF,
		TokenNull,
		TokenTrue,
		TokenFalse,
		TokenColon,
		TokenComma,
		TokenReduction,
		TokenSemicolon,
	}
	for _, t := range tokens {
		if t == tokenType {
			return true
		}
	}
	return false
}

var (
	CharsMatchers = []TokenMatcher{
		CharsMatcher([]byte{'{'}, TokenBraceOpen),
		CharsMatcher([]byte{'}'}, TokenBraceClose),                // }
		CharsMatcher([]byte{'['}, TokenBracketsOpen),              // [
		CharsMatcher([]byte{']'}, TokenBracketsClose),             // ]
		CharsMatcher([]byte{'('}, TokenParenthesesOpen),           // (
		CharsMatcher([]byte{')'}, TokenParenthesesClose),          // )
		CharsMatcher([]byte{'='}, TokenAssignation),               // =
		CharsMatcher([]byte{'=', '='}, TokenEqual),                // ==
		CharsMatcher([]byte{'!', '='}, TokenNotEqual),             // !=
		CharsMatcher([]byte{'>'}, TokenGreateThan),                // >
		CharsMatcher([]byte{'<'}, TokenLessThan),                  // <
		CharsMatcher([]byte{'>', '='}, TokenGreateThanEqual),      // >=
		CharsMatcher([]byte{'<', '='}, TokenLessThanEqual),        // <=
		CharsMatcher([]byte{'|', '|'}, TokenOr),                   // ||
		CharsMatcher([]byte{'&', '&'}, TokenAnd),                  // &&
		CharsMatcher([]byte{';'}, TokenSemicolon),                 // ;
		CharsMatcher([]byte{'+'}, TokenAddition),                  // +
		CharsMatcher([]byte{'-'}, TokenMinus),                     // -
		CharsMatcher([]byte{'*'}, TokenMultiplication),            // *
		CharsMatcher([]byte{'/'}, TokenDevision),                  // /
		CharsMatcher([]byte{':'}, TokenColon),                     // :
		CharsMatcher([]byte{','}, TokenComma),                     // ,
		CharsMatcher([]byte{'.'}, TokenDot),                       // .
		CharsMatcher([]byte{'!'}, TokenExclamation),               // !
		CharsMatcher([]byte{'n', 'u', 'l', 'l'}, TokenNull),       // null
		CharsMatcher([]byte{'t', 'r', 'u', 'e'}, TokenTrue),       // true
		CharsMatcher([]byte{'f', 'a', 'l', 's', 'e'}, TokenFalse), // false
		CharsMatcher([]byte{'=', '>'}, TokenReduction),            // =>
		CharsMatcher([]byte{'.', '.', '.'}, TokenRange),           // ... // [1 ... 10].map({"key": "" + v + "_x"})
	}
)
