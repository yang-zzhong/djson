package djson

import (
	"bytes"
	"errors"
	"io"
)

type parseState int

const (
	stateStart = parseState(iota)
	stateWhitespace
	stateIdentifier
	stateBlockSeperator
	stateString
	stateNumber
	stateVoid
	stateSemicolon
	stateEqStarted
	stateEq
	stateReduction
	stateExclamation
	stateComment
	stateLt
	stateNeq
	stateGt
	stateGte
	stateLte
	stateEOF
	stateTrue
	stateFalse
	stateNull
	stateAnd
	stateOr
	stateAndStarted
	stateOrStarted
	stateAdd
	stateMinus
	stateMultiple
	stateDevide
)

const (
	stashSize = 256
)

type lexer interface {
	NextToken(token *Token) error
}

type lexer_ struct {
	source        io.Reader
	row, col      int
	offset        int
	buf           []byte
	max           int
	state         parseState
	stash         []byte
	stashStartRow int
	stashStartCol int
	slashOpenned  bool
	stashOffset   int
	keyword       [][]byte
}

func NewLexer(source io.Reader, bufSize uint) *lexer_ {
	return &lexer_{
		source: source,
		row:    0, col: 0, offset: 0,
		buf: make([]byte, bufSize), max: 0,
		state: stateStart,
		stash: make([]byte, stashSize),
	}
}

func (g *lexer_) NextToken(token *Token) (err error) {
	var b byte
	var catched bool
	for {
		if g.state == stateEOF {
			token.Type = TokenEOF
			return
		}
		if err = g.nextChar(&b); err != nil {
			if errors.Is(err, io.EOF) {
				return g.matchEOF(token)
			}
			return
		}
		if b == 0 {
			continue
		}
		switch g.state {
		case stateStart:
			catched, err = g.matchStart(b, token)
		case stateWhitespace:
			catched, err = g.matchWhitespace(b, token)
		case stateIdentifier:
			catched, err = g.matchVariable(b, token)
		case stateBlockSeperator:
			catched, err = g.matchBlockSeperator(b, token)
		case stateString:
			catched, err = g.matchString(b, token)
		case stateVoid:
			catched, err = g.matchVoid(b, token)
		case stateNumber:
			catched, err = g.matchNumber(b, token)
		case stateAnd, stateEq, stateLte, stateGte, stateReduction:
			catched, err = g.matchComplete(b, token)
		case stateEqStarted:
			catched, err = g.matchLogicEq(b, token)
		case stateLt:
			catched, err = g.matchLogicLte(b, token)
		case stateGt:
			catched, err = g.matchLogicGte(b, token)
		case stateTrue, stateFalse, stateNull:
			catched, err = g.matchKeyword(b, token)
		case stateAndStarted:
			catched, err = g.matchAnd(b, token)
		case stateComment:
			catched, err = g.matchComment(b, token)
		case stateExclamation:
			catched, err = g.matchExclamation(b, token)
		}
		if b == '\n' {
			g.row++
			g.col = 0
		} else {
			g.col++
		}
		if catched || err != nil {
			return
		}
	}
}

func (g *lexer_) matchComplete(b byte, token *Token) (bool, error) {
	return g.shiftState(0, stateVoid, token), nil
}

func (g *lexer_) matchAnd(b byte, token *Token) (bool, error) {
	if b == '&' {
		g.state = stateAnd
		g.addToStash(b)
		return false, nil
	}
	return g.matchVoid(b, token)
}

func (g *lexer_) matchComment(b byte, token *Token) (bool, error) {
	if b == '\n' {
		return g.shiftState(b, stateVoid, token), nil
	}
	g.addToStash(b)
	return false, nil
}

func (g *lexer_) matchKeyword(b byte, token *Token) (bool, error) {
	for i := 0; i < len(g.keyword); i++ {
		if len(g.keyword[i]) == 0 {
			if matched, catched := g.matchSimple(b, token); matched {
				g.keyword = [][]byte{}
				return catched, nil
			}
			g.keyword = append(g.keyword[:i], g.keyword[i+1:]...)
			i--
			continue
		}
		if g.keyword[i][0] == b {
			g.keyword[i] = g.keyword[i][1:]
			g.addToStash(b)
			return false, nil
		}
		g.keyword = append(g.keyword[:i], g.keyword[i+1:]...)
		i--
	}
	g.state = stateIdentifier
	if matched, catched := g.matchSimple(b, token); matched {
		g.keyword = [][]byte{}
		return catched, nil
	}
	if g.isVarChar(b) {
		g.keyword = [][]byte{}
		g.addToStash(b)
		return false, nil
	}
	return false, &Error{
		Row:          g.row,
		Col:          g.col,
		CurrentBytes: []byte{b},
		Info:         UnexpectedChar,
	}
}

func (g *lexer_) matchExclamation(b byte, token *Token) (bool, error) {
	switch {
	case b == '=':
		g.state = stateNeq
		g.addToStash(b)
		return false, nil
	}
	return g.matchVoid(b, token)
}

func (g *lexer_) matchLogicEq(b byte, token *Token) (bool, error) {
	switch {
	case b == '=':
		g.state = stateEq
		g.addToStash(b)
		return false, nil
	case b == '>':
		g.state = stateReduction
		g.addToStash(b)
		return false, nil
	}
	return g.matchVoid(b, token)
}

func (g *lexer_) matchLogicGte(b byte, token *Token) (bool, error) {
	if b == '=' {
		g.state = stateGte
		g.addToStash(b)
		return false, nil
	}
	return g.matchVoid(b, token)
}

func (g *lexer_) matchLogicLte(b byte, token *Token) (bool, error) {
	if b == '=' {
		g.state = stateLte
		g.addToStash(b)
		return false, nil
	}
	return g.matchVoid(b, token)
}

func (g *lexer_) matchEOF(token *Token) error {
	err := &Error{
		Info: UnexpectedEOF,
		Row:  g.row,
		Col:  g.col,
	}
	if g.state != stateEOF && g.state != stateWhitespace {
		switch {
		case stateTrue == g.state && !bytes.Equal(g.stash[:4], []byte{'t', 'r', 'u', 'e'}):
			return err
		case stateFalse == g.state && !bytes.Equal(g.stash[:5], []byte{'f', 'a', 'l', 's', 'e'}):
			return err
		case stateNull == g.state && !bytes.Equal(g.stash[:4], []byte{'n', 'u', 'l', 'l'}):
			return err
		case stateString == g.state:
			return err
		}
		g.shiftState(' ', stateEOF, token)
		return nil
	}
	token.Type = TokenEOF
	return nil
}

func (g *lexer_) matchNumber(b byte, token *Token) (bool, error) {
	switch {
	case g.isNumber(b) || b == '.':
		g.addToStash(b)
		return false, nil
	}
	return g.matchNormal(b, token)
}

func (g *lexer_) matchString(b byte, token *Token) (bool, error) {
	switch b {
	case '"':
		if g.slashOpenned {
			g.addToStash(b)
			g.slashOpenned = false
			return false, nil
		}
		g.addToStash(b)
		return g.shiftState(0, stateVoid, token), nil
	case '\\':
		g.slashOpenned = true
		g.addToStash(b)
	default:
		g.addToStash(b)
	}
	return false, nil
}

func (g *lexer_) matchStart(b byte, token *Token) (bool, error) {
	if b == 0 {
		return false, nil
	}
	_, err := g.matchVoid(b, token)
	return false, err
}

func (g *lexer_) matchVoid(b byte, token *Token) (bool, error) {
	switch b {
	case 't':
		g.keyword = [][]byte{{'r', 'u', 'e'}}
		return g.shiftState(b, stateTrue, token), nil
	case 'f':
		g.keyword = [][]byte{{'a', 'l', 's', 'e'}}
		return g.shiftState(b, stateFalse, token), nil
	case 'n':
		g.keyword = [][]byte{{'u', 'l', 'l'}}
		return g.shiftState(b, stateNull, token), nil
	}
	return g.matchNormal(b, token)
}

func (g *lexer_) matchWhitespace(b byte, token *Token) (bool, error) {
	if g.isWhitespace(b) {
		return false, nil
	}
	return g.matchVoid(b, token)
}

func (g *lexer_) matchBlockSeperator(b byte, token *Token) (bool, error) {
	return g.matchVoid(b, token)
}

func (g *lexer_) matchVariable(b byte, token *Token) (bool, error) {
	switch {
	case g.isVarChar(b):
		g.addToStash(b)
		return false, nil
	case g.isWhitespace(b):
		return g.shiftState(b, stateWhitespace, token), nil
	}
	return g.matchNormal(b, token)
}

func (g *lexer_) matchNormal(b byte, token *Token) (bool, error) {
	if matched, catched := g.matchSimple(b, token); matched {
		return catched, nil
	}
	switch {
	case g.isNumber(b):
		return g.shiftState(b, stateNumber, token), nil
	case g.isAlpha(b):
		return g.shiftState(b, stateIdentifier, token), nil
	}
	return false, &Error{Row: g.row, Col: g.col, Info: UnexpectedChar, CurrentBytes: []byte{b}}
}

func (g *lexer_) matchSimple(b byte, token *Token) (matched bool, catched bool) {
	switch b {
	case '{', '}', '[', ']', ',', '(', ')', ':', '.':
		return true, g.shiftState(b, stateBlockSeperator, token)
	case '=':
		return true, g.shiftState(b, stateEqStarted, token)
	case '+':
		return true, g.shiftState(b, stateAdd, token)
	case '-':
		return true, g.shiftState(b, stateMinus, token)
	case '*':
		return true, g.shiftState(b, stateMultiple, token)
	case '/':
		return true, g.shiftState(b, stateDevide, token)
	case ';':
		return true, g.shiftState(b, stateSemicolon, token)
	case '<':
		return true, g.shiftState(b, stateLt, token)
	case '>':
		return true, g.shiftState(b, stateGt, token)
	case '&':
		return true, g.shiftState(b, stateAndStarted, token)
	case '|':
		return true, g.shiftState(b, stateOrStarted, token)
	case '#':
		return true, g.shiftState(b, stateComment, token)
	case '!':
		return true, g.shiftState(b, stateExclamation, token)
	case '"':
		return true, g.shiftState(b, stateString, token)
	}
	if g.isWhitespace(b) {
		return true, g.shiftState(b, stateWhitespace, token)
	}
	return false, false
}

func (g *lexer_) nextChar(b *byte) error {
	if g.offset == g.max {
		var err error
		for {
			if g.max, err = g.source.Read(g.buf); err != nil {
				return err
			} else if g.max == 0 {
				continue
			}
			g.offset = 0
			break
		}
	}
	*b = g.buf[g.offset]
	g.offset++
	return nil
}

func (g *lexer_) shiftState(b byte, state parseState, token *Token) bool {
	catched := false
	if g.stashOffset > 0 && g.state != stateStart && g.state != stateWhitespace {
		token.Type = func() TokenType {
			switch g.state {
			case stateIdentifier:
				return TokenIdentifier
			case stateBlockSeperator:
				seperators := map[byte]TokenType{
					'[': TokenBracketsOpen,
					']': TokenBracketsClose,
					'{': TokenBraceOpen,
					'}': TokenBraceClose,
					'(': TokenParenthesesOpen,
					')': TokenParenthesesClose,
					':': TokenColon,
					'.': TokenDot,
					',': TokenComma,
				}
				return seperators[g.stash[0]]
			case stateReduction:
				return TokenReduction
			case stateAdd:
				return TokenAddition
			case stateMinus:
				return TokenMinus
			case stateMultiple:
				return TokenMultiplication
			case stateDevide:
				return TokenDevision
			case stateEq:
				return TokenEqual
			case stateGt:
				return TokenGreateThan
			case stateGte:
				return TokenGreateThanEqual
			case stateLt:
				return TokenLessThan
			case stateLte:
				return TokenLessThanEqual
			case stateEqStarted:
				return TokenAssignation
			case stateNumber:
				return TokenNumber
			case stateString:
				return TokenString
			case stateTrue:
				return TokenTrue
			case stateFalse:
				return TokenFalse
			case stateAnd:
				return TokenAnd
			case stateOr:
				return TokenOr
			case stateExclamation:
				return TokenExclamation
			case stateComment:
				return TokenComment
			case stateNull:
				return TokenNull
			case stateNeq:
				return TokenNotEqual
			case stateSemicolon:
				return TokenSemicolon
			}
			return TokenType(100000)
		}()
		g.fillToken(token, b)
		catched = true
	}
	if g.state == stateWhitespace && state != stateWhitespace {
		g.stashStartRow = g.row
		g.stashStartCol = g.col
	}
	if stateWhitespace != state && b != 0 {
		g.addToStash(b)
	}
	g.state = state
	return catched
}

func (g *lexer_) addToStash(b byte) {
	if len(g.stash) <= g.stashOffset {
		stash := make([]byte, len(g.stash)*2)
		copy(stash, g.stash)
		g.stash = stash
	}
	g.stash[g.stashOffset] = b
	g.stashOffset++
}

func (g *lexer_) clearStash(next byte) {
	if len(g.stash) != stashSize {
		g.stash = make([]byte, stashSize)
	}
	g.stashOffset = 0
	g.stashStartRow = g.row
	g.stashStartCol = g.col
	if next == 0 {
		g.stashStartCol++
	}
}

func (g *lexer_) fillToken(token *Token, next byte) {
	if !exclodeRawToken(token.Type) {
		token.Raw = make([]byte, g.stashOffset)
		copy(token.Raw, g.stash[:g.stashOffset])
	}
	token.Row = g.stashStartRow
	token.Col = g.stashStartCol
	g.clearStash(next)
}

func (g *lexer_) isWhitespace(b byte) bool {
	return b == '\n' || b == '\t' || b == ' '
}

func (g *lexer_) isVarChar(b byte) bool {
	return g.isAlpha(b) || g.isNumber(b) || b == '_' || b == '.'
}

func (g *lexer_) isAlpha(b byte) bool {
	return g.isLowerCaseAlpha(b) || g.isUpperCaseAlpha(b)
}

func (g *lexer_) isLowerCaseAlpha(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func (g *lexer_) isUpperCaseAlpha(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

func (g *lexer_) isNumber(b byte) bool {
	return b >= '0' && b <= '9'
}
