package djson

import (
	"errors"
	"fmt"
	"io"
)

// MatchStatus match status for Matcher.Match
type MatchStatus int

const (
	Matching = MatchStatus(iota) // token matched the current byte but maybe unmatch for any of incoming byte
	Match                        // token totally matched
	NotMatch                     // token not match
	Matched                      // token matched before the current byte

	stashSize = 256 // the history of the matching bytes
)

// Lexer is a tool for analysis byte stream and get the matched token
type Lexer interface {
	NextToken(token *Token) error
}

// Stash hold the history of the current matching process
type Stash interface {
	Len() int                    // history bytes length
	CopyTo(bs []byte, start int) // copy the stash to ather []byte
}

// TokenMatcher a matching tool for specific token
type TokenMatcher interface {
	Match(b byte, stash Stash) MatchStatus // match byte and stash for a token
	Token() *Token                         // get the token of the matcher
}

// SourceReplacer replace source for the Lexer
type SourceReplacer interface {
	ReplaceSource(io.Reader, int)
}

// Buffer a read buffer for lexer
type Buffer interface {
	Take([]byte) (int, error) // get []byte from the buffer
	TakeBack()                // put the last take back for some reason, such as token matched with Matched status returned
}

type buffer struct {
	r            io.Reader
	bs           []byte
	lastTakeSize int
	offset       int
	total        int
}

// NewBuffer new a buffer from a reader and a buffer size
func NewBuffer(r io.Reader, size int) Buffer {
	return &buffer{
		r:  r,
		bs: make([]byte, size),
	}
}

// read the underlying bytes
func (b *buffer) read() (err error) {
	b.total, err = b.r.Read(b.bs)
	if err != nil {
		return
	}
	b.offset = 0
	return
}

// Take get []byte from the buffer
func (b *buffer) Take(bs []byte) (taked int, err error) {
	if b.total == 0 || b.offset == b.total {
		if err = b.read(); err != nil {
			return
		}
	}
	b.lastTakeSize = copy(bs, b.bs[b.offset:])
	b.offset += b.lastTakeSize
	taked = b.lastTakeSize
	return
}

// TakeBack put the last take back for some reason, such as token matched with Matched status returned
func (b *buffer) TakeBack() {
	b.offset -= b.lastTakeSize
	b.lastTakeSize = 0
}

type charsMatcher struct {
	chars []byte
	token Token
}

// CharsMatcher match the specific chars
func CharsMatcher(chars []byte, tokenType TokenType) TokenMatcher {
	return &charsMatcher{chars: chars, token: Token{Type: tokenType}}
}

// Match match current byte and stash
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

// Token get the token associates the matcher
func (m *charsMatcher) Token() *Token {
	return &m.token
}

type identifierMatcher struct {
	token Token
}

// IdentifierMatcher for matching identifier
func IdentifierMatcher() TokenMatcher {
	return &identifierMatcher{token: Token{Type: TokenIdentifier}}
}

// Match implement tokenMatcher.Match
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

// Token implement tokenMatcher.Token
func (m *identifierMatcher) Token() *Token {
	return &m.token
}

type whitespaceMatcher struct {
	token Token
}

// IdentifierMatcher for matching whitespace
func WhitespaceMatcher() TokenMatcher {
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

// IdentifierMatcher for matching number
func NumberMatcher() TokenMatcher {
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

// IdentifierMatcher for matching string
func StringMatcher() TokenMatcher {
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

// IdentifierMatcher for matching eof
func EOFMatcher() TokenMatcher {
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

// IdentifierMatcher for matching comment
func CommentMatcher() TokenMatcher {
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

type tokenMatcherStatus struct {
	matcher  TokenMatcher
	excloded bool
}

type stash struct {
	offset int
	buf    []byte
}

func newStash(size int) *stash {
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

// Len implement stash.Len
func (s *stash) Len() int {
	return s.offset
}

// Len implement stash.CopyTo
func (s *stash) CopyTo(bs []byte, start int) {
	copy(bs, s.buf[start:])
}

type lexer struct {
	matcherStatuses        []*tokenMatcherStatus
	buf                    Buffer
	bs                     []byte
	stash                  *stash
	row, col               int
	tokenAtCol, tokenAtRow int
	includeMatchers        int
}

// NewLexer new a Lexer
func NewLexer(source io.Reader, bufSize uint) *lexer {
	return &lexer{
		buf:   NewBuffer(source, int(bufSize)),
		bs:    make([]byte, 1),
		row:   1,
		col:   1,
		stash: newStash(stashSize),
		matcherStatuses: []*tokenMatcherStatus{
			{matcher: CharsMatcher([]byte{'{'}, TokenBraceOpen)},
			{matcher: CharsMatcher([]byte{'}'}, TokenBraceClose)},
			{matcher: CharsMatcher([]byte{'['}, TokenBracketsOpen)},
			{matcher: CharsMatcher([]byte{']'}, TokenBracketsClose)},
			{matcher: CharsMatcher([]byte{'('}, TokenParenthesesOpen)},
			{matcher: CharsMatcher([]byte{')'}, TokenParenthesesClose)},
			{matcher: CharsMatcher([]byte{'='}, TokenAssignation)},
			{matcher: CharsMatcher([]byte{'=', '='}, TokenEqual)},
			{matcher: CharsMatcher([]byte{'!', '='}, TokenNotEqual)},
			{matcher: CharsMatcher([]byte{'>'}, TokenGreateThan)},
			{matcher: CharsMatcher([]byte{'<'}, TokenLessThan)},
			{matcher: CharsMatcher([]byte{'>', '='}, TokenGreateThanEqual)},
			{matcher: CharsMatcher([]byte{'<', '='}, TokenLessThanEqual)},
			{matcher: CharsMatcher([]byte{'|', '|'}, TokenOr)},
			{matcher: CharsMatcher([]byte{'&', '&'}, TokenAnd)},
			{matcher: CharsMatcher([]byte{';'}, TokenSemicolon)},
			{matcher: CharsMatcher([]byte{'+'}, TokenAddition)},
			{matcher: CharsMatcher([]byte{'-'}, TokenMinus)},
			{matcher: CharsMatcher([]byte{'*'}, TokenMultiplication)},
			{matcher: CharsMatcher([]byte{'/'}, TokenDevision)},
			{matcher: CharsMatcher([]byte{':'}, TokenColon)},
			{matcher: CharsMatcher([]byte{','}, TokenComma)},
			{matcher: CharsMatcher([]byte{'.'}, TokenDot)},
			{matcher: CharsMatcher([]byte{'!'}, TokenExclamation)},
			{matcher: CharsMatcher([]byte{'%'}, TokenMod)},
			{matcher: CharsMatcher([]byte{'n', 'u', 'l', 'l'}, TokenNull)},
			{matcher: CharsMatcher([]byte{'t', 'r', 'u', 'e'}, TokenTrue)},
			{matcher: CharsMatcher([]byte{'f', 'a', 'l', 's', 'e'}, TokenFalse)},
			{matcher: CharsMatcher([]byte{'e', 'x', 'i', 't'}, TokenExit)},
			{matcher: CharsMatcher([]byte{'=', '>'}, TokenReduction)},
			{matcher: CharsMatcher([]byte{'.', '.', '.'}, TokenRange)},
			{matcher: IdentifierMatcher()},
			{matcher: WhitespaceMatcher()},
			{matcher: CommentMatcher()},
			{matcher: StringMatcher()},
			{matcher: NumberMatcher()},
			{matcher: EOFMatcher()},
		},
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

func (g *lexer) ReplaceMatchers(matchers ...TokenMatcher) {
	g.matcherStatuses = func() []*tokenMatcherStatus {
		ret := make([]*tokenMatcherStatus, len(matchers))
		for i, m := range matchers {
			ret[i] = &tokenMatcherStatus{matcher: m}
		}
		return ret
	}()
}

func (g *lexer) RegisterMatcher(matchers ...TokenMatcher) {
	g.matcherStatuses = append(g.matcherStatuses, func() []*tokenMatcherStatus {
		ret := make([]*tokenMatcherStatus, len(matchers))
		for i, m := range matchers {
			ret[i] = &tokenMatcherStatus{matcher: m}
		}
		return ret
	}()...)
}

func (g *lexer) ReplaceSource(source io.Reader, bufSize int) {
	g.buf = NewBuffer(source, bufSize)
}

// NextToken let's take a look on the implementation. the structure like below
//
// ```
//      matched token
//     <---------+                    +----------+   +--> Match
//               |                  +----------+ |   |--> Token
//      +------------------+       +---------+ |_+----+
//   +->| candidate tokens |       | matcher |-+
//   |  +------------------+       +---------+
//   |  +------------------+             |
//   +--| filtered matcher | <-----------+
//      +------------------+
//               |
//               |
//             bytes
// ```
func (g *lexer) NextToken(token *Token) error {
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

func (g *lexer) match(cs *candidates, newc *[]candidate) {
	for i := 0; i < len(g.matcherStatuses); i++ {
		if g.matcherStatuses[i].excloded {
			continue
		}
		m := g.matcherStatuses[i].matcher
		switch m.Match(g.bs[0], g.stash) {
		case NotMatch:
			cs.del(m.Token())
			g.excludeMatcher(i)
		case Matched:
			*newc = append(*newc, candidate{token: m.Token(), dropLastChar: true})
			g.excludeMatcher(i)
		case Match:
			*newc = append(*newc, candidate{token: m.Token()})
			g.excludeMatcher(i)
		case Matching:
			*newc = append(*newc, candidate{token: m.Token()})
		}
	}
}

func (g *lexer) forwardChar(b byte) {
	if b == '\n' {
		g.row++
		g.col = 1
	} else {
		g.col++
	}
}

func (g *lexer) noAvailableMatchers() bool {
	return g.includeMatchers == 0
}

func (g *lexer) backwardChar(b byte) {
	if b == '\n' {
		g.row--
		g.col = 0
	} else {
		g.col--
	}
}

func (g *lexer) initMatchers() {
	for _, m := range g.matcherStatuses {
		m.excloded = false
	}
	g.includeMatchers = len(g.matcherStatuses)
}

func (g *lexer) excludeMatcher(i int) {
	g.matcherStatuses[i].excloded = true
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

type lexmock struct {
	offset int
	tokens []*Token
}

func (g *lexmock) NextToken(token *Token) error {
	if len(g.tokens) == g.offset {
		token.Type = TokenEOF
		return nil
	}
	token.Raw = g.tokens[g.offset].Raw
	token.Type = g.tokens[g.offset].Type
	g.offset++
	return nil
}

func newLexMock(tokens []*Token) *lexmock {
	return &lexmock{
		tokens: tokens,
	}
}
