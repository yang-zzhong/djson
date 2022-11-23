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
	TokenWhitespace                         // whitespace
	TokenMod                                // %
	TokenExit                               // exit
	TokenReturn                             // return
)

type Token struct {
	Type     TokenType
	Raw      []byte
	Row, Col int
}

func (t Token) Skip() bool {
	return t.Type == TokenWhitespace
}

func (t Token) Name() string {
	return map[TokenType]string{
		TokenBraceOpen:        "BraceOpen",
		TokenBraceClose:       "BraceClose",
		TokenBracketsOpen:     "BracketsOpen",     //  [
		TokenBracketsClose:    "BracketsClose",    //  ]
		TokenParenthesesOpen:  "ParenthesesOpen",  //  (
		TokenParenthesesClose: "ParenthesesClose", //  )
		TokenAssignation:      "Assignation",      // =
		TokenEqual:            "Equal",            // ==
		TokenNotEqual:         "NotEqual",         // !=
		TokenGreateThan:       "GreateThan",       // >
		TokenLessThan:         "LessThan",         // <
		TokenGreateThanEqual:  "GreateThanEqual",  // >=
		TokenLessThanEqual:    "LessThanEqual",    // <=
		TokenOr:               "Or",               // ||
		TokenAnd:              "And",              // &&
		TokenSemicolon:        "Semicolon",        // ;
		TokenAddition:         "Addition",         // +
		TokenMinus:            "Minus",            // -
		TokenMultiplication:   "Multiplication",   // *
		TokenDevision:         "Devision",         // /
		TokenColon:            "Colon",            // :
		TokenComma:            "Comma",            // ,
		TokenDot:              "Dot",              // .
		TokenEOF:              "EOF",              // eof
		TokenExclamation:      "Exclamation",      // !
		TokenComment:          "Comment",          // comment
		TokenNull:             "Null",             // null
		TokenTrue:             "True",             // true
		TokenFalse:            "False",            // false
		TokenReduction:        "Reduction",        // =>
		TokenNumber:           "Number",           // number
		TokenMod:              "Mod",
		TokenString:           "String",     // string
		TokenRange:            "Range",      // ... // [1 ... 10].map({"key": "" + v + "_x"})
		TokenIdentifier:       "Identifier", // identifier
		TokenWhitespace:       "Whitespace",
	}[t.Type]
}
