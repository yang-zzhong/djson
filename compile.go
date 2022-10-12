package djson

import "io"

type compileState int

const (
	compileStart compileState = iota
	compileAssign
	compileAssignValue
)

type compiler struct {
	scope   *scope
	r       io.Reader
	bufSize uint
	state   compileState
	stash   []*Token
}

func (c *compiler) Compile(p *program) error {
	c.scope = &scope{typ: scopeGlobal}
	p.scope = c.scope
	tokenGetter := NewTokenGetter(c.r, c.bufSize)
	var token Token
	c.state = compileStart
	for {
		if err := tokenGetter.NextToken(&token); err != nil {
			return err
		}
		if token.Type == TokenEOF {
			return nil
		}
		var err error
		switch c.state {
		case compileStart:
			err = c.handleStart(&token, p)
		case compileAssign:
			err = c.handleAssign(&token, p)
		}
	}
	return nil
}

func (c *compiler) handleStart(token *Token, p *program) error {
	switch token.Type {
	case TokenWhitespace:
	case TokenVariable:
		c.state = compileAssign
		c.stash = append(c.stash, token)
	}
	return nil
}

func (c *compiler) handleAssign(token *Token, p *program) error {
	switch {
	case token.Type == TokenWhitespace:
	case token.Raw[0] == '=':
		c.state = compileAssignValue
	}
}
