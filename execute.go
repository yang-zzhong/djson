package djson

import "io"

type executeInstance struct {
	r          io.Reader
	scope      *scope
	bufferSize int
}

type Execute struct {
	BufferSize int
}

func (executer *Execute) Execute(r io.Reader) error {
	ins := executeInstance{r: r, bufferSize: executer.BufferSize, scope: newScope(scopeGlobal)}
	return ins.execute()
}

func (ins *executeInstance) execute() error {
	tokenGetter := NewTokenGetter(ins.r, uint(ins.bufferSize))
	token := &Token{}
	var err error
	for {
		if err = tokenGetter.NextToken(token); err != nil {
			return err
		}
		if token.Type == TokenEOF {
			break
		}
		switch ins.scope.typ {
		case scopeGlobal:
			err = ins.executeGlobal(token)
		case scopeObject:
			err = ins.executeObject(token)
		case scopeKey:
			err = ins.executeKey(token)
		}
	}
	return nil
}

func (ins *executeInstance) executeGlobal(token *Token) error {
	switch token.Type {
	case TokenBlockStart:
		ins.enterNewScope(token)
	case TokenVariable:
		ins.executeVariable(token)
	}
	return ErrFromToken(UnexpectedChar, token)
}

func (ins *executeInstance) executeVariable(token *Token) {

}

func (ins *executeInstance) executeKey(token *Token) error {
	switch token.Type {
	case TokenWhitespace:
	case TokenBlockSeperator:
		if token.Raw[0] == ':' {
			ins.pushScope(&scope{typ: scopeValue})
		}
		if token.Raw[0] == '(' {
			ins.pushScope(&scope{typ: scopeExpr})
		}
	case TokenVariable:
		ins.pushScope(&scope{typ: scopeExpr})
		ins.stash(token)
	}
	return nil
}

func (ins *executeInstance) enterNewScope(token *Token) {
	switch token.Raw[0] {
	case '{':
		ins.pushScope(&scope{typ: scopeObject})
	}
}

func (ins *executeInstance) stash(token *Token) {
	ins.scope.stash = append(ins.scope.stash, token)
}

func (ins *executeInstance) executeObject(token *Token) error {
	switch token.Type {
	case TokenWhitespace:
	case TokenString:
		ins.pushScope(&scope{typ: scopeKey})
		ins.scope.setPair(pair{key: token.Raw})
	}
	return nil
}

func (ins *executeInstance) pushScope(s *scope) {
	ins.scope.addChild(s)
	ins.scope = s
}
