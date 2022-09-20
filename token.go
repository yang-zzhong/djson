package djson

type TokenType int

const (
	TokenMapOpen        = TokenType(iota) // {
	TokenMapClose                         // }
	TokenWhitespace                       // '\n', '\t', ' '
	TokenVariable                         // variable
	TokenBlockSeperator                   //  [ ] < > ( ) , " :
	TokenAssignation                      // =
	TokenComparation                      // == < > <= >=
	TokenLogicOperator                    // || &&
	TokenOperator                         // + - * / %
	TokenEOF                              // eof
	TokenNumber                           // number
	TokenString                           // string
	TokenBoolean                          // boolean
)
