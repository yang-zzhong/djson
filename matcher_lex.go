package djson

import (
	"errors"
	"fmt"
	"io"
)

// NOTE this implementation is 5x unefficient than implementation of lex.go because of the using of interface
// use this implementation carefully

type MatchStatus int

const (
	Matching = MatchStatus(iota)
	Match
	NotMatch
	Matched
)

type Stash interface {
	Len() int
	CopyTo(bs []byte, start int)
}

type TokenMatcher interface {
	Match(b byte, stash Stash) MatchStatus
	Token() *Token
}

type charsMatcher struct {
	chars []byte
	token Token
}

func CharsMatcher(chars []byte, tokenType TokenType) *charsMatcher {
	return &charsMatcher{chars: chars, token: Token{Type: tokenType}}
}

func (m *charsMatcher) Match(b byte, stash Stash) MatchStatus {
	sl := stash.Len()
	if sl == len(m.chars) {
		return Matched
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

func IdentifierMatcher() *identifierMatcher {
	return &identifierMatcher{token: Token{Type: TokenIdentifier}}
}

func (m *identifierMatcher) Match(b byte, stash Stash) MatchStatus {
	sl := stash.Len()
	if sl == 0 && (isAlpha(b) || b == '_') || sl > 0 && isVarChar(b) {
		return Matching
	} else if sl > 0 && !isVarChar(b) {
		m.token.Raw = make([]byte, sl)
		stash.CopyTo(m.token.Raw, 0)
		return Matched
	}
	return NotMatch
}

func (m *identifierMatcher) Token() *Token {
	return &m.token
}

type whitespaceMatcher struct {
	token Token
}

func WhitespaceMatcher() *whitespaceMatcher {
	return &whitespaceMatcher{Token{Type: TokenWhitespace}}
}

func (m *whitespaceMatcher) Match(b byte, stash Stash) MatchStatus {
	if isWhitespace(b) {
		return Matching
	}
	if stash.Len() > 0 {
		return Matched
	}
	return NotMatch
}

func (m *whitespaceMatcher) Token() *Token {
	return &m.token
}

type numberMatcher struct {
	token Token
}

func NumberMatcher() *numberMatcher {
	return &numberMatcher{Token{Type: TokenNumber}}
}

func (m *numberMatcher) Match(b byte, stash Stash) MatchStatus {
	sl := stash.Len()
	if isNumber(b) || sl > 0 && b == '.' {
		return Matching
	}
	if sl > 0 && !isAlpha(b) && b != '_' {
		m.token.Raw = make([]byte, sl)
		stash.CopyTo(m.token.Raw, 0)
		return Matched
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

func StringMatcher() *stringMatcher {
	return &stringMatcher{token: Token{Type: TokenString}}
}

func (m *stringMatcher) Match(b byte, stash Stash) MatchStatus {
	sl := stash.Len()
	if sl == 0 && b != '"' || b == 0 {
		return NotMatch
	}
	if b == '\\' {
		m.slashed = true
		return Matching
	}
	if sl > 0 && !m.slashed && b == '"' {
		m.token.Raw = make([]byte, sl-1)
		stash.CopyTo(m.token.Raw, 1)
		return Match
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

func EOFMatcher() *eofMatcher {
	return &eofMatcher{token: Token{Type: TokenEOF}}
}

func (m *eofMatcher) Match(b byte, stash Stash) MatchStatus {
	if b != 0 {
		return NotMatch
	}
	return Match
}

func (m *eofMatcher) Token() *Token {
	return &m.token
}

type commentMatcher struct {
	token Token
}

func CommentMatcher() *commentMatcher {
	return &commentMatcher{Token{Type: TokenComment}}
}

func (m *commentMatcher) Match(b byte, stash Stash) MatchStatus {
	sl := stash.Len()
	if sl == 0 && b != '#' {
		return NotMatch
	}
	if sl > 0 && b == '\n' {
		m.token.Raw = make([]byte, sl-1)
		stash.CopyTo(m.token.Raw, 1)
		return Matched
	}
	return Matching
}

func (m *commentMatcher) Token() *Token {
	return &m.token
}

type stash struct {
	offset int
	buf    []byte
}

func NewStash(size int) *stash {
	return &stash{
		buf: make([]byte, size),
	}
}

func (s *stash) append(b byte) {
	if len(s.buf) <= s.offset {
		stash := make([]byte, len(s.buf)*2)
		copy(stash, s.buf)
		s.buf = stash
	}
	s.buf[s.offset] = b
	s.offset++
}

func (s *stash) reset() {
	s.offset = 0
}

func (s *stash) Len() int {
	return s.offset
}

func (s *stash) CopyTo(bs []byte, start int) {
	copy(bs, s.buf[start:])
}

type matcherLexer struct {
	charsMatchers          []*charsMatcher
	identifierMatcher      *identifierMatcher
	commentMatcher         *commentMatcher
	stringMatcher          *stringMatcher
	numberMatcher          *numberMatcher
	eofMatcher             *eofMatcher
	whitespaceMatcher      *whitespaceMatcher
	tokenStatus            map[TokenType]bool
	buf                    Buffer
	bs                     []byte
	stash                  *stash
	row, col               int
	tokenAtCol, tokenAtRow int
	includeMatchers        int
}

func NewMatcherLexer(source io.Reader, bufSize uint) *matcherLexer {
	lexer := &matcherLexer{
		buf:   NewBuffer(source, int(bufSize)),
		bs:    make([]byte, 1),
		stash: NewStash(stashSize),
		charsMatchers: []*charsMatcher{
			CharsMatcher([]byte{'{'}, TokenBraceOpen),
			CharsMatcher([]byte{'}'}, TokenBraceClose),
			CharsMatcher([]byte{'['}, TokenBracketsOpen),
			CharsMatcher([]byte{']'}, TokenBracketsClose),
			CharsMatcher([]byte{'('}, TokenParenthesesOpen),
			CharsMatcher([]byte{')'}, TokenParenthesesClose),
			CharsMatcher([]byte{'='}, TokenAssignation),
			CharsMatcher([]byte{'=', '='}, TokenEqual),
			CharsMatcher([]byte{'!', '='}, TokenNotEqual),
			CharsMatcher([]byte{'>'}, TokenGreateThan),
			CharsMatcher([]byte{'<'}, TokenLessThan),
			CharsMatcher([]byte{'>', '='}, TokenGreateThanEqual),
			CharsMatcher([]byte{'<', '='}, TokenLessThanEqual),
			CharsMatcher([]byte{'|', '|'}, TokenOr),
			CharsMatcher([]byte{'&', '&'}, TokenAnd),
			CharsMatcher([]byte{';'}, TokenSemicolon),
			CharsMatcher([]byte{'+'}, TokenAddition),
			CharsMatcher([]byte{'-'}, TokenMinus),
			CharsMatcher([]byte{'*'}, TokenMultiplication),
			CharsMatcher([]byte{'/'}, TokenDevision),
			CharsMatcher([]byte{':'}, TokenColon),
			CharsMatcher([]byte{','}, TokenComma),
			CharsMatcher([]byte{'.'}, TokenDot),
			CharsMatcher([]byte{'!'}, TokenExclamation),
			CharsMatcher([]byte{'n', 'u', 'l', 'l'}, TokenNull),
			CharsMatcher([]byte{'t', 'r', 'u', 'e'}, TokenTrue),
			CharsMatcher([]byte{'f', 'a', 'l', 's', 'e'}, TokenFalse),
			CharsMatcher([]byte{'=', '>'}, TokenReduction),
			CharsMatcher([]byte{'.', '.', '.'}, TokenRange),
		},
		identifierMatcher: IdentifierMatcher(),
		whitespaceMatcher: WhitespaceMatcher(),
		commentMatcher:    CommentMatcher(),
		stringMatcher:     StringMatcher(),
		numberMatcher:     NumberMatcher(),
		eofMatcher:        EOFMatcher(),
	}
	lexer.tokenStatus = make(map[TokenType]bool, len(lexer.charsMatchers)+6)
	return lexer
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

func (g *matcherLexer) ReplaceSource(source io.Reader, bufSize int) {
	g.buf = NewBuffer(source, bufSize)
}

func (g *matcherLexer) NextToken(token *Token) error {
nextToken:
	g.tokenAtCol = g.col
	g.tokenAtRow = g.row
	var cs candidates
	g.stash.reset()
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
		var newc []candidate
		g.match(&cs, &newc)
		g.stash.append(g.bs[0])
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

func (g *matcherLexer) match(cs *candidates, newc *[]candidate) {
	mth := func(match func(byte, Stash) MatchStatus, getToken func() *Token) {
		s := match(g.bs[0], g.stash)
		token := getToken()
		switch s {
		case NotMatch:
			cs.del(token)
			g.excludeMatcher(token.Type)
		case Matched:
			*newc = append(*newc, candidate{token: token, dropLastChar: true})
			g.excludeMatcher(token.Type)
		case Match:
			*newc = append(*newc, candidate{token: token})
			g.excludeMatcher(token.Type)
		case Matching:
			*newc = append(*newc, candidate{token: token})
		}
	}
	for _, m := range g.charsMatchers {
		if g.tokenStatus[m.Token().Type] {
			continue
		}
		mth(m.Match, m.Token)
	}
	if !g.tokenStatus[TokenIdentifier] {
		mth(g.identifierMatcher.Match, g.identifierMatcher.Token)
	}
	if !g.tokenStatus[TokenNumber] {
		mth(g.numberMatcher.Match, g.numberMatcher.Token)
	}
	if !g.tokenStatus[TokenComment] {
		mth(g.commentMatcher.Match, g.commentMatcher.Token)
	}
	if !g.tokenStatus[TokenString] {
		mth(g.stringMatcher.Match, g.stringMatcher.Token)
	}
	if !g.tokenStatus[TokenEOF] {
		mth(g.eofMatcher.Match, g.eofMatcher.Token)
	}
	if !g.tokenStatus[TokenWhitespace] {
		mth(g.whitespaceMatcher.Match, g.whitespaceMatcher.Token)
	}
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
	return g.includeMatchers == 0
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
	for _, m := range g.charsMatchers {
		g.tokenStatus[m.Token().Type] = false
	}
	g.tokenStatus[TokenIdentifier] = false
	g.tokenStatus[TokenNumber] = false
	g.tokenStatus[TokenString] = false
	g.tokenStatus[TokenEOF] = false
	g.tokenStatus[TokenWhitespace] = false
	g.tokenStatus[TokenComment] = false
	g.includeMatchers = len(g.charsMatchers) + 6
}

func (g *matcherLexer) excludeMatcher(tt TokenType) {
	g.tokenStatus[tt] = true
	g.includeMatchers--
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
