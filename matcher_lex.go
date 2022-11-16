package djson

import (
	"errors"
	"fmt"
	"io"
)

type MatchStatus int

const (
	Matching = MatchStatus(iota)
	Matched
	NotMatch
	MatchedUntilThisTry
)

type TokenMatcher interface {
	Match(b byte, stashed []byte) MatchStatus
	Token() *Token
}

type charsMatcher struct {
	chars []byte
	token Token
}

func CharsMatcher(chars []byte, tokenType TokenType) TokenMatcher {
	return &charsMatcher{chars: chars, token: Token{Type: tokenType}}
}

func (m *charsMatcher) Match(b byte, stash []byte) MatchStatus {
	sl := len(stash)
	if sl == len(m.chars) {
		return MatchedUntilThisTry
	}
	if m.chars[sl] == b {
		return Matching
	}
	return NotMatch
}

func (m *charsMatcher) Token() *Token {
	return &m.token
}

type identifierMatcher struct {
	token Token
}

func IdentifierMatcher() TokenMatcher {
	return &identifierMatcher{token: Token{Type: TokenIdentifier}}
}

func (m *identifierMatcher) Match(b byte, stashed []byte) MatchStatus {
	sl := len(stashed)
	if sl == 0 && (isAlpha(b) || b == '_') || sl > 0 && isVarChar(b) {
		return Matching
	} else if sl > 0 && !isVarChar(b) {
		m.token.Raw = make([]byte, sl)
		copy(m.token.Raw, stashed)
		return MatchedUntilThisTry
	}
	return NotMatch
}

func (m *identifierMatcher) Token() *Token {
	ret := m.token
	return &ret
}

type whitespaceMatcher struct {
	token Token
}

func WhitespaceMatcher() TokenMatcher {
	return &whitespaceMatcher{Token{Type: TokenWhitespace}}
}

func (m *whitespaceMatcher) Match(b byte, stashed []byte) MatchStatus {
	if isWhitespace(b) {
		return Matching
	}
	sl := len(stashed)
	if sl > 0 {
		return MatchedUntilThisTry
	}
	return NotMatch
}

func (m *whitespaceMatcher) Token() *Token {
	return &m.token
}

type numberMatcher struct {
	token Token
}

func NumberMatcher() TokenMatcher {
	return &numberMatcher{Token{Type: TokenNumber}}
}

func (m *numberMatcher) Match(b byte, stashed []byte) MatchStatus {
	sl := len(stashed)
	if isNumber(b) || sl > 0 && b == '.' {
		return Matching
	}
	if sl > 0 && !isAlpha(b) && b != '_' {
		m.token.Raw = make([]byte, sl)
		copy(m.token.Raw, stashed)
		return MatchedUntilThisTry
	}
	return NotMatch
}

func (m *numberMatcher) Token() *Token {
	return &m.token
}

type stringMatcher struct {
	token   Token
	slashed bool
}

func StringMatcher() TokenMatcher {
	return &stringMatcher{token: Token{Type: TokenString}}
}

func (m *stringMatcher) Match(b byte, stashed []byte) MatchStatus {
	sl := len(stashed)
	if sl == 0 && b != '"' || b == 0 {
		return NotMatch
	}
	if b == '\\' {
		m.slashed = true
		return Matching
	}
	if sl > 0 && !m.slashed && b == '"' {
		m.token.Raw = make([]byte, len(stashed)-1)
		copy(m.token.Raw, stashed[1:])
		return Matched
	}
	m.slashed = false
	return Matching
}

func (m *stringMatcher) Token() *Token {
	return &m.token
}

type eofMatcher struct {
	token Token
}

func EOFMatcher() TokenMatcher {
	return &eofMatcher{token: Token{Type: TokenEOF}}
}

func (m *eofMatcher) Match(b byte, stashed []byte) MatchStatus {
	if b != 0 {
		return NotMatch
	}
	return Matched
}

func (m *eofMatcher) Token() *Token {
	return &m.token
}

type commentMatcher struct {
	token Token
}

func CommentMatcher() TokenMatcher {
	return &commentMatcher{Token{Type: TokenComment}}
}

func (m *commentMatcher) Match(b byte, stashed []byte) MatchStatus {
	sl := len(stashed)
	if sl == 0 && b != '#' {
		return NotMatch
	}
	if sl > 0 && b == '\n' {
		m.token.Raw = make([]byte, sl-1)
		copy(m.token.Raw, stashed[1:])
		return MatchedUntilThisTry
	}
	return Matching
}

func (m *commentMatcher) Token() *Token {
	return &m.token
}

type tokenMatcherStatus struct {
	matcher  TokenMatcher
	excloded bool
}

type matcherLexer struct {
	matchers               []*tokenMatcherStatus
	buf                    Buffer
	bs                     []byte
	stash                  []byte
	row, col               int
	tokenAtCol, tokenAtRow int
}

func NewMatcherLexer(source io.Reader, bufSize uint) Lexer {
	matchers := append(
		CharsMatchers,
		IdentifierMatcher(),
		WhitespaceMatcher(),
		CommentMatcher(),
		StringMatcher(),
		NumberMatcher(),
		EOFMatcher())
	return &matcherLexer{
		matchers: func() []*tokenMatcherStatus {
			ret := make([]*tokenMatcherStatus, len(matchers))
			for i, m := range matchers {
				ret[i] = &tokenMatcherStatus{matcher: m}
			}
			return ret
		}(),
		buf: NewBuffer(source, int(bufSize)),
		bs:  make([]byte, 1),
	}
}

type candidate struct {
	token        *Token
	dropLastChar bool
}

type candidates []candidate

func (cs candidates) slct() candidate {
	var selected candidate
	for i, c := range cs {
		if i == 0 {
			selected = c
			continue
		}
		if selected.token.Type == TokenIdentifier && c.token.Type != TokenIdentifier {
			selected = c
		}
	}
	return selected
}

func (cs *candidates) del(token *Token) {
	for i, c := range *cs {
		if c.token.Type != token.Type {
			continue
		}
		*cs = append((*cs)[:i], (*cs)[i+1:]...)
		return
	}
}

func (g *matcherLexer) NextToken(token *Token) error {
nextToken:
	g.tokenAtCol = g.col
	g.tokenAtRow = g.row
	var cs candidates
	g.stash = []byte{}
	g.initMatchers()
	for {
		if g.noAvailableMatchers() {
			break
		}
		if _, err := g.buf.Take(g.bs); err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			g.bs[0] = 0
		}
		newc := []candidate{}
		for i := 0; i < len(g.matchers); i++ {
			m := g.matchers[i].matcher
			if g.matchers[i].excloded {
				continue
			}
			switch m.Match(g.bs[0], g.stash) {
			case NotMatch:
				cs.del(m.Token())
				g.excludeMatcher(i)
			case MatchedUntilThisTry:
				newc = append(newc, candidate{token: m.Token(), dropLastChar: true})
				g.excludeMatcher(i)
			case Matched:
				newc = append(newc, candidate{token: m.Token()})
				g.excludeMatcher(i)
			case Matching:
				newc = append(newc, candidate{token: m.Token()})
			}
		}
		g.stash = append(g.stash, g.bs...)
		if len(newc) > 0 {
			g.forwardChar(g.bs[0])
			cs = newc
			continue
		}
		g.buf.TakeBack()
		break
	}
	if len(cs) == 0 {
		return fmt.Errorf("upexpected char: [%s] at %d, %d", g.bs, g.row, g.col)
	}
	cand := cs.slct()
	if cand.dropLastChar {
		g.backwardChar(g.bs[0])
		g.buf.TakeBack()
	}
	*token = *cand.token
	if token.Skip() {
		goto nextToken
	}
	token.Col = g.tokenAtCol
	token.Row = g.tokenAtRow
	return nil
}

func (g *matcherLexer) forwardChar(b byte) {
	if b == '\n' {
		g.row++
		g.col = 0
	} else {
		g.col++
	}
}

func (g *matcherLexer) noAvailableMatchers() bool {
	for _, m := range g.matchers {
		if m.excloded == false {
			return false
		}
	}
	return true
}

func (g *matcherLexer) backwardChar(b byte) {
	if b == '\n' {
		g.row--
		g.col = 0
	} else {
		g.col--
	}
}

func (g *matcherLexer) initMatchers() {
	for _, m := range g.matchers {
		m.excloded = false
	}
}

func (g *matcherLexer) excludeMatcher(i int) {
	g.matchers[i].excloded = true
}

func isWhitespace(b byte) bool {
	return b == '\n' || b == '\t' || b == ' '
}

func isVarChar(b byte) bool {
	return isAlpha(b) || isNumber(b) || b == '_'
}

func isAlpha(b byte) bool {
	return isLowerCaseAlpha(b) || isUpperCaseAlpha(b)
}

func isLowerCaseAlpha(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func isUpperCaseAlpha(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

func isNumber(b byte) bool {
	return b >= '0' && b <= '9'
}
