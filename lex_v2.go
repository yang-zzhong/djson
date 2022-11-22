package djson

import (
	"errors"
	"fmt"
	"io"
	"sync"
)

var (
	matcherPool = sync.Pool{
		New: func() interface{} {
			return &matchers{
				matchers: make([]TokenMatcher, 4),
			}
		},
	}
)

type matchers struct {
	matchers []TokenMatcher
	total    int
}

func (ms *matchers) append(t TokenMatcher) {
	ms.matchers[ms.total] = t
	ms.total++
}

func (ms *matchers) reset() {
	ms.total = 0
}

func (tm matchers) match(b byte, s *stash, newc *candidates, matchers *matchers) (matching bool) {
	for i := 0; i < tm.total; i++ {
		m := tm.matchers[i]
		switch m.Match(b, s) {
		case Matched:
			newc.append(candidate{token: m.Token(), dropLastChar: true})
		case Match:
			newc.append(candidate{token: m.Token()})
		case Matching:
			matching = true
			matchers.append(m)
		}
	}
	return
}

type lexerv2 struct {
	matchers               matchers
	buf                    Buffer
	bs                     []byte
	stash                  *stash
	row, col               int
	tokenAtCol, tokenAtRow int
}

// NewLexer new a Lexer
func NewLexerV2(source io.Reader, bufSize uint) *lexerv2 {
	return &lexerv2{
		buf:   NewBuffer(source, int(bufSize)),
		bs:    make([]byte, 1),
		row:   1,
		col:   1,
		stash: newStash(stashSize),
		matchers: matchers{matchers: []TokenMatcher{
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
			CharsMatcher([]byte{'%'}, TokenMod),
			CharsMatcher([]byte{'n', 'u', 'l', 'l'}, TokenNull),
			CharsMatcher([]byte{'t', 'r', 'u', 'e'}, TokenTrue),
			CharsMatcher([]byte{'f', 'a', 'l', 's', 'e'}, TokenFalse),
			CharsMatcher([]byte{'e', 'x', 'i', 't'}, TokenExit),
			CharsMatcher([]byte{'=', '>'}, TokenReduction),
			CharsMatcher([]byte{'.', '.', '.'}, TokenRange),
			IdentifierMatcher(),
			WhitespaceMatcher(),
			CommentMatcher(),
			StringMatcher(),
			NumberMatcher(),
			EOFMatcher(),
		}, total: 37},
	}
}

func (g *lexerv2) ReplaceSource(source io.Reader, bufSize int) {
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
func (g *lexerv2) NextToken(token *Token) error {
nextToken:
	g.tokenAtCol = g.col
	g.tokenAtRow = g.row
	cs := newCandidates()
	g.stash.reset()
	ms := g.matchers
	for {
		if ms.total == 0 {
			break
		}
		if _, err := g.buf.Take(g.bs); err != nil {
			if !errors.Is(err, io.EOF) {
				return err
			}
			g.bs[0] = 0
		}
		newc := candidatesPool.Get().(*candidates)
		newc.reset()
		m := matcherPool.Get().(*matchers)
		m.reset()
		matching := ms.match(g.bs[0], g.stash, newc, m)
		ms = *m
		matcherPool.Put(m)
		g.stash.append(g.bs[0])
		if matching || newc.len() > 0 {
			g.forwardChar(g.bs[0])
			cs.copyFrom(newc)
			candidatesPool.Put(newc)
			continue
		}
		g.buf.PutLast()
		break
	}
	if cs.len() == 0 {
		return fmt.Errorf("upexpected char: [%s] at %d, %d", g.bs, g.row, g.col)
	}
	cand := cs.slct()
	if cand.dropLastChar {
		g.backwardChar(g.bs[0])
		g.buf.PutLast()
	}
	*token = *cand.token
	if token.Skip() {
		goto nextToken
	}
	token.Col = g.tokenAtCol
	token.Row = g.tokenAtRow
	return nil
}

func (g *lexerv2) forwardChar(b byte) {
	if b == '\n' {
		g.row++
		g.col = 1
	} else {
		g.col++
	}
}

func (g *lexerv2) backwardChar(b byte) {
	if b == '\n' {
		g.row--
		g.col = 0
	} else {
		g.col--
	}
}
