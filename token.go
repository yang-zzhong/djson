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
)

type Token struct {
	Type     TokenType
	Raw      []byte
	Row, Col int
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
	}
	for _, t := range tokens {
		if t == tokenType {
			return true
		}
	}
	return false
}
