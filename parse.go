package djson

import (
	"bytes"
	"errors"
	"io"
)

type state int

const (
	stateStart = state(iota)
	stateWhitespace
	stateVar
	stateBlockSeperator
	stateString
	stateBool
	stateNumber
	stateVoid
	stateEqStarted
	stateLtStarted
	stateGtStarted
	stateEq
	stateLt
	stateGt
	stateGte
	stateLte
	stateAssign
	stateEOF
	stateTrue
	stateFalse
	stateAnd
	stateOr
	stateAndStarted
	stateOrStarted
)

const (
	stashSize = 256
)

type Token struct {
	Type     TokenType
	Raw      []byte
	Row, Col int
}

type TokenGetter interface {
	NextToken(token *Token) error
}

type tokenGetter struct {
	source       io.Reader
	row, col     int
	offset       int
	buf          []byte
	max          int
	state        state
	stash        []byte
	slashOpenned bool
	stashOffset  int
	keyword      []byte
}

func NewTokenGetter(source io.Reader, bufSize uint) *tokenGetter {
	return &tokenGetter{
		source: source,
		row:    0, col: 0, offset: 0,
		buf: make([]byte, bufSize), max: 0,
		state: stateWhitespace,
		stash: make([]byte, stashSize),
	}
}

func (g *tokenGetter) NextToken(token *Token) (err error) {
	var b byte
	var catched bool
	for {
		if err = g.nextChar(&b); err != nil {
			if errors.Is(err, io.EOF) {
				return g.handleEOF(token)
			}
			return
		}
		if b == 0 {
			continue
		}
		switch g.state {
		case stateStart:
			catched, err = g.handleStart(b, token)
		case stateWhitespace:
			catched, err = g.handleWhitespace(b, token)
		case stateVar:
			catched, err = g.handleVariable(b, token)
		case stateBlockSeperator:
			catched, err = g.handleBlockSeperator(b, token)
		case stateString:
			catched, err = g.handleString(b, token)
		case stateVoid:
			catched, err = g.handleVoid(b, token)
		case stateNumber:
			catched, err = g.handleNumber(b, token)
		case stateEqStarted:
			catched, err = g.handleLogicOperatorStarted(b, token, stateEq)
		case stateLtStarted:
			catched, err = g.handleLogicOperatorStarted(b, token, stateLte)
		case stateGtStarted:
			catched, err = g.handleLogicOperatorStarted(b, token, stateGte)
		case stateTrue, stateFalse:
			catched, err = g.handleKeyword(b, token)
		case stateAndStarted:
			catched, err = g.handleAnd(b, token)
		}
		if catched || err != nil {
			return
		}
	}
}

func (g *tokenGetter) handleAnd(b byte, token *Token) (bool, error) {
	if b == '&' {
		g.state = stateAnd
		g.addToStash(b)
		g.shiftState(0, stateVoid, token)
		return true, nil
	}
	return g.handleVoid(b, token)
}

func (g *tokenGetter) handleKeyword(b byte, token *Token) (bool, error) {
	if len(g.keyword) > 0 && g.keyword[0] == b {
		g.keyword = g.keyword[1:]
		g.addToStash(b)
		return false, nil
	}
	if g.matchEndable(b, token) {
		return true, nil
	}
	if g.isVarChar(b) {
		g.addToStash(b)
		g.state = stateVar
		return false, nil
	}
	return false, &ParseError{
		Row:          g.row,
		Col:          g.col,
		CurrentBytes: []byte{b},
		Info:         UnexpectedChar,
	}
}

func (g *tokenGetter) handleLogicOperatorStarted(b byte, token *Token, next state) (bool, error) {
	switch {
	case b == '=':
		g.state = next
		g.addToStash(b)
		g.shiftState(b, stateVoid, token)
		return true, nil
	}
	return g.handleVoid(b, token)
}

func (g *tokenGetter) handleEOF(token *Token) error {
	err := &ParseError{
		Info: UnexpectedEOF,
		Row:  g.row,
		Col:  g.col,
	}
	if g.state != stateEOF {
		switch {
		case stateTrue == g.state && !bytes.Equal(g.stash[:4], []byte{'t', 'r', 'u', 'e'}):
			return err
		case stateFalse == g.state && !bytes.Equal(g.stash[:5], []byte{'f', 'a', 'l', 's', 'e'}):
			return err
		case stateString == g.state:
			return err
		}
		g.shiftState(' ', stateEOF, token)
	}
	return nil
}

func (g *tokenGetter) handleNumber(b byte, token *Token) (bool, error) {
	switch {
	case g.isNumber(b) || b == '.':
		g.addToStash(b)
		return false, nil
	}
	return g.matchNormal(b, token)
}

func (g *tokenGetter) handleString(b byte, token *Token) (bool, error) {
	switch b {
	case '"':
		if g.slashOpenned {
			g.addToStash(b)
			g.slashOpenned = false
			return false, nil
		}
		g.addToStash(b)
		g.shiftState(0, stateVoid, token)
		return true, nil
	case '\\':
		g.slashOpenned = true
		g.addToStash(b)
	default:
		g.addToStash(b)
	}
	return false, nil
}

func (g *tokenGetter) handleEqStarted(b byte, token *Token) (bool, error) {
	switch {
	case b == '=':
		g.addToStash(b)
		g.state = stateEq
		g.shiftState(b, stateVoid, token)
		return true, nil
	}
	return g.handleVoid(b, token)
}

func (g *tokenGetter) handleStart(b byte, token *Token) (bool, error) {
	if b == 0 {
		return false, nil
	}
	_, err := g.handleVoid(b, token)
	return false, err
}

func (g *tokenGetter) handleVoid(b byte, token *Token) (bool, error) {
	switch b {
	case 't', 'T':
		g.keyword = []byte{'r', 'u', 'e'}
		g.shiftState(b, stateTrue, token)
		return true, nil
	case 'f', 'F':
		g.keyword = []byte{'a', 'l', 's', 'e'}
		g.shiftState(b, stateFalse, token)
		return true, nil
	}
	return g.matchNormal(b, token)
}

func (g *tokenGetter) handleWhitespace(b byte, token *Token) (bool, error) {
	if g.isWhitespace(b) {
		g.addToStash(b)
		return false, nil
	}
	switch b {
	case 't':
		g.shiftState(b, stateTrue, token)
	case 'f':
		g.shiftState(b, stateFalse, token)
	}
	return g.matchNormal(b, token)
}

func (g *tokenGetter) handleBlockSeperator(b byte, token *Token) (bool, error) {
	return g.handleVoid(b, token)
}

func (g *tokenGetter) handleVariable(b byte, token *Token) (bool, error) {
	switch {
	case g.isVarChar(b):
		g.addToStash(b)
		return false, nil
	case g.isWhitespace(b):
		g.shiftState(b, stateWhitespace, token)
		return true, nil
	}
	return g.matchNormal(b, token)
}

func (g *tokenGetter) matchNormal(b byte, token *Token) (bool, error) {
	if g.matchEndable(b, token) {
		return true, nil
	}
	if b == '"' {
		g.shiftState(b, stateString, token)
		return true, nil
	}
	switch {
	case g.isNumber(b):
		g.shiftState(b, stateNumber, token)
		return true, nil
	case g.isAlpha(b):
		g.shiftState(b, stateVar, token)
		return true, nil
	}
	return false, &ParseError{Row: g.row, Col: g.col, Info: UnexpectedChar, CurrentBytes: []byte{b}}
}

func (g *tokenGetter) matchEndable(b byte, token *Token) bool {
	switch b {
	case '{', '}', '[', ']', ',', '(', ')':
		g.shiftState(b, stateBlockSeperator, token)
		return true
	case '=':
		g.shiftState(b, stateEqStarted, token)
		return true
	case '<':
		g.shiftState(b, stateLtStarted, token)
		return true
	case '>':
		g.shiftState(b, stateGtStarted, token)
		return true
	case '&':
		g.shiftState(b, stateAndStarted, token)
	case '|':
		g.shiftState(b, stateOrStarted, token)
	}
	if g.isWhitespace(b) {
		g.shiftState(b, stateWhitespace, token)
		return true
	}
	return false
}

func (g *tokenGetter) nextChar(b *byte) error {
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

func (g *tokenGetter) shiftState(b byte, state state, token *Token) {
	if len(g.stash) > 0 && g.state != stateStart {
		token.Row = g.row
		token.Col = g.col
		token.Type = func() TokenType {
			switch g.state {
			case stateVar:
				return TokenVariable
			case stateEq:
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
			}
			return TokenType(100000)
		}()
		g.takeTokenRaw(token)
	}
	if b != 0 {
		g.addToStash(b)
	}
	g.state = state
}

func (g *tokenGetter) addToStash(b byte) {
	if len(g.stash) <= g.stashOffset {
		stash := make([]byte, len(g.stash)*2)
		copy(stash, g.stash)
		g.stash = stash
	}
	g.stash[g.stashOffset] = b
	g.stashOffset++
}

func (g *tokenGetter) takeTokenRaw(token *Token) {
	token.Raw = make([]byte, g.stashOffset)
	copy(token.Raw, g.stash[:g.stashOffset])
	if len(g.stash) != stashSize {
		g.stash = make([]byte, stashSize)
	}
	g.stashOffset = 0
}

func (g *tokenGetter) isWhitespace(b byte) bool {
	return b == '\n' || b == '\t' || b == ' '
}

func (g *tokenGetter) isVarChar(b byte) bool {
	return g.isAlpha(b) || g.isNumber(b) || b == '_'
}

func (g *tokenGetter) isAlpha(b byte) bool {
	return g.isLowerCaseAlpha(b) || g.isUpperCaseAlpha(b)
}

func (g *tokenGetter) isLowerCaseAlpha(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func (g *tokenGetter) isUpperCaseAlpha(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

func (g *tokenGetter) isNumber(b byte) bool {
	return b >= '0' && b <= '9'
}
