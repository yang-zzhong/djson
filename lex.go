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
	stateVar
	stateBlockSeperator
	stateString
	stateBool
	stateNumber
	stateVoid
	stateEqStarted
	stateEq
	stateLt
	stateGt
	stateGte
	stateLte
	stateAssign
	stateEOF
	stateTrue
	stateFalse
	stateNull
	stateKeyword
	stateAnd
	stateOr
	stateAndStarted
	stateOrStarted
)

const (
	stashSize = 256
)

var (
	keywords = [][]byte{
		{'g', 'e', 't'},
		{'s', 'e', 't'},
		{'d', 'e', 'l'},
		{'c', 'o', 'n'},
		{'G', 'E', 'T'},
		{'S', 'E', 'T'},
		{'D', 'E', 'L'},
		{'C', 'O', 'N'},
	}
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

type tokenNexter struct {
	lexer       lexer
	token       Token
	tokenUnused bool
	ends        [][]byte
	ended       bool
}

func newTokenNexter(l lexer, ends [][]byte, startToken ...Token) *tokenNexter {
	n := &tokenNexter{
		lexer: l,
		ends:  ends,
	}
	if len(startToken) > 0 {
		n.token = startToken[0]
		n.tokenUnused = true
	}
	return n
}

func (t *tokenNexter) next() (end bool, err error) {
	if t.tokenUnused || t.ended {
		end = t.ended
		return
	}
	if err = t.lexer.NextToken(&t.token); err != nil {
		return
	}
	if t.token.Type == TokenEOF {
		t.useToken(func(_ Token) {
			end = true
			t.ended = end
		})
		return
	}
	t.tokenUnused = true
	for _, ed := range t.ends {
		if bytes.Equal(ed, t.token.Raw) {
			t.useToken(func(_ Token) {
				end = true
				t.ended = end
			})
			return
		}
	}
	return
}

func (t *tokenNexter) useToken(use func(token Token)) {
	t.tokenUnused = false
	use(t.token)
}

func (t *tokenNexter) endAt() []byte {
	return t.token.Raw
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
		case stateVar:
			catched, err = g.matchVariable(b, token)
		case stateBlockSeperator:
			catched, err = g.matchBlockSeperator(b, token)
		case stateString:
			catched, err = g.matchString(b, token)
		case stateVoid:
			catched, err = g.matchVoid(b, token)
		case stateNumber:
			catched, err = g.matchNumber(b, token)
		case stateEqStarted:
			catched, err = g.matchLogicOperatorStarted(b, token, stateEq)
		case stateLt:
			catched, err = g.matchLogicOperatorStarted(b, token, stateLte)
		case stateGt:
			catched, err = g.matchLogicOperatorStarted(b, token, stateGte)
		case stateTrue, stateFalse, stateKeyword, stateNull:
			catched, err = g.matchKeyword(b, token)
		case stateAndStarted:
			catched, err = g.matchAnd(b, token)
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

func (g *lexer_) matchAnd(b byte, token *Token) (bool, error) {
	if b == '&' {
		g.state = stateAnd
		g.addToStash(b)
		return g.shiftState(0, stateVoid, token), nil
	}
	return g.matchVoid(b, token)
}

func (g *lexer_) matchKeyword(b byte, token *Token) (bool, error) {
	for i := 0; i < len(g.keyword); i++ {
		if len(g.keyword[i]) == 0 {
			if matched, catched := g.matchEndable(b, token); matched {
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
	g.state = stateVar
	if matched, catched := g.matchEndable(b, token); matched {
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

func (g *lexer_) matchLogicOperatorStarted(b byte, token *Token, next parseState) (bool, error) {
	switch {
	case b == '=':
		g.state = next
		g.addToStash(b)
		return g.shiftState(0, stateVoid, token), nil
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
	for _, keyword := range keywords {
		if keyword[0] == b {
			g.keyword = append(g.keyword, keyword[1:])
		}
	}
	if len(g.keyword) > 0 {
		return g.shiftState(b, stateKeyword, token), nil
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
	if matched, catched := g.matchEndable(b, token); matched {
		return catched, nil
	}
	if b == '"' {
		return g.shiftState(b, stateString, token), nil
	}
	switch {
	case g.isNumber(b):
		return g.shiftState(b, stateNumber, token), nil
	case g.isAlpha(b):
		return g.shiftState(b, stateVar, token), nil
	}
	return false, &Error{Row: g.row, Col: g.col, Info: UnexpectedChar, CurrentBytes: []byte{b}}
}

func (g *lexer_) matchEndable(b byte, token *Token) (matched bool, catched bool) {
	switch b {
	case '{', '}', '[', ']', ',', '(', ')', ':':
		return true, g.shiftState(b, stateBlockSeperator, token)
	case '=':
		return true, g.shiftState(b, stateEqStarted, token)
	case '<':
		return true, g.shiftState(b, stateLt, token)
	case '>':
		return true, g.shiftState(b, stateGt, token)
	case '&':
		return true, g.shiftState(b, stateAndStarted, token)
	case '|':
		return true, g.shiftState(b, stateOrStarted, token)
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
			case stateVar:
				return TokenVariable
			case stateBlockSeperator:
				switch g.stash[0] {
				case '[', '(', '{':
					return TokenBlockStart
				case ']', ')', '}':
					return TokenBlockEnd
				case ',', ':':
					return TokenBlockSeperator
				}
			case stateKeyword:
				return TokenKeyword
			case stateEq, stateGt, stateGte, stateLt, stateLte:
				return TokenComparation
			case stateAssign:
				return TokenAssignation
			case stateNumber:
				return TokenNumber
			case stateString:
				return TokenString
			case stateTrue, stateFalse:
				return TokenBoolean
			case stateAnd, stateOr:
				return TokenLogicOperator
			case stateWhitespace:
				return TokenWhitespace
			case stateNull:
				return TokenNull
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
	token.Raw = make([]byte, g.stashOffset)
	switch token.Type {
	case TokenEOF, TokenAssignation:
	default:
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
