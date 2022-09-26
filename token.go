package djson

type TokenType int

const (
	TokenMapOpen        = TokenType(iota) // {
	TokenMapClose                         // }
	TokenWhitespace                       // '\n', '\t', ' '
	TokenVariable                         // variable
	TokenBlockStart                       //  [  <  (
	TokenBlockEnd                         //  ]  >  }
	TokenBlockSeperator                   //  , :
	TokenAssignation                      // =
	TokenComparation                      // == < > <= >=
	TokenLogicOperator                    // || &&
	TokenOperator                         // + - * / %
	TokenEOF                              // eof
	TokenNumber                           // number
	TokenKeyword                          // keyword
	TokenString                           // string
	TokenBoolean                          // boolean
	TokenNull                             // null
)

type Token struct {
	Type     TokenType
	Raw      []byte
	Row, Col int
}

func (token *Token) IsTemplateStart() bool {
	return token.Type == TokenComparation && token.Raw[0] == '<'
}

func (token *Token) IsTemplateEnd() bool {
	return token.Type == TokenComparation && token.Raw[0] == '>'
}
